package whitebit

import (
	"log"
	"sync/atomic"

	"github.com/anuramat/arbitrage/internal/exchanges"
	"github.com/anuramat/arbitrage/internal/models"
)

type Whitebit struct {
	exchanges.BaseExchange
	requestId atomic.Int64
}

func (r *Whitebit) Subscribe(currencyPairs []string, logger *log.Logger, updateChannel chan<- models.UpdateNotification) {
	go r.priceUpdater(currencyPairs, logger, updateChannel)
}

func New() *Whitebit {
	return &Whitebit{exchanges.BaseExchange{Name: "whitebit"}, atomic.Int64{}}
}
