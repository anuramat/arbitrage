package okx

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

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

func (r *Okx) priceUpdater(currencyPairs []string) {
	for _, currencyPair := range currencyPairs {
		go r.singlePriceUpdater(currencyPair)
	}
}

func (r *Okx) singlePriceUpdater(currencyPair string) {
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
	currencyPair = strings.Replace(currencyPair, "_", "-", 1)
	request := subscribeRequest{Op: "subscribe", Args: []subscriptionArg{{Channel: "bbo-tbt", InstID: currencyPair}}}
	err = request.send(conn)
	if err != nil {
		errPrinter("Error subscribing", err)
		return
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
	if response.Event == "error" {
		errPrinter("Error subscribing", errors.New(response.Msg))
		return
	}

	// receive price updates
	currencyPair = strings.Replace(currencyPair, "-", "_", 1)
	market := r.Markets[currencyPair]
	for {
		// read ws message
		_, msg, err := conn.ReadMessage()
		if err != nil {
			errPrinter("Error reading update", err)
			return
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
	}
}
