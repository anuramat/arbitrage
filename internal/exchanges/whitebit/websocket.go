package whitebit

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
)

func makeConnection() (*websocket.Conn, error) {
	u := url.URL{Scheme: "wss", Host: "api.whitebit.com", Path: "/ws"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (msg *subscriptionRequest) send(c *websocket.Conn) error {
	msgByte, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return c.WriteMessage(websocket.TextMessage, msgByte)
}

func (r *Whitebit) priceUpdater(currencyPairs []string) {
	for _, currencyPair := range currencyPairs {
		go r.singlePriceUpdater(currencyPair)
	}
}

func (r *Whitebit) singlePriceUpdater(currencyPair string) {
	errPrinter := func(description string, err error) {
		fmt.Printf("%s, %s pair on exchange %s: %e\n", description, currencyPair, r.Name, err)
	}
	conn, err := makeConnection()
	if err != nil {
		errPrinter("Error making ws connection", err)
		return
	}
	defer conn.Close()

	// subscribe to prices
	requestID := r.requestId.Add(1)
	request := subscriptionRequest{
		ID:     requestID,
		Method: "depth_subscribe",
		Params: []any{currencyPair, 1, "0", true},
	}
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
		update := depthUpdate{}
		err = json.Unmarshal(msg, &update)
		if err != nil {
			errPrinter("Error unmarshalling update", err)
			return
		}
		// enjoy some hot steamy action with unstructured data
		orderBook := depthUpdateData{}
		json.Unmarshal(update.Params[1], &orderBook)

		// orderBook := update.Params[1].(map[string]any)
		newAsk := extractPrice(orderBook.Asks)
		newBid := extractPrice(orderBook.Bids)

		// update values
		market.BestPrice.Lock()
		if !newAsk.IsZero() {
			market.BestPrice.Ask = newAsk
		}
		if !newBid.IsZero() {
			market.BestPrice.Bid = newBid
		}
		// TODO this is a pessimistic approximation of the timestamp
		market.BestPrice.Timestamp = time.Now().UnixMilli() - 1500
		market.BestPrice.Unlock()
	}

}

func extractPrice(prices [][]string) decimal.Decimal {
	for _, pair := range prices {
		price, _ := decimal.NewFromString(pair[0])
		amount, _ := decimal.NewFromString(pair[1])
		if !amount.IsZero() {
			return price
		}
	}
	return decimal.Zero
}