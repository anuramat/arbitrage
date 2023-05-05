package whitebit

import (
	"sync/atomic"

	"github.com/anuramat/arbitrage/internal/exchanges"
)

type Whitebit struct {
	exchanges.BaseExchange
	requestId atomic.Int64
}

func (r *Whitebit) Subscribe(currencyPairs []string) {
	go r.priceUpdater(currencyPairs)
}

func New() *Whitebit {
	return &Whitebit{exchanges.BaseExchange{Name: "whitebit"}, atomic.Int64{}}
}
