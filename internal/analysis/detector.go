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
				market.OrderBook.RLock()
				asks := make([]models.OrderBookEntry, len(market.OrderBook.Asks))
				copy(asks, market.OrderBook.Asks)
				market.OrderBook.RUnlock()

				market2.OrderBook.RLock()
				bids := make([]models.OrderBookEntry, len(market2.OrderBook.Bids))
				copy(bids, market2.OrderBook.Bids)
				market2.OrderBook.RUnlock()

				absoluteProfit := CalculateAbsoluteProfit(bids, asks)

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
	i_asks := 0
	received := decimal.Zero
	sent := decimal.Zero
	leftover := asks[0].Amount
outer:
	for i_bids := 0; i_bids < len(bids) && i_asks < len(asks); i_bids++ {
		subtotal := bids[i_bids].Amount
		for i_asks < len(asks) {
			if asks[i_asks].Price.GreaterThanOrEqual(bids[i_bids].Price) {
				break outer
			}
			if subtotal.LessThanOrEqual(leftover) {
				leftover = asks[i_asks].Amount.Sub(subtotal)
				received = received.Add(subtotal.Mul(bids[i_bids].Price))
				sent = sent.Add(subtotal.Mul(asks[i_asks].Price))
				if leftover.IsZero() {
					i_asks++
					leftover = asks[i_asks].Amount
				}
				break
			}
			subtotal = subtotal.Sub(asks[i_asks].Amount)
			received = received.Add(asks[i_asks].Amount.Mul(bids[i_bids].Price))
			sent = sent.Add(asks[i_asks].Amount.Mul(asks[i_asks].Price))
			i_asks++
			leftover = asks[i_asks].Amount
		}
	}
	if sent.GreaterThanOrEqual(received) {
		panic("non-positive profit")
	}
	return received.Sub(sent)
}
