package exchanges

import (
	"context"
	"sync"

	"github.com/anuramat/arbitrage/internal/models"
)

type okx struct {
	markets models.ExchangeMarkets
}

func init() {
	okx := &okx{}
	okx.markets = make(models.ExchangeMarkets)
	currencyPairs := []string{"BTC-USD"} // TODO get all currency pairs from okx
	for _, currencyPair := range currencyPairs {
		market := &models.Market{}
		market.Exchange = okx
		okx.markets[currencyPair] = market
		allMarkets[currencyPair] = append(allMarkets[currencyPair], market)
	}
	Exchanges = append(Exchanges, okx)
}

func (r okx) StartUpdates(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		// TODO Do work
		select {
		case <-ctx.Done():
			// TODO Context canceled
			return
		default:
			// TODO Continue working
		}
	}
}
