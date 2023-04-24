package okx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"

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

func (msg *SubscribeRequest) send(c *websocket.Conn) error {
	msgByte, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return c.WriteMessage(websocket.TextMessage, msgByte)
}

func (r *Okx) priceUpdater(ctx context.Context, wg *sync.WaitGroup, currencyPairs []string) {
	defer wg.Done()

	c := makeConnection()
	defer c.Close()

	// subscribe to prices
	args := []SubscriptionArgs{}
	for _, currencyPair := range currencyPairs {
		currencyPair = strings.Replace(currencyPair, "_", "-", 1)
		args = append(args, SubscriptionArgs{Channel: "bbo-tbt", InstID: currencyPair})
	}
	request := SubscribeRequest{Op: "subscribe", Args: args}
	request.send(c)

	// receive subscription confirmation
	for range currencyPairs {
		_, _, _ = c.ReadMessage()
		// TODO check that all subscriptions were successful
	}

	// receive price updates
	for {
		select {
		case <-ctx.Done():
			c.Close()
			return
		default:
			// read ws message
			_, msg, err := c.ReadMessage()
			if err != nil {
				fmt.Println("Error reading ws message:", err)
				return
			}
			// parse json
			var update BookSnapshotResponse
			err = json.Unmarshal(msg, &update)
			if err != nil {
				fmt.Println("Error unmarshalling message: ", err)
				return
			}
			// update values
			currencyPair := strings.Replace(update.Arg.InstID, "-", "_", 1)
			r.Markets[currencyPair].BestPrice.RWMutex.Lock()
			r.Markets[currencyPair].BestPrice.Ask, _ = decimal.NewFromString(update.Data[0].Asks[0][0])
			r.Markets[currencyPair].BestPrice.Bid, _ = decimal.NewFromString(update.Data[0].Bids[0][0])
			r.Markets[currencyPair].BestPrice.Timestamp, _ = strconv.ParseInt(update.Data[0].Ts, 10, 64)
			r.Markets[currencyPair].BestPrice.RWMutex.Unlock()
		}

	}
}
