package gate

import (
	"context"
	"sync"

	"github.com/anuramat/arbitrage/internal/models"
)

type Gate struct {
	markets       models.ExchangeMarkets
	currencyPairs []string
}

func (r *Gate) Subscribe(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	wg.Add(1)
	go r.priceUpdater(ctx, wg)
}

func (r *Gate) GetMarkets() *models.ExchangeMarkets {
	return &r.markets
}

func (r *Gate) NewMarket(currencyPair string) *models.Market {
	newMarket := &models.Market{}
	r.currencyPairs = append(r.currencyPairs, currencyPair)
	r.markets[currencyPair] = newMarket
	return newMarket
}

func New() *Gate {
	gate := Gate{}
	gate.markets = make(models.ExchangeMarkets)
	return &gate
}
