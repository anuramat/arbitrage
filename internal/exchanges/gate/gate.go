package gate

import (
	"context"
	"sync"

	"github.com/anuramat/arbitrage/internal/exchanges"
	"github.com/anuramat/arbitrage/internal/models"
)

type gate struct {
	markets       models.ExchangeMarkets
	currencyPairs []string // TODO move to config, get it from there instead
}

func init() {
	gate := &gate{}
	gate.markets = make(models.ExchangeMarkets)
	gate.currencyPairs = []string{"BTC_USDT", "ETH_USDT"} // TODO move to config
	for _, currencyPair := range gate.currencyPairs {
		market := &models.Market{}
		market.Exchange = gate
		gate.markets[currencyPair] = market
		exchanges.AllMarkets[currencyPair] = append(exchanges.AllMarkets[currencyPair], market)
	}
	exchanges.Exchanges = append(exchanges.Exchanges, gate)
}

func (r gate) Subscribe(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go r.priceUpdater(ctx, wg)
}
