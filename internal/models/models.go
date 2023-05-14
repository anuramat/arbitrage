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

func (market *Market) CopyAsksBids() (asks []OrderBookEntry, bids []OrderBookEntry) {
	market.OrderBook.RLock()
	defer market.OrderBook.RUnlock()
	asks = make([]OrderBookEntry, len(market.OrderBook.Asks))
	copy(market.OrderBook.Asks, asks)
	bids = make([]OrderBookEntry, len(market.OrderBook.Bids))
	copy(market.OrderBook.Bids, bids)
	return asks, bids
}

func (market *Market) WriteOrderbook(asks []OrderBookEntry, bids []OrderBookEntry, ts int64) {
	market.OrderBook.Lock()
	defer market.OrderBook.Unlock()
	market.OrderBook.Asks = asks
	market.OrderBook.Bids = bids
	market.OrderBook.Timestamp = ts
}

func MergeBooks(updates, book []OrderBookEntry, isAsks bool) []OrderBookEntry {
	var comparator func(a, b decimal.Decimal) bool
	if isAsks {
		comparator = func(a, b decimal.Decimal) bool {
			return a.LessThan(b)
		}
	} else {
		comparator = func(a, b decimal.Decimal) bool {
			return a.GreaterThan(b)
		}
	}
	j := 0
	for _, update := range updates {
		for ; j < len(book); j++ {
			if book[j].Price.Equal(update.Price) {
				if update.Amount.IsZero() {
					book = append(book[:j], book[j+1:]...)
					j--
				} else {
					book[j].Amount = update.Amount
				}
				break
			}
			if comparator(update.Price, book[j].Price) {
				if update.Amount.IsZero() {
					break
				}
				entry := OrderBookEntry{Price: update.Price, Amount: update.Amount}
				book = append(book[:j], append([]OrderBookEntry{entry}, book[j:]...)...)
				break
			}
		}
		if j == len(book) {
			if update.Amount.IsZero() {
				break
			}
			entry := OrderBookEntry{Price: update.Price, Amount: update.Amount}
			book = append(book, entry)
		}
	}
	return book
}
