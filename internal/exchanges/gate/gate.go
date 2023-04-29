package gate

import (
	"github.com/anuramat/arbitrage/internal/exchanges"
)

type Gate struct {
	exchanges.BaseExchange
}

func (r *Gate) Subscribe(currencyPairs []string) {
	go r.priceUpdater(currencyPairs)
}

func New() *Gate {
	return &Gate{exchanges.BaseExchange{Name: "gate.io"}}
}
