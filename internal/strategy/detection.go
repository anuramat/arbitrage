package strategy

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/anuramat/arbitrage/internal/models"
	"github.com/rivo/tview"
	"github.com/shopspring/decimal"
)

func TableUpdater(allMarkets *models.AllMarkets, exchanges []models.Exchange, app *tview.Application, table *tview.Table, logger *log.Logger) {

	currencyPairs := []string{}
	for currencyPair := range *allMarkets {
		currencyPairs = append(currencyPairs, currencyPair)
	}
	rows := len(currencyPairs) + 1
	cols := len(exchanges) + 1

	// make headers
	for row := 1; row < rows; row++ {
		table.SetCell(row, 0, tview.NewTableCell(currencyPairs[row-1]))
	}

	for column := 1; column < cols; column++ {
		name := exchanges[column-1].GetName()
		table.SetCell(0, column, tview.NewTableCell(name))
	}

	table.SetCell(0, 0, tview.NewTableCell(""))

	// make price cells
	for row := 1; row < rows; row++ {
		for column := 1; column < cols; column++ {
			table.SetCell(row, column, tview.NewTableCell(""))
		}
	}

	// update price cells
	go func() {

		for {
			maxProfit := decimal.Zero
			opportunityString := ""
			for row := 1; row < rows; row++ {
				currency := currencyPairs[row-1]
				highestBid := decimal.Zero
				lowestAsk := decimal.NewFromFloat(math.MaxFloat64)
				highestBidExchange := ""
				lowestAskExchange := ""
				currentProfit := decimal.Zero
				for column := 1; column < cols; column++ {
					exchange := exchanges[column-1]
					market, ok := (*exchange.GetMarkets())[currency]
					if !ok {
						continue
					}
					market.BestPrice.RLock()
					ask := market.BestPrice.Ask
					bid := market.BestPrice.Bid
					market.BestPrice.RUnlock()

					text := fmt.Sprintf("%v/%v", bid, ask)
					table.GetCell(row, column).SetText(text)

					if bid.GreaterThan(highestBid) && !bid.IsZero() {
						highestBid = bid
						highestBidExchange = exchange.GetName()
					}
					if ask.LessThan(lowestAsk) && !ask.IsZero() {
						lowestAsk = ask
						lowestAskExchange = exchange.GetName()
					}
				}
				if highestBid.GreaterThan(decimal.Zero) && lowestAsk.LessThan(decimal.NewFromFloat(math.MaxFloat64)) {
					currentProfit = highestBid.Sub(lowestAsk).Div(lowestAsk).Mul(decimal.NewFromInt(100))
					if currentProfit.GreaterThan(maxProfit) {
						maxProfit = currentProfit
						opportunityString = fmt.Sprintf("%v  %v/%v  %v/%v  %v%%", currency, lowestAsk, highestBid, lowestAskExchange, highestBidExchange, maxProfit)
					}
				}
			}
			if len(opportunityString) != 0 {
				logger.Println(opportunityString)
			}
			app.Draw()
			time.Sleep(1000 * time.Millisecond)
		}
	}()

}
