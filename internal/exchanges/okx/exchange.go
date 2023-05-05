package okx

import (
	"github.com/anuramat/arbitrage/internal/exchanges"
)

type Okx struct {
	exchanges.BaseExchange
}

// XXX might need pinger

func (r *Okx) Subscribe(currencyPairs []string) {
	// each pair updates every 10ms if there is a change
	r.priceUpdater(currencyPairs)
}

func New() *Okx {
	return &Okx{exchanges.BaseExchange{Name: "okx"}}
}
