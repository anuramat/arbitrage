package gate

import (
	"github.com/anuramat/arbitrage/internal/exchanges"
)

type Gate struct {
	exchanges.BaseExchange
}

func (r *Gate) Subscribe(currencyPairs []string) {
	go r.priceUpdater(currencyPairs)

	// go r.orderBookUpdater(ctx, wg, currencyPairs)
}
