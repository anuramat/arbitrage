package okx

import (
	"encoding/json"
	"errors"
	"hash/crc32"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/anuramat/arbitrage/internal/models"
	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
)

func makeConnection() (*websocket.Conn, error) {
	u := url.URL{Scheme: "wss", Host: "wsaws.okx.com:8443", Path: "/ws/v5/public"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (msg *subscribeRequest) send(c *websocket.Conn) error {
	msgByte, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return c.WriteMessage(websocket.TextMessage, msgByte)
}

func (r *Okx) priceUpdater(pairs []string, logger *log.Logger, updateChannel chan<- models.UpdateNotification) {
	for _, pair := range pairs {
		go r.singlePriceUpdater(pair, logger, updateChannel)
		go r.singleBookUpdater(pair, logger)
	}
}

func (r *Okx) singlePriceUpdater(pair string, logger *log.Logger, updateChannel chan<- models.UpdateNotification) {
	errPrinter := func(description string, err error) {
		logger.Printf("%s:singlePriceUpdater, %s, %s: %v", r.Name, pair, description, err)
	}

	conn, err := makeConnection()
	if err != nil {
		errPrinter("Error making ws connection", err)
		return
	}
	defer conn.Close()

	// subscribe to prices
	pair = strings.Replace(pair, "_", "-", 1)
	request := subscribeRequest{Op: "subscribe", Args: []subscriptionArg{{Channel: "bbo-tbt", InstID: pair}}}
	err = request.send(conn)
	if err != nil {
		errPrinter("Error subscribing", err)
		return
	}

	// receive subscription confirmation
	if !subscriptionCheck(conn, errPrinter) {
		return
	}

	// start pinging
	go pinger(conn, errPrinter)

	// receive price updates
	pair = strings.Replace(pair, "-", "_", 1)
	market := r.Markets[pair]
	for {
		// read ws message
		_, msg, err := conn.ReadMessage()
		if err != nil {
			errPrinter("Error reading update", err)
			return
		}
		if string(msg) == "pong" {
			continue
		}
		// parse json
		var update bookSnapshotUpdate
		err = json.Unmarshal(msg, &update)
		if err != nil {
			errPrinter("Error unmarshalling update", err)
			return
		}
		// update values
		market.BestPrice.Lock()
		market.BestPrice.Bid, _ = decimal.NewFromString(update.Data[0].Bids[0][0])
		market.BestPrice.Ask, _ = decimal.NewFromString(update.Data[0].Asks[0][0])
		market.BestPrice.Timestamp = update.Data[0].Ts
		market.BestPrice.Unlock()
		updateChannel <- models.UpdateNotification{Pair: pair, ExchangeIndex: market.Index, ExchangeName: r.Name}
	}
}

func (r *Okx) singleBookUpdater(pair string, logger *log.Logger) {
	errPrinter := func(description string, err error) {
		logger.Printf("%s:singleBookUpdater, %s, %s: %v", r.Name, pair, description, err)
	}

	conn, err := makeConnection()
	if err != nil {
		errPrinter("Error making ws connection", err)
		return
	}
	defer conn.Close()

	// subscribe to orderbook
	pair = strings.Replace(pair, "_", "-", 1)
	request := subscribeRequest{Op: "subscribe", Args: []subscriptionArg{{Channel: "books", InstID: pair}}}
	err = request.send(conn)
	if err != nil {
		errPrinter("Error subscribing", err)
		return
	}

	// receive subscription confirmation
	if !subscriptionCheck(conn, errPrinter) {
		return
	}

	// start pinging
	go pinger(conn, errPrinter)

	// receive orderbook updates
	pair = strings.Replace(pair, "-", "_", 1)
	market := r.Markets[pair]
	for {

		// read ws message
		_, msg, err := conn.ReadMessage()
		if err != nil {
			errPrinter("Error reading update", err)
			return
		}
		if string(msg) == "pong" {
			continue
		}
		// parse json
		var update bookSnapshotUpdate

		err = json.Unmarshal(msg, &update)
		if err != nil {
			errPrinter("Error unmarshalling update", err)
			return
		}

		// copy book
		asks, bids := market.CopyAsksBids()

		// merge updates into copies
		asks_update := parseOrderStrings(update.Data[0].Asks)
		asks = models.MergeBooks(asks_update, asks, true)
		bids_update := parseOrderStrings(update.Data[0].Bids)
		bids = models.MergeBooks(bids_update, bids, false)

		clientChecksum := checksum(asks, bids)
		serverChecksum := uint32(update.Data[0].Checksum)
		if clientChecksum != serverChecksum {
			errPrinter("Checksum error", ErrOrderbookDesync)
			// TODO restart
			return
		}

		market.WriteOrderbook(asks, bids, update.Data[0].Ts)

	}

}

func pinger(conn *websocket.Conn, errPrinter func(string, error)) {
	for {
		err := conn.WriteMessage(websocket.TextMessage, []byte("ping"))
		if err != nil {
			errPrinter("Error sending ping", err)
			return
		}
		time.Sleep(15 * time.Second)
	}
}

func subscriptionCheck(conn *websocket.Conn, errPrinter func(string, error)) (ok bool) {
	_, msg, err := conn.ReadMessage()
	if err != nil {
		errPrinter("Error reading subscription response", err)
		return false
	}
	response := &subscriptionResponse{}
	err = json.Unmarshal(msg, response)
	if err != nil {
		errPrinter("Error unmarshalling subscription response", err)
		return false
	}
	if response.Event == "error" {
		errPrinter("Error subscribing", errors.New(response.Msg))
		return false
	}
	return true
}

func parseOrderStrings(orders [][4]string) []models.OrderBookEntry {
	entries := make([]models.OrderBookEntry, len(orders))
	for i, order := range orders {
		price, _ := decimal.NewFromString(order[0])
		amount, _ := decimal.NewFromString(order[1])
		entries[i] = models.OrderBookEntry{Price: price, Amount: amount}
	}
	return entries
}

func checksum(asks, bids []models.OrderBookEntry) uint32 {
	pieces := []string{}
	smallerLen := min(min(len(asks), len(bids)), 25)
	for i := 0; i < smallerLen; i++ {
		pieces = append(pieces, bids[i].Price.String(), bids[i].Amount.String(), asks[i].Price.String(), asks[i].Amount.String())
	}
	if len(asks) > len(bids) {
		for i := smallerLen; i < min(25, len(asks)); i++ {
			pieces = append(pieces, asks[i].Price.String(), asks[i].Amount.String())
		}
	} else {
		for i := smallerLen; i < min(25, len(bids)); i++ {
			pieces = append(pieces, bids[i].Price.String(), bids[i].Amount.String())
		}
	}
	clientChecksum := crc32.ChecksumIEEE([]byte(strings.Join(pieces, ":")))
	return clientChecksum
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
