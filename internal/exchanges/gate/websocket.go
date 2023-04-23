package gate

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

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
	websocket.DefaultDialer.TLSClientConfig = &tls.Config{RootCAs: nil, InsecureSkipVerify: true} // XXX might be insecure?
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		panic(err)
	}
	return c
}

func (r *Gate) priceUpdater(ctx context.Context, wg *sync.WaitGroup, currencyPairs []string) {
	defer wg.Done()

	c := makeConnection()
	defer c.Close()

	// subscribe to prices
	t := time.Now().Unix()
	orderBookMsg := message{t, "spot.book_ticker", "subscribe", currencyPairs}
	err := orderBookMsg.send(c)
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
		select {
		case <-ctx.Done():
			c.Close()
			return
		default:
			_, msg, err := c.ReadMessage()
			if err != nil {
				fmt.Println("Error reading ws message:", err)
				return
			}
			var ticker tickerUpdate
			err = json.Unmarshal(msg, &ticker)
			if err != nil {
				fmt.Println("Error unmarshalling message: ", err)
				return
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
