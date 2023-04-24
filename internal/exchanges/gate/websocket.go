package gate

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

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

func makeConnection() *websocket.Conn {
	u := url.URL{Scheme: "wss", Host: "api.gateio.ws", Path: "/ws/v4/"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		panic(err)
	}
	return c
}

func (r *Gate) priceUpdater(currencyPairs []string) {
	c := makeConnection()
	defer c.Close()

	// subscribe to prices
	t := time.Now().Unix()
	request := subscriptionRequest{t, "spot.book_ticker", "subscribe", currencyPairs}
	err := request.send(c)
	if err != nil {
		fmt.Println("Error sending ws message:", err)
	}

	// receive subscription confirmation
	_, msg, err := c.ReadMessage()
	if err != nil {
		fmt.Println("Error reading ws message:", err)
		return
	}
	response := &subscriptionResponse{}
	err = json.Unmarshal(msg, response)
	if err != nil {
		fmt.Println("Error unmarshalling message: ", err)
		return
	}
	if response.Error != nil {
		fmt.Println("Error subscribing:", response.Error)
		return
	}

	// receive price updates
	for {
		// read ws message
		_, msg, err := c.ReadMessage()
		if err != nil {
			fmt.Println("Error reading ws message:", err)
			return
		}
		// parse json
		var update tickerUpdate
		err = json.Unmarshal(msg, &update)
		if err != nil {
			fmt.Println("Error unmarshalling message: ", err)
			return
		}
		// update values
		r.Markets[update.Result.CurrencyPair].BestPrice.RWMutex.Lock()
		r.Markets[update.Result.CurrencyPair].BestPrice.Ask, _ = decimal.NewFromString(update.Result.AskPrice)
		r.Markets[update.Result.CurrencyPair].BestPrice.Bid, _ = decimal.NewFromString(update.Result.BidPrice)
		r.Markets[update.Result.CurrencyPair].BestPrice.Timestamp = update.Result.TimeMs
		r.Markets[update.Result.CurrencyPair].BestPrice.RWMutex.Unlock()

	}
}

// func (r *Gate) orderBookUpdater(ctx context.Context, wg *sync.WaitGroup, currencyPairs []string) {
// 	defer wg.Done()

// 	c := makeConnection()
// 	defer c.Close()

// 	// subscribe to order books
// 	t := time.Now().Unix()
// 	payload := []string{}
// 	for _, currencyPair := range currencyPairs {
// 		payload = append(payload, currencyPair, "100ms")
// 	}
// 	request := subscriptionRequest{t, "spot.order_book_update", "subscribe", payload}
// 	err := request.send(c)
// 	if err != nil {
// 		fmt.Println("Error sending ws message:", err)
// 	}

// 	// receive subscription confirmation
// 	_, msg, err := c.ReadMessage()
// 	if err != nil {
// 		fmt.Println("Error reading ws message:", err)
// 		return
// 	}
// 	response := &subscriptionResponse{}
// 	err = json.Unmarshal(msg, response)
// 	if err != nil {
// 		fmt.Println("Error unmarshalling message: ", err)
// 		return
// 	}
// 	if response.Error != nil {
// 		fmt.Println("Error subscribing:", response.Error)
// 		return
// 	}
// 	fmt.Println(response)
// }
