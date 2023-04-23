package gate

import (
	"context"
	"sync"

	"github.com/anuramat/arbitrage/internal/exchanges"
)

type Gate struct {
	exchanges.BaseExchange
}

func (r *Gate) Subscribe(ctx context.Context, wg *sync.WaitGroup, currencyPairs []string) {
	defer wg.Done()
	wg.Add(1)
	go r.priceUpdater(ctx, wg, currencyPairs)
}
