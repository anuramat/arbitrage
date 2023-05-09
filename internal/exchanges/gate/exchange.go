package gate

import (
	"log"

	"github.com/anuramat/arbitrage/internal/exchanges"
	"github.com/anuramat/arbitrage/internal/models"
)

type Gate struct {
	exchanges.BaseExchange
}

func (r *Gate) Subscribe(currencyPairs []string, logger *log.Logger, updateChannel chan<- models.UpdateNotification) {
	// each currency pair updates in real time
	go r.updater(currencyPairs, logger, updateChannel)
}

func New() *Gate {
	return &Gate{exchanges.BaseExchange{Name: "gate"}}
}
