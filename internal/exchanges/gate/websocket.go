package gate

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

// TODO remove hardcode?? note that those values can only take discrete values, details are in the documentation
const (
	depth    = "100"
	interval = "100ms"
)

func (request *subscriptionRequest) send(c *websocket.Conn) error {
	msg, err := json.Marshal(request)
	if err != nil {
		return err
	}
	return c.WriteMessage(websocket.TextMessage, msg)
}

// TODO remove hardcode
func makeConnection() (*websocket.Conn, error) {
	u := url.URL{Scheme: "wss", Host: "api.gateio.ws", Path: "/ws/v4/"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *Gate) updater(pairs []string, logger *log.Logger, updateChannel chan<- models.UpdateNotification) {
	for _, pair := range pairs {
		go r.singlePriceUpdater(pair, logger, updateChannel)
		go r.singleBookUpdater(pair, logger)
	}
}

func (r *Gate) singlePriceUpdater(pair string, logger *log.Logger, updateChannel chan<- models.UpdateNotification) {
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
	t := time.Now().Unix()
	request := subscriptionRequest{t, "spot.book_ticker", "subscribe", []string{pair}}
	err = request.send(conn)
	if err != nil {
		errPrinter("Error subscribing", err)
	}

	// receive subscription confirmation
	if !subscriptionCheck(conn, errPrinter) {
		return
	}

	// receive price updates
	market := r.Markets[pair]
	for {
		// read ws message
		_, msg, err := conn.ReadMessage()
		if err != nil {
			errPrinter("Error reading update", err)
			return
		}
		// parse json
		var update tickerUpdate
		err = json.Unmarshal(msg, &update)
		if err != nil {
			errPrinter("Error unmarshalling update", err)
			return
		}
		// update values
		market.BestPrice.Lock()
		market.BestPrice.Bid, _ = decimal.NewFromString(update.Result.BidPrice)
		market.BestPrice.Ask, _ = decimal.NewFromString(update.Result.AskPrice)
		market.BestPrice.Timestamp = update.Result.TimeMs
		market.BestPrice.Unlock()
		updateChannel <- models.UpdateNotification{Pair: pair, ExchangeIndex: market.Index, ExchangeName: r.Name}
	}

}

// TODO transition to updates? kinda hard tho
func (r *Gate) singleBookUpdater(pair string, logger *log.Logger) {
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
	t := time.Now().Unix()
	request := subscriptionRequest{t, "spot.order_book", "subscribe", []string{pair, depth, interval}}
	err = request.send(conn)
	if err != nil {
		errPrinter("Error subscribing", err)
	}

	// receive subscription confirmation
	if !subscriptionCheck(conn, errPrinter) {
		return
	}

	// receive price updates
	market := r.Markets[pair]
	for {
		// read ws message
		_, msg, err := conn.ReadMessage()
		if err != nil {
			errPrinter("Error reading update", err)
			return
		}
		// parse json
		var update bookUpdate
		err = json.Unmarshal(msg, &update)
		if err != nil {
			errPrinter("Error unmarshalling update", err)
			return
		}

		// parse string arrays to proper data types
		asks := fillEntryArray(update.Result.Asks)
		bids := fillEntryArray(update.Result.Bids)
		// write
		market.OrderBook.Lock()
		market.OrderBook.Asks = asks
		market.OrderBook.Bids = bids
		market.OrderBook.Timestamp = update.Result.TimeMs
		market.OrderBook.Unlock()
	}

}

func fillEntryArray(stringArray [][2]string) []models.OrderBookEntry {
	result := []models.OrderBookEntry{}
	for _, stringEntry := range stringArray {
		price, _ := decimal.NewFromString(stringEntry[0])
		amount, _ := decimal.NewFromString(stringEntry[1])
		result = append(result, models.OrderBookEntry{
			Price:  price,
			Amount: amount,
		})
	}
	return result
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
		return false
	}
	return true
}
