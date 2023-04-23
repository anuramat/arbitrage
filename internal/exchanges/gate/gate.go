package gate

import (
	"context"
	"sync"

	"github.com/anuramat/arbitrage/internal/models"
)

type Gate struct {
	markets models.ExchangeMarkets
}

func (r *Gate) Subscribe(ctx context.Context, wg *sync.WaitGroup, currencyPairs []string) {
	defer wg.Done()
	wg.Add(1)
	go r.priceUpdater(ctx, wg, currencyPairs)
}

func New() *Gate {
	gate := Gate{}
	gate.markets = make(models.ExchangeMarkets)
	return &gate
}

func (r *Gate) MakeMarkets(currencyPairs []string, allMarkets *models.AllMarkets) {
	for _, currencyPair := range currencyPairs {
		newMarket := &models.Market{
			Exchange:  r,
			OrderBook: models.OrderBook{Bids: []models.OrderBookEntry{}, Asks: []models.OrderBookEntry{}},
			BestPrice: models.BestPrice{},
		}
		r.markets[currencyPair] = newMarket
		(*allMarkets)[currencyPair] = append((*allMarkets)[currencyPair], newMarket)
	}
}
