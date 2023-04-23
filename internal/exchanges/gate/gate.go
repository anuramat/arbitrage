package gate

import (
	"context"
	"sync"

	"github.com/anuramat/arbitrage/internal/exchanges"
	"github.com/anuramat/arbitrage/internal/models"
)

type gate struct {
	markets       models.ExchangeMarkets
	currencyPairs []string // active currency pairs
}

func init() {
	gate := &gate{}
	gate.markets = make(models.ExchangeMarkets)
	currencyPairs := []string{"BTC_USDT"} // TODO move to config
	// only add the currency pairs that are in config on two or more exchanges
	for _, currencyPair := range currencyPairs {
		if _, ok := exchanges.CurrencyPairs[currencyPair]; ok {
			gate.currencyPairs = append(gate.currencyPairs, currencyPair)
		}
	}
	if len(gate.currencyPairs) == 0 {
		return
	}
	for _, currencyPair := range gate.currencyPairs {
		market := &models.Market{}
		market.Exchange = gate
		gate.markets[currencyPair] = market
		exchanges.AllMarkets[currencyPair] = append(exchanges.AllMarkets[currencyPair], market)
	}
	exchanges.Exchanges = append(exchanges.Exchanges, gate)
}

func (r gate) Subscribe(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	if len(r.currencyPairs) == 0 {
		return
	}
	wg.Add(1)
	go r.priceUpdater(ctx, wg)
}
