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
	c := makeConnection()
	defer c.Close()

	// subscribe to prices
	args := []subscriptionArgs{}
	for _, currencyPair := range currencyPairs {
		currencyPair = strings.Replace(currencyPair, "_", "-", 1)
		args = append(args, subscriptionArgs{Channel: "bbo-tbt", InstID: currencyPair})
	}
	request := subscribeRequest{Op: "subscribe", Args: args}
	request.send(c)

	// receive subscription confirmation
	for range currencyPairs {
		_, _, _ = c.ReadMessage()
		// TODO check that all subscriptions were successful
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
		var update bookSnapshotUpdate
		err = json.Unmarshal(msg, &update)
		if err != nil {
			fmt.Println("Error unmarshalling message: ", err)
			return
		}
		// update values
		currencyPair := strings.Replace(update.Arg.InstID, "-", "_", 1)
		r.Markets[currencyPair].BestPriceValue.RWMutex.Lock()
		r.Markets[currencyPair].BestPriceValue.Ask, _ = decimal.NewFromString(update.Data[0].Asks[0][0])
		r.Markets[currencyPair].BestPriceValue.Bid, _ = decimal.NewFromString(update.Data[0].Bids[0][0])
		r.Markets[currencyPair].BestPriceValue.Timestamp = update.Data[0].Ts
		r.Markets[currencyPair].BestPriceValue.RWMutex.Unlock()

	}
}
