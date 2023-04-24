package strategy

import (
	"fmt"
	"time"

	"github.com/anuramat/arbitrage/internal/models"
)

func DetectArbitrage(allMarkets *models.AllMarkets) {
	counter := 0
	for {
		time.Sleep(1 * time.Second)
		counter++
		fmt.Printf("Checking for arbitrage... (%d)\n", counter)
		for currencyPair, markets := range *allMarkets {
			if len(markets) < 2 {
				continue
			}
			markets[0].BestPriceValue.RWMutex.RLock()
			highestBid := markets[0].BestPriceValue.Bid
			lowestAsk := markets[0].BestPriceValue.Ask
			markets[0].BestPriceValue.RWMutex.RUnlock()
			for _, market := range markets[1:] {
				market.BestPriceValue.RWMutex.RLock()
				if market.BestPriceValue.Bid.GreaterThan(highestBid) {
					highestBid = market.BestPriceValue.Bid
				}
				if market.BestPriceValue.Ask.LessThan(lowestAsk) {
					lowestAsk = market.BestPriceValue.Ask
				}
				market.BestPriceValue.RWMutex.RUnlock()
			}
			if highestBid.GreaterThan(lowestAsk) {
				// arbitrage detected
				fmt.Printf("Opportunity detected, %s: %s > %s\n", currencyPair, highestBid.String(), lowestAsk.String())
			}

		}
	}
}
