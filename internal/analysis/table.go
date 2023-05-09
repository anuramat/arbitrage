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

}

func sortedCurrencyPairs(allMarkets *models.AllMarkets) []string {
	pairs := []string{}
	for pair := range *allMarkets {
		pairs = append(pairs, pair)
	}
	sort.Strings(pairs)
	return pairs
}
