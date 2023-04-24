package strategy

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/anuramat/arbitrage/internal/models"
)

func DetectArbitrage(ctx context.Context, wg *sync.WaitGroup, allMarkets *models.AllMarkets) {
	defer wg.Done()
	counter := 0
	for {
		time.Sleep(1 * time.Second)
		counter++
		fmt.Printf("Checking for arbitrage... (%d)\n", counter)
		for currencyPair, markets := range *allMarkets {
			select {
			case <-ctx.Done():
				return
			default:
				if len(markets) < 2 {
					continue
				}
				markets[0].BestPrice.RWMutex.RLock()
				highestBid := markets[0].BestPrice.Bid
				lowestAsk := markets[0].BestPrice.Ask
				markets[0].BestPrice.RWMutex.RUnlock()
				for _, market := range markets[1:] {
					market.BestPrice.RWMutex.RLock()
					if market.BestPrice.Bid.GreaterThan(highestBid) {
						highestBid = market.BestPrice.Bid
					}
					if market.BestPrice.Ask.LessThan(lowestAsk) {
						lowestAsk = market.BestPrice.Ask
					}
					market.BestPrice.RWMutex.RUnlock()
				}
				if highestBid.GreaterThan(lowestAsk) {
					// arbitrage detected
					fmt.Printf("Opportunity detected, %s: %s > %s\n", currencyPair, highestBid.String(), lowestAsk.String())
				}
			}
		}
	}
}
