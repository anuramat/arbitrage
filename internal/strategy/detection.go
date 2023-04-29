package strategy

import (
	"fmt"
	"time"

	"github.com/anuramat/arbitrage/internal/models"
)

func DetectArbitrage(allMarkets *models.AllMarkets) {
	counter := 0
	fmt.Println("Waiting for exchanges to start...")
	time.Sleep(3 * time.Second)
	for {
		time.Sleep(1 * time.Second)
		counter++
		fmt.Printf("Checking for arbitrage opportunities... (%d)\n", counter)
		for currencyPair, markets := range *allMarkets {

			markets[0].BestPrice.RLock()
			highestBid := markets[0].BestPrice.Bid
			lowestAsk := markets[0].BestPrice.Ask
			if currencyPair == "TUSD_USDT" {
				fmt.Println(markets[0].BestPrice.Bid)
				fmt.Println(markets[0].BestPrice.Ask)
			}
			markets[0].BestPrice.RUnlock()
			for _, market := range markets[1:] {
				market.BestPrice.RLock()
				if market.BestPrice.Bid.GreaterThan(highestBid) {
					highestBid = market.BestPrice.Bid
				}
				if market.BestPrice.Ask.LessThan(lowestAsk) {
					lowestAsk = market.BestPrice.Ask
				}
				if currencyPair == "TUSD_USDT" {
					fmt.Println(markets[1].BestPrice.Bid)
					fmt.Println(markets[1].BestPrice.Ask)
				}
				market.BestPrice.RUnlock()
			}
		}
	}
}
