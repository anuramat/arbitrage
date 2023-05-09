package okx

import (
	"log"

	"github.com/anuramat/arbitrage/internal/exchanges"
	"github.com/anuramat/arbitrage/internal/models"
)

type Okx struct {
	exchanges.BaseExchange
}

// XXX might need pinger

func (r *Okx) Subscribe(currencyPairs []string, logger *log.Logger, updateChannel chan<- models.UpdateNotification) {
	// each pair updates every 10ms if there is a change
	r.priceUpdater(currencyPairs, logger, updateChannel)
}

func New() *Okx {
	return &Okx{exchanges.BaseExchange{Name: "okx"}}
}
