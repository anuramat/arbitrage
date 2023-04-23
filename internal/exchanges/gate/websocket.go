package gate

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
)

type message struct {
	Time    int64    `json:"time"`
	Channel string   `json:"channel"`
	Event   string   `json:"event"`
	Payload []string `json:"payload"`
}

func (msg *message) send(c *websocket.Conn) error {
	msgByte, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return c.WriteMessage(websocket.TextMessage, msgByte)
}

func makeConnection() *websocket.Conn {
	u := url.URL{Scheme: "wss", Host: "api.gateio.ws", Path: "/ws/v4/"}
	websocket.DefaultDialer.TLSClientConfig = &tls.Config{RootCAs: nil, InsecureSkipVerify: true} // TODO might be insecure?
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		panic(err)
	}
	return c
}

func (r gate) priceUpdater(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	c := makeConnection()

	// subscribe to prices
	t := time.Now().Unix()
	orderBookMsg := message{t, "spot.book_ticker", "subscribe", r.currencyPairs}
	err := orderBookMsg.send(c)
	if err != nil {
		panic(err)
	}

	// receive subscription confirmation
	_, msg, err := c.ReadMessage()
	if err != nil {
		// TODO better error handling
		fmt.Println(err)
	}
	response := &subscriptionResponse{}
	err = json.Unmarshal(msg, response)
	if err != nil {
		// TODO better error handling
		fmt.Println(err)
	}
	if response.Error != nil {
		// TODO better error handling
		fmt.Println(response.Error)
		return
	}

	for {
		spew.Dump(r.markets) // TODO remove
		select {
		case <-ctx.Done():
			c.Close()
			return
		default:
			_, msg, err := c.ReadMessage()
			if err != nil {
				// TODO better error handling
				fmt.Println(err)
			}
			// TODO check if error in message
			var ticker tickerUpdate
			err = json.Unmarshal(msg, &ticker)
			if err != nil {
				fmt.Println("Error unmarshalling message: ", err) // TODO error handling
			}
			// check if ticker is for a currency pair we are interested in
			if _, ok := r.markets[ticker.Result.CurrencyPair]; !ok {
				continue
			}
			r.markets[ticker.Result.CurrencyPair].BestPrice.RWMutex.Lock()
			r.markets[ticker.Result.CurrencyPair].BestPrice.Ask, _ = decimal.NewFromString(ticker.Result.AskPrice)
			r.markets[ticker.Result.CurrencyPair].BestPrice.Bid, _ = decimal.NewFromString(ticker.Result.BidPrice)
			r.markets[ticker.Result.CurrencyPair].BestPrice.Timestamp = ticker.Result.TimeMs
			r.markets[ticker.Result.CurrencyPair].BestPrice.RWMutex.Unlock()
		}
	}
}
