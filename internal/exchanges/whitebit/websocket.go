package whitebit

import (
	"encoding/json"
	"errors"
	"log"
	"net/url"
	"time"

	"github.com/anuramat/arbitrage/internal/models"
	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
)

const (
	maxDepth = 100
	interval = "0"
)

func makeConnection() (*websocket.Conn, error) {
	u := url.URL{Scheme: "wss", Host: "api.whitebit.com", Path: "/ws"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (msg *request) send(c *websocket.Conn) error {
	msgByte, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return c.WriteMessage(websocket.TextMessage, msgByte)
}

func (r *Whitebit) priceUpdater(currencyPairs []string, logger *log.Logger, updateChannel chan<- models.UpdateNotification) {
	for _, currencyPair := range currencyPairs {
		go r.singlePriceUpdater(currencyPair, logger, updateChannel)
	}
}

func (r *Whitebit) singlePriceUpdater(currencyPair string, logger *log.Logger, updateChannel chan<- models.UpdateNotification) {
	errPrinter := func(description string, err error) {
		logger.Printf("%s, %s pair on exchange %s: %v\n", description, currencyPair, r.Name, err)
	}
	conn, err := makeConnection()
	if err != nil {
		errPrinter("Error making ws connection", err)
		return
	}
	defer conn.Close()

	// subscribe to prices
	requestID := r.requestId.Add(1)
	req := request{
		ID:     requestID,
		Method: "depth_subscribe",
		Params: []any{currencyPair, 1, interval, true},
	}
	err = req.send(conn)
	if err != nil {
		errPrinter("Error subscribing", err)
	}

	// receive subscription confirmation
	if !subscriptionCheck(conn, errPrinter) {
		return
	}

	// start pinging
	go r.pinger(conn, errPrinter)

	// receive price updates
	market := r.Markets[currencyPair]
	for {
		// read ws message
		_, msg, err := conn.ReadMessage()
		ts := tsApprox()
		if err != nil {
			errPrinter("Error reading update", err)
			return
		}
		// parse json
		update := depthUpdate{}
		err = json.Unmarshal(msg, &update)
		if err != nil {
			errPrinter("Error unmarshalling update", err)
			return
		}
		// check for ping
		if update.Result == "pong" {
			continue
		}
		// enjoy some hot steamy action with unstructured data
		orderBook := depthUpdateData{}
		json.Unmarshal(update.Params[1], &orderBook)
		lowestAsk := extractPrice(orderBook.Asks)
		highestBid := extractPrice(orderBook.Bids)

		// update best price
		market.BestPrice.Lock()
		if !lowestAsk.IsZero() {
			market.BestPrice.Ask = lowestAsk
		}
		if !highestBid.IsZero() {
			market.BestPrice.Bid = highestBid
		}
		market.BestPrice.Timestamp = ts
		market.BestPrice.Unlock()

		updateChannel <- models.UpdateNotification{Pair: currencyPair, ExchangeIndex: market.Index, ExchangeName: r.Name}
	}

}

func (r *Whitebit) singleBookUpdater(currencyPair string, logger *log.Logger, updateChannel chan<- models.UpdateNotification) {
	errPrinter := func(description string, err error) {
		logger.Printf("%s, %s pair on exchange %s: %v\n", description, currencyPair, r.Name, err)
	}
	conn, err := makeConnection()
	if err != nil {
		errPrinter("Error making ws connection", err)
		return
	}
	defer conn.Close()

	// subscribe to prices
	requestID := r.requestId.Add(1)
	req := request{
		ID:     requestID,
		Method: "depth_subscribe",
		Params: []any{currencyPair, maxDepth, interval, true},
	}
	err = req.send(conn)
	if err != nil {
		errPrinter("Error subscribing", err)
	}

	// receive subscription confirmation
	if !subscriptionCheck(conn, errPrinter) {
		return
	}

	// start pinging
	go r.pinger(conn, errPrinter)

	// receive price updates
	market := r.Markets[currencyPair]
	for {
		// read ws message
		_, msg, err := conn.ReadMessage()
		ts := tsApprox()
		if err != nil {
			errPrinter("Error reading update", err)
			return
		}
		// parse json
		update := depthUpdate{}
		err = json.Unmarshal(msg, &update)
		if err != nil {
			errPrinter("Error unmarshalling update", err)
			return
		}
		// check for ping
		if update.Result == "pong" {
			continue
		}
		// enjoy some hot steamy action with unstructured data
		bookUpdate := depthUpdateData{}
		json.Unmarshal(update.Params[1], &bookUpdate)

		// copy book
		asks := []models.OrderBookEntry{}
		bids := []models.OrderBookEntry{}
		market.OrderBook.RLock()
		copy(market.OrderBook.Asks, asks)
		copy(market.OrderBook.Bids, bids)
		market.OrderBook.RUnlock()

		// merge updates into copies
		asks, err = mergeBooks(bookUpdate.Asks, asks, func(a, b decimal.Decimal) bool { return a.LessThan(b) }) // asks are sorted ascending
		if err != nil {
			errPrinter("Error merging asks", err)
			return
		}
		bids, err = mergeBooks(bookUpdate.Bids, bids, func(a, b decimal.Decimal) bool { return a.GreaterThan(b) }) // bids are sorted descending
		if err != nil {
			errPrinter("Error merging bids", err)
			return
		}
		// lowestAsk := decimal.Zero
		// highestBid := decimal.Zero

		// if len(asks) > 0 {
		// 	lowestAsk = asks[0].Price
		// }
		// if len(bids) > 0 {
		// 	highestBid = bids[0].Price
		// }

		// write
		market.OrderBook.Lock()
		market.OrderBook.Asks = asks
		market.OrderBook.Bids = bids
		market.OrderBook.Timestamp = ts
		market.OrderBook.Unlock()

		// // update best price
		// market.BestPrice.Lock()
		// if !lowestAsk.IsZero() {
		// 	market.BestPrice.Ask = lowestAsk
		// }
		// if !highestBid.IsZero() {
		// 	market.BestPrice.Bid = highestBid
		// }
		// market.BestPrice.Timestamp = ts
		// market.BestPrice.Unlock()

		// updateChannel <- models.UpdateNotification{Pair: currencyPair, ExchangeIndex: market.Index, ExchangeName: r.Name}
	}

}

func extractPrice(prices [][2]string) decimal.Decimal {
	for _, pair := range prices {
		price, _ := decimal.NewFromString(pair[0])
		amount, _ := decimal.NewFromString(pair[1])
		if !amount.IsZero() {
			return price
		}
	}
	return decimal.Zero
}

func tsApprox() int64 {
	// TODO this is a pessimistic approximation of the timestamp given we read ws message on arrival
	return time.Now().UnixMilli() - 1500
}

func (r *Whitebit) pinger(conn *websocket.Conn, errPrinter func(string, error)) {
	for {
		req := request{r.requestId.Add(1), "ping", []any{}}
		err := req.send(conn)
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
	if response.Error != nil {
		errPrinter("Error subscribing", errors.New(response.Error.Message))
		return
	}
	return true
}

func mergeBooks(updates [][2]string, book []models.OrderBookEntry, comparator func(a, b decimal.Decimal) bool) ([]models.OrderBookEntry, error) {
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
