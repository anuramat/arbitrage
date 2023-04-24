package okx

import (
	"context"
	"sync"

	"github.com/anuramat/arbitrage/internal/exchanges"
)

type Okx struct {
	exchanges.BaseExchange
}

// XXX might need pinger

func (r *Okx) Subscribe(ctx context.Context, wg *sync.WaitGroup, currencyPairs []string) {
	defer wg.Done()

	wg.Add(1)
	r.priceUpdater(ctx, wg, currencyPairs)
}
