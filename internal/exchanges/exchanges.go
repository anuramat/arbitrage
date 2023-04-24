package exchanges

import (
	"github.com/anuramat/arbitrage/internal/models"
)

type BaseExchange struct {
	Markets models.ExchangeMarkets
}

func (r *BaseExchange) Subscribe(currencyPairs []string) {
	panic("not implemented")
}

func (r *BaseExchange) MakeMarkets(currencyPairs []string, allMarkets *models.AllMarkets) {
	r.Markets = make(models.ExchangeMarkets)
	for _, currencyPair := range currencyPairs {
		newMarket := &models.Market{
			Exchange:  r,
			OrderBook: models.OrderBook{Bids: []models.OrderBookEntry{}, Asks: []models.OrderBookEntry{}},
			BestPrice: models.BestPrice{},
		}
		r.Markets[currencyPair] = newMarket
		(*allMarkets)[currencyPair] = append((*allMarkets)[currencyPair], newMarket)
	}
}
