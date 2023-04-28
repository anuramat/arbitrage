package okx

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
)

func makeConnection() *websocket.Conn {
	u := url.URL{Scheme: "wss", Host: "wsaws.okx.com:8443", Path: "/ws/v5/public"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		panic(err)
	}
	return c
}

func (msg *subscribeRequest) send(c *websocket.Conn) error {
	msgByte, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return c.WriteMessage(websocket.TextMessage, msgByte)
}

func (r *Okx) priceUpdater(currencyPairs []string) {
	for _, currencyPair := range currencyPairs {
		go r.singlePriceUpdater(currencyPair)
	}
}

func (r *Okx) singlePriceUpdater(currencyPair string) {
	conn := makeConnection()
	defer conn.Close()

	// subscribe to prices
	currencyPair = strings.Replace(currencyPair, "_", "-", 1)
	request := subscribeRequest{Op: "subscribe", Args: []subscriptionArg{{Channel: "bbo-tbt", InstID: currencyPair}}}
	request.send(conn)

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
	if response.Event == "error" {
		fmt.Println("Error subscribing:", response.Msg)
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
		var update bookSnapshotUpdate
		err = json.Unmarshal(msg, &update)
		if err != nil {
			fmt.Println("Error unmarshalling message: ", err)
			return
		}
		// update values
		market.BestPrice.RWMutex.Lock()
		market.BestPrice.Bid, _ = decimal.NewFromString(update.Data[0].Bids[0][0])
		market.BestPrice.Ask, _ = decimal.NewFromString(update.Data[0].Asks[0][0])
		market.BestPrice.Timestamp = update.Data[0].Ts
		market.BestPrice.RWMutex.Unlock()
	}
}
