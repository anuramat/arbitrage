package analysis

import (
	"log"

	"github.com/anuramat/arbitrage/internal/models"
	"github.com/shopspring/decimal"
)

func AbsoluteDetectorCycle(allMarkets *models.AllMarkets, logger *log.Logger) {
	// TODO set frequency
	// basic idea: for current exchange, take its lowest ask, and compare it to the highest bid of all other exchanges
	// if highest bid is higher than lowest ask, then iterate through orderbooks
	//
	bestProfit := decimal.Zero
	bestBidExchange := ""
	bestAskExchange := ""
	bestPair := ""
	for pair, markets := range *allMarkets {
		for _, market := range markets {
			market.BestPrice.RLock()
			ask := market.BestPrice.Ask
			market.BestPrice.RUnlock()

			if ask.IsZero() {
				continue
			}

			for _, market2 := range markets {
				market2.BestPrice.RLock()
				bid := market2.BestPrice.Bid
				market2.BestPrice.RUnlock()

				if bid.IsZero() {
					continue
				}

				// check if there is any arbitrage opportunity at all, if not, continue
				if !bid.Sub(ask).IsPositive() {
					continue
				}

				// calculate absolute profit
				absoluteProfit := bid.Sub(ask).Div(ask).Mul(decimal.NewFromInt(100))
				if bestProfit.LessThan(absoluteProfit) {
					bestProfit = absoluteProfit
					bestBidExchange = market2.Exchange.GetName()
					bestAskExchange = market.Exchange.GetName()
					bestPair = pair
				}
			}
		}

	}
	if !bestProfit.IsZero() {
		logger.Printf("Biggest opportunity in absolute terms: %v:%v B:%v/A:%v", bestProfit, bestBidExchange, bestAskExchange, bestPair)
	} else {
		logger.Printf("No opportunity found")
	}
}

func CalculateAbsoluteProfit(bids, asks []models.OrderBookEntry) decimal.Decimal {
	// TODO
	return decimal.Zero
}

// func bestAbsoluteOpportunityDetector(freq time.Duration, allMarkets *models.AllMarkets, logger *log.Logger) {
// 	for pair, markets := range *allMarkets {
// 		for _, market := range markets {

// 		}
// 	}
// }

// update price cells
// go func() {

// 	for {
// 		maxProfit := decimal.Zero
// 		opportunityString := ""
// 		for row := 1; row < rows; row++ {
// 			currency := currencyPairs[row-1]
// 			highestBid := decimal.Zero
// 			lowestAsk := decimal.NewFromFloat(math.MaxFloat64)
// 			highestBidExchange := ""
// 			lowestAskExchange := ""
// 			currentProfit := decimal.Zero
// 			for column := 1; column < cols; column++ {
// 				exchange := exchanges[column-1]
// 				market, ok := (*exchange.GetMarkets())[currency]
// 				if !ok {
// 					continue
// 				}
// 				market.BestPrice.RLock()
// 				ask := market.BestPrice.Ask
// 				bid := market.BestPrice.Bid
// 				market.BestPrice.RUnlock()

// 				text := fmt.Sprintf("%v/%v", bid, ask)
// 				table.GetCell(row, column).SetText(text)

// 				if bid.GreaterThan(highestBid) && !bid.IsZero() {
// 					highestBid = bid
// 					highestBidExchange = exchange.GetName()
// 				}
// 				if ask.LessThan(lowestAsk) && !ask.IsZero() {
// 					lowestAsk = ask
// 					lowestAskExchange = exchange.GetName()
// 				}
// 			}
// 			if highestBid.GreaterThan(decimal.Zero) && lowestAsk.LessThan(decimal.NewFromFloat(math.MaxFloat64)) {
// 				currentProfit = highestBid.Sub(lowestAsk).Div(lowestAsk).Mul(decimal.NewFromInt(100))
// 				if currentProfit.GreaterThan(maxProfit) {
// 					maxProfit = currentProfit
// 					opportunityString = fmt.Sprintf("%v  %v/%v  %v/%v  %v%%", currency, lowestAsk, highestBid, lowestAskExchange, highestBidExchange, maxProfit)
// 				}
// 			}
// 		}
// 		if len(opportunityString) != 0 {
// 			logger.Println(opportunityString)
// 		}
// 		app.Draw()
// 		time.Sleep(1000 * time.Millisecond)
// 	}
// }()
