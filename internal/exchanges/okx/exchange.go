package okx

import (
	"log"

	"github.com/anuramat/arbitrage/internal/exchanges"
)

type Okx struct {
	exchanges.BaseExchange
}

// XXX might need pinger

func (r *Okx) Subscribe(currencyPairs []string, logger *log.Logger) {
	// each pair updates every 10ms if there is a change
	r.priceUpdater(currencyPairs, logger)
}

func New() *Okx {
	return &Okx{exchanges.BaseExchange{Name: "okx"}}
}
