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
	for _, currencyPair := range currencyPairs {
		go r.singlePriceUpdater(currencyPair)
	}
}

func (r *Gate) singlePriceUpdater(currencyPair string) {
	conn := makeConnection()
	defer conn.Close()

	// subscribe to prices
	t := time.Now().Unix()
	request := subscriptionRequest{t, "spot.book_ticker", "subscribe", []string{currencyPair}}
	err := request.send(conn)
	if err != nil {
		fmt.Println("Error sending ws message:", err)
	}

	// receive subscription confirmation
	_, msg, err := conn.ReadMessage()
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
	market := r.Markets[currencyPair]
	for {
		// read ws message
		_, msg, err := conn.ReadMessage()
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
		market.BestPrice.RWMutex.Lock()
		market.BestPrice.Bid, _ = decimal.NewFromString(update.Result.BidPrice)
		market.BestPrice.Ask, _ = decimal.NewFromString(update.Result.AskPrice)
		market.BestPrice.Timestamp = update.Result.TimeMs
		market.BestPrice.RWMutex.Unlock()
	}

}
