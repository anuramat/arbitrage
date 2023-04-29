package whitebit

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

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
		Params: []interface{}{currencyPair, 1, "0", true},
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
	// market := r.Markets[currencyPair]
	for {
		// read ws message
		_, msg, err := conn.ReadMessage()
		if err != nil {
			errPrinter("Error reading update", err)
			return
		}
		// parse json
		fmt.Println(string(msg))
		update := depthUpdate{}
		err = json.Unmarshal(msg, &update)
		if err != nil {
			errPrinter("Error unmarshalling update", err)
			return
		}
		// enjoy some hot steamy action with unstructured data
		orderBook := update.Params[1].(map[string]interface{})

		newAsk := getNewPrice(orderBook, "asks")
		newBid := getNewPrice(orderBook, "bids")

		fmt.Println(newAsk, newBid)

		// update values
		// market.BestPrice.Lock()
		// market.BestPrice.Bid, _ = decimal.NewFromString(update.Result.BidPrice)
		// market.BestPrice.Ask, _ = decimal.NewFromString(update.Result.AskPrice)
		// market.BestPrice.Timestamp = update.Result.TimeMs
		// market.BestPrice.Unlock()
	}

}

func getNewPrice(orderBook map[string]interface{}, side string) decimal.Decimal {
	if orders_i, ok := orderBook[side]; ok {
		orders_si := orders_i.([]interface{})
		for _, order_i := range orders_si {
			order_si := order_i.([]interface{})
			price, _ := decimal.NewFromString(order_si[0].(string))
			amount, _ := decimal.NewFromString(order_si[1].(string))
			if !amount.IsZero() {
				return price
			}
		}
	}
	return decimal.Zero
}
