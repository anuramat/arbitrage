package analysis

// func triggerDetector(allMarkets *models.AllMarkets, updateChannel <-chan models.UpdateNotification, logger *log.Logger) {
// 	// TODO implement: on each change check the trigger as described in the proposal
// }

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
