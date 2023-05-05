package strategy

import (
	"fmt"
	"time"

	"github.com/anuramat/arbitrage/internal/models"
	"github.com/shopspring/decimal"
)

func DetectArbitrage(allMarkets *models.AllMarkets) {
	counter := 0
	fmt.Println("Waiting for exchanges to start...")
	time.Sleep(5 * time.Second)
	for {
		time.Sleep(1 * time.Second)
		counter++
		fmt.Printf("Checking for arbitrage opportunities... (%d)\n", counter)
		highestProfit := decimal.Zero
		highestProfitPair := ""
		bidExchange := ""
		askExchange := ""
		for currencyPair, markets := range *allMarkets {
			markets[0].BestPrice.RLock()
			highestBid := markets[0].BestPrice.Bid
			lowestAsk := markets[0].BestPrice.Ask
			highestBidExchange := markets[0].Exchange.GetName()
			lowestAskExchange := markets[0].Exchange.GetName()
			markets[0].BestPrice.RUnlock()
			for _, market := range markets[1:] {
				market.BestPrice.RLock()
				if market.BestPrice.Bid.GreaterThan(highestBid) {
					highestBid = market.BestPrice.Bid
					highestBidExchange = market.Exchange.GetName()
				}
				if market.BestPrice.Ask.LessThan(lowestAsk) {
					lowestAsk = market.BestPrice.Ask
					lowestAskExchange = market.Exchange.GetName()
				}
				market.BestPrice.RUnlock()
			}
			// calculate profit
			if highestBid.IsZero() || lowestAsk.IsZero() {
				continue
			}
			profit := highestBid.Sub(lowestAsk)
			profit = profit.Div(lowestAsk.Add(highestBid))
			if profit.GreaterThan(highestProfit) {
				highestProfit = profit
				highestProfitPair = currencyPair
				bidExchange = highestBidExchange
				askExchange = lowestAskExchange
			}

		}
		fmt.Printf("Highest profit: %s on %s, %s/%s\n", highestProfit.StringFixed(8), highestProfitPair, bidExchange, askExchange)
	}
}
