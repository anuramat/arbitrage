package okx

import (
	"encoding/json"
	"errors"
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
		logger.Printf("%s, %s pair on exchange %s: %v\n", description, pair, r.Name, err)
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
		logger.Printf("%s, %s pair on exchange %s: %v\n", description, pair, r.Name, err)
	}

	conn, err := makeConnection()
	if err != nil {
		errPrinter("Error making ws connection", err)
		return
	}
	defer conn.Close()

	// subscribe to prices
	pair = strings.Replace(pair, "_", "-", 1)
	request := subscribeRequest{Op: "subscribe", Args: []subscriptionArg{{Channel: "books5", InstID: pair}}}
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
		// copy asks/bids
		asks := []models.OrderBookEntry{}
		bids := []models.OrderBookEntry{}
		market.OrderBook.RLock()
		copy(market.OrderBook.Asks, asks)
		copy(market.OrderBook.Bids, bids)
		market.OrderBook.RUnlock()

		// merge updates into copies
		asks = mergeBooks(update.Data[0].Asks, asks, func(a, b decimal.Decimal) bool { return a.LessThan(b) })    // asks are sorted ascending
		bids = mergeBooks(update.Data[0].Bids, bids, func(a, b decimal.Decimal) bool { return a.GreaterThan(b) }) // bids are sorted descending

		// write
		market.OrderBook.Lock()
		market.OrderBook.Asks = asks
		market.OrderBook.Bids = bids
		market.OrderBook.Timestamp = update.Data[0].Ts
		market.OrderBook.Unlock()
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

// XXX this is gonna be a problem, needs testing
// TODO remove panics after adding checksum and restarting on lost update
func mergeBooks(updates [][]string, book []models.OrderBookEntry, comparator func(a, b decimal.Decimal) bool) []models.OrderBookEntry {
	j := 0
	for _, update := range updates {
		newPrice, _ := decimal.NewFromString(update[0])
		newAmount, _ := decimal.NewFromString(update[1])
		for ; j < len(book); j++ {
			if book[j].Price.Equal(newPrice) {
				if newAmount.IsZero() {
					book = append(book[:j], book[j+1:]...)
					j--
				} else {
					book[j].Amount = newAmount
				}
				break
			}
			if comparator(newPrice, book[j].Price) {
				if newAmount.IsZero() {
					panic("zero amount with a new price, probably lost update")
				}
				entry := models.OrderBookEntry{Price: newPrice, Amount: newAmount}
				book = append(book[:j], append([]models.OrderBookEntry{entry}, book[j:]...)...)
				break
			}
		}
		if j == len(book) {
			if newAmount.IsZero() {
				panic("zero amount in book update, probably lost update")
			}
			entry := models.OrderBookEntry{Price: newPrice, Amount: newAmount}
			book = append(book, entry)
		}
	}
	return book
}
