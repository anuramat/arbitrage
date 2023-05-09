package analysis

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/anuramat/arbitrage/internal/models"
	"github.com/rivo/tview"
)

const (
	colWidth = 25
)

func TableUpdater(allMarkets *models.AllMarkets, exchanges []models.Exchange, app *tview.Application, table *tview.Table, logger *log.Logger, updateChannel <-chan models.UpdateNotification) {
	// make table
	table.SetFixed(1, 1)
	table.SetCell(0, 0, tview.NewTableCell(""))

	pairRow := make(map[string]int)     // maps currency pair to row number
	exchangeCol := make(map[string]int) // maps exchange name to column number

	// left column
	for i, pair := range sortedCurrencyPairs(allMarkets) {
		pairRow[pair] = i + 1
		table.SetCell(i+1, 0, tview.NewTableCell(pair))
		i++
	}

	// top row
	for i, exchange := range exchanges {
		name := exchange.GetName()
		displayName := name
		if len(name) < colWidth {
			displayName = name + strings.Repeat(" ", colWidth-len(name))
		}
		exchangeCol[name] = i + 1
		table.SetCell(0, i+1, tview.NewTableCell(displayName))
	}

	// fill table with empty cells
	for row := 1; row < len(pairRow)+1; row++ {
		for column := 1; column < len(exchanges)+1; column++ {
			table.SetCell(row, column, tview.NewTableCell("").SetMaxWidth(colWidth))
		}
	}

	// update price cells
	for updateNotification := range updateChannel {
		pair := updateNotification.Pair
		exchangeIndex := updateNotification.ExchangeIndex
		name := updateNotification.ExchangeName

		market := (*allMarkets)[pair][exchangeIndex]

		market.BestPrice.RLock()
		ask := market.BestPrice.Ask
		bid := market.BestPrice.Bid
		market.BestPrice.RUnlock()

		text := fmt.Sprintf("%v/%v", bid, ask)
		table.GetCell(pairRow[pair], exchangeCol[name]).SetText(text)

		app.Draw()
	}
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

}

func sortedCurrencyPairs(allMarkets *models.AllMarkets) []string {
	pairs := []string{}
	for pair := range *allMarkets {
		pairs = append(pairs, pair)
	}
	sort.Strings(pairs)
	return pairs
}
