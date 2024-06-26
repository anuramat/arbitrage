package exchanges

import (
	"log"

	"github.com/anuramat/arbitrage/internal/models"
)

type BaseExchange struct {
	Markets models.ExchangeMarkets
	Name    string
}

func (r *BaseExchange) Subscribe(currencyPairs []string, logger *log.Logger, updateChannel chan<- models.UpdateNotification) {
	panic("not implemented")
}

func (r *BaseExchange) GetName() string {
	return r.Name
}

func (r *BaseExchange) GetMarkets() *models.ExchangeMarkets {
	return &r.Markets
}

func (r *BaseExchange) MakeMarkets(currencyPairs []string, allMarkets *models.AllMarkets) {
	r.Markets = make(models.ExchangeMarkets)
	for _, currencyPair := range currencyPairs {
		newMarket := &models.Market{
			Exchange:  r,
			OrderBook: models.OrderBook{Bids: []models.OrderBookEntry{}, Asks: []models.OrderBookEntry{}},
			BestPrice: models.BestPrice{},
			Index:     len((*allMarkets)[currencyPair]),
		}
		r.Markets[currencyPair] = newMarket
		(*allMarkets)[currencyPair] = append((*allMarkets)[currencyPair], newMarket)
	}
}
