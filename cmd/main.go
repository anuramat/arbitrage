package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/anuramat/arbitrage/internal/exchanges/gate"
	"github.com/anuramat/arbitrage/internal/exchanges/okx"
	"github.com/anuramat/arbitrage/internal/exchanges/whitebit"
	"github.com/anuramat/arbitrage/internal/models"
	"github.com/anuramat/arbitrage/internal/strategy"
	"github.com/rivo/tview"
	"github.com/spf13/viper"
)

func main() {
	// read config name from args
	args := os.Args
	if len(args) != 2 {
		fmt.Println("Usage: arbitrage <config>.toml")
		return
	}

	// load configs
	allMarkets := make(models.AllMarkets)
	viper.SetConfigFile(args[1])
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	// initialize tview
	app := tview.NewApplication()
	table := tview.NewTable().
		SetBorders(true)
	info := tview.NewTextView()
	info.SetBorder(true).SetTitle("info").SetTitleAlign(tview.AlignCenter)
	flex := tview.NewFlex().AddItem(table, 0, 1, true).AddItem(info, 0, 1, false)
	logger := log.New(info, "", log.LstdFlags)

	// start exchange goroutines, apply configs
	exchanges := []models.Exchange{gate.New(), okx.New(), whitebit.New()}
	for _, exchange := range exchanges {
		currencyPairs := viper.GetStringSlice("all.currencyPairs")
		if len(currencyPairs) == 0 {
			currencyPairs = viper.GetStringSlice(exchange.GetName() + ".currencyPairs")
		}
		if len(currencyPairs) == 0 {
			continue
		}
		exchange.MakeMarkets(currencyPairs, &allMarkets)
		go exchange.Subscribe(currencyPairs, logger)
		fmt.Println("Started " + exchange.GetName() + " exchange for currency pairs: " + strings.Join(currencyPairs, ", "))
	}

	// start showing updates
	go strategy.TableUpdater(&allMarkets, exchanges, app, table, logger)

	// start tview
	if err := app.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}
}
