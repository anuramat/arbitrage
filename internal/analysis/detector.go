package analysis

import (
	"log"
	"time"

	"github.com/anuramat/arbitrage/internal/models"
	"github.com/shopspring/decimal"
)

func AbsoluteDetector(allMarkets *models.AllMarkets, logger *log.Logger) {
	for {
		AbsoluteDetectorOneCycle(allMarkets, logger)
		// TODO parameterize frequency
		// TODO rwlock to rlock
		time.Sleep(1 * time.Second)
	}
}

func AbsoluteDetectorOneCycle(allMarkets *models.AllMarkets, logger *log.Logger) {
	// TODO set frequency
	// basic idea: for current exchange, take its lowest ask, and compare it to the highest bid of all other exchanges
	// if highest bid is higher than lowest ask, then iterate through orderbooks
	//
	bestProfit := decimal.Zero
	bestAmount := decimal.Zero
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

				profit, amount := CalculateProfit(bids, asks)

				if bestProfit.LessThan(profit) {
					bestProfit = profit
					bestAmount = amount
					bestBidExchange = market2.Exchange.GetName()
					bestAskExchange = market.Exchange.GetName()
					bestPair = pair
				}
			}
		}
	}
	if !bestProfit.IsZero() {
		logger.Printf("Biggest opportunity in absolute terms: %v/%v (B:%v/A:%v) [%v]", bestProfit, bestAmount, bestBidExchange, bestAskExchange, bestPair)
	} else {
		logger.Printf("No opportunity found")
	}
}

func CalculateProfit(bids, asks []models.OrderBookEntry) (profit, amount decimal.Decimal) {
	if len(bids) == 0 || len(asks) == 0 {
		return decimal.Zero, decimal.Zero
	}
	i_asks := 0
	received := decimal.Zero
	sent := decimal.Zero
	leftover := asks[0].Amount
outer:
	for i_bids := 0; i_bids < len(bids); i_bids++ {
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
	if sent.GreaterThan(received) {
		panic("non-positive profit")
	}
	return received.Sub(sent), received.Add(sent)
}
