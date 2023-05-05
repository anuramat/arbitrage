package strategy

import (
	"fmt"
	"strings"
	"time"

	"github.com/anuramat/arbitrage/internal/models"
	"github.com/rivo/tview"
)

func TableUpdater(allMarkets *models.AllMarkets, exchanges []models.Exchange) {
	app := tview.NewApplication()
	table := tview.NewTable().
		SetBorders(true)
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
		if len(name) < 20 {
			name = name + strings.Repeat(" ", 20-len(name))
		}
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
			for row := 1; row < rows; row++ {
				for column := 1; column < cols; column++ {
					exchange := exchanges[column-1]
					market := (*exchange.GetMarkets())[currencyPairs[row-1]]
					market.BestPrice.RLock()
					ask := market.BestPrice.Ask
					bid := market.BestPrice.Bid
					market.BestPrice.RUnlock()
					text := fmt.Sprintf("%v/%v", bid, ask)
					app.QueueUpdateDraw(func() {
						table.GetCell(row, column).SetText(text)
					})
				}
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()

	// main loop
	if err := app.SetRoot(table, true).Run(); err != nil {
		panic(err)
	}

}
