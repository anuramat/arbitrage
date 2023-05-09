package whitebit

import (
	"log"
	"sync/atomic"

	"github.com/anuramat/arbitrage/internal/exchanges"
)

type Whitebit struct {
	exchanges.BaseExchange
	requestId atomic.Int64
}

func (r *Whitebit) Subscribe(currencyPairs []string, logger *log.Logger) {
	go r.priceUpdater(currencyPairs, logger)
}

func New() *Whitebit {
	return &Whitebit{exchanges.BaseExchange{Name: "whitebit"}, atomic.Int64{}}
}
