package gate

import (
	"log"

	"github.com/anuramat/arbitrage/internal/exchanges"
)

type Gate struct {
	exchanges.BaseExchange
}

func (r *Gate) Subscribe(currencyPairs []string, logger *log.Logger) {
	// each currency pair updates in real time
	go r.priceUpdater(currencyPairs, logger)
}

func New() *Gate {
	return &Gate{exchanges.BaseExchange{Name: "gate"}}
}
