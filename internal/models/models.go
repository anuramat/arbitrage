// This package implements data structures that store market data for different exchanges.
package models

import (
	"log"
	"sync"

	"github.com/shopspring/decimal"
)

type OrderBookEntry struct {
	Price  decimal.Decimal
	Amount decimal.Decimal
}

// XXX: optimize? fixed size array, saves time on memory allocation
type OrderBook struct {
	Bids      []OrderBookEntry
	Asks      []OrderBookEntry
	Timestamp int64
	sync.RWMutex
}

type BestPrice struct {
	Bid       decimal.Decimal
	Ask       decimal.Decimal
	Timestamp int64
	sync.RWMutex
}

// XXX: optimize? single mutex for some exchanges, saves time on lock/unlock
// We assume that the order book and best price will be updated independently.
// It would be more efficient to replace mutex with a pointer to mutex,
// which would be the same for both OrderBook and BestPrice for exchanges
// with frequent order book updates.

// Market is a data structure that stores market data for a specific exchange
// and currency pair.
type Market struct {
	Exchange
	OrderBook
	BestPrice
	Index int
}

type UpdateNotification struct {
	Pair          string
	ExchangeIndex int // index of the market in the allMarkets array for given currency pair
	ExchangeName  string
}

// ExchangeMarkets stores market data for a specific exchange.
// Key is a currency pair.
type ExchangeMarkets map[string]*Market

// AllMarkets stores market data for all exchanges.
// Key is a currency pair.
type AllMarkets map[string][]*Market

type Exchange interface {
	Subscribe(currencyPairs []string, logger *log.Logger, updateChannel chan<- UpdateNotification)
	MakeMarkets(currencyPairs []string, allMarkets *AllMarkets)
	GetName() string
	GetMarkets() *ExchangeMarkets
}
