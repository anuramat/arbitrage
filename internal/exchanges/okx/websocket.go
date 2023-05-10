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

		// copy book
		asks := []models.OrderBookEntry{}
		bids := []models.OrderBookEntry{}
		market.OrderBook.RLock()
		copy(market.OrderBook.Asks, asks)
		copy(market.OrderBook.Bids, bids)
		market.OrderBook.RUnlock()

		// // debug check sort
		// asks_string := update.Data[0].Asks
		// for i := 0; i < len(asks_string)-1; i++ {
		// 	cur, _ := strconv.ParseFloat(asks_string[i][0], 64)
		// 	next, _ := strconv.ParseFloat(asks_string[i+1][0], 64)
		// 	if cur > next {
		// 		logger.Println("asks not sorted")
		// 	}
		// }

		// bids_string := update.Data[0].Bids
		// for i := 0; i < len(bids_string)-1; i++ {
		// 	cur, _ := strconv.ParseFloat(bids_string[i][0], 64)
		// 	next, _ := strconv.ParseFloat(bids_string[i+1][0], 64)
		// 	if cur < next {
		// 		logger.Println("bids not sorted")
		// 	}
		// }

		// merge updates into copies
		asks, err = mergeBooks(update.Data[0].Asks, asks, func(a, b decimal.Decimal) bool { return a.LessThan(b) }) // asks are sorted ascending
		if err != nil {
			errPrinter("Error merging asks", err)
			return
		}
		bids, err = mergeBooks(update.Data[0].Bids, bids, func(a, b decimal.Decimal) bool { return a.GreaterThan(b) }) // bids are sorted descending
		if err != nil {
			errPrinter("Error merging bids", err)
			return
		}

		err = checksum(asks, bids, uint32(update.Data[0].Checksum))
		if err != nil {
			errPrinter("Checksum error", err)
			return
		}

		// logger.Println(len(asks), len(bids)) // XXX good check, should be equal to max depth on low depths

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

func mergeBooks(updates [][4]string, book []models.OrderBookEntry, comparator func(a, b decimal.Decimal) bool) ([]models.OrderBookEntry, error) {
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
					return nil, ErrOrderbookDesync
				}
				entry := models.OrderBookEntry{Price: newPrice, Amount: newAmount}
				book = append(book[:j], append([]models.OrderBookEntry{entry}, book[j:]...)...)
				break
			}
		}
		if j == len(book) {
			if newAmount.IsZero() {
				return nil, ErrOrderbookDesync
			}
			entry := models.OrderBookEntry{Price: newPrice, Amount: newAmount}
			book = append(book, entry)
		}
	}
	return book, nil
}

func checksum(asks, bids []models.OrderBookEntry, serverChecksum uint32) error {
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
	if clientChecksum != serverChecksum {
		return ErrOrderbookDesync
	}
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
