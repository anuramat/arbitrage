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
	fmt.Println("SENDING: ", string(msgByte)) // TODO remove
	return c.WriteMessage(websocket.TextMessage, msgByte)
}

func makeConnection() *websocket.Conn {
	// TODO check this code
	u := url.URL{Scheme: "wss", Host: "api.gateio.ws", Path: "/ws/v4/"}
	websocket.DefaultDialer.TLSClientConfig = &tls.Config{RootCAs: nil, InsecureSkipVerify: true}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		panic(err)
	}
	return c
}

type Ticker struct {
	Time    int64  `json:"time"`
	TimeMs  int64  `json:"time_ms"`
	Channel string `json:"channel"`
	Event   string `json:"event"`
	Result  struct {
		TimeMs       int64  `json:"t"`
		UpdateID     int64  `json:"u"`
		CurrencyPair string `json:"s"`
		BidPrice     string `json:"b"`
		AskPrice     string `json:"a"`
		BidAmount    string `json:"B"`
		AskAmount    string `json:"A"`
	} `json:"result"`
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
	fmt.Printf("subscription result: %s\n", msg) // TODO change
	for {
		spew.Dump(r.markets)
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
			var ticker Ticker
			err = json.Unmarshal(msg, &ticker)
			if err != nil {
				fmt.Println("Error unmarshalling message: ", err) // TODO change
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
