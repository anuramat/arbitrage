package okx

import (
	"github.com/anuramat/arbitrage/internal/exchanges"
)

type Okx struct {
	exchanges.BaseExchange
}

// XXX might need pinger

func (r *Okx) Subscribe(currencyPairs []string) {
	r.priceUpdater(currencyPairs)
}

func New() *Okx {
	return &Okx{exchanges.BaseExchange{Name: "OKX"}}
}
