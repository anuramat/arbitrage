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

func (request *subscriptionRequest) send(c *websocket.Conn) error {
	msg, err := json.Marshal(request)
	if err != nil {
		return err
	}
	return c.WriteMessage(websocket.TextMessage, msg)
}

func makeConnection() (*websocket.Conn, error) {
	u := url.URL{Scheme: "wss", Host: "api.gateio.ws", Path: "/ws/v4/"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *Gate) priceUpdater(currencyPairs []string, logger *log.Logger, updateChannel chan<- models.UpdateNotification) {
	for _, currencyPair := range currencyPairs {
		go r.singlePriceUpdater(currencyPair, logger, updateChannel)
	}
}

func (r *Gate) singlePriceUpdater(currencyPair string, logger *log.Logger, updateChannel chan<- models.UpdateNotification) {
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
	t := time.Now().Unix()
	request := subscriptionRequest{t, "spot.book_ticker", "subscribe", []string{currencyPair}}
	err = request.send(conn)
	if err != nil {
		errPrinter("Error subscribing", err)
	}

	// receive subscription confirmation
	_, msg, err := conn.ReadMessage()
	if err != nil {
		errPrinter("Error reading subscription response", err)
		return
	}
	response := &subscriptionResponse{}
	err = json.Unmarshal(msg, response)
	if err != nil {
		errPrinter("Error unmarshalling subscription response", err)
		return
	}
	if response.Error != nil {
		errPrinter("Error subscribing", errors.New(response.Error.Message))
		return
	}

	// receive price updates
	market := r.Markets[currencyPair]
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
		updateChannel <- models.UpdateNotification{CurrencyPair: currencyPair, ExchangeIndex: market.Index}
	}

}
