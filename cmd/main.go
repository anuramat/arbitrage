package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/anuramat/arbitrage/internal/analysis"
	"github.com/anuramat/arbitrage/internal/exchanges/gate"
	"github.com/anuramat/arbitrage/internal/exchanges/okx"
	"github.com/anuramat/arbitrage/internal/exchanges/whitebit"
	"github.com/anuramat/arbitrage/internal/models"
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

	debugTextView := tview.NewTextView().SetChangedFunc(func() { app.Draw() })
	debugTextView.SetBorder(true).SetTitle("debug").SetTitleAlign(tview.AlignCenter)
	debugLogger := log.New(debugTextView, "", log.LstdFlags)

	infoTextView := tview.NewTextView().SetChangedFunc(func() { app.Draw() })
	infoTextView.SetBorder(true).SetTitle("info").SetTitleAlign(tview.AlignCenter)
	infoLogger := log.New(infoTextView, "", log.LstdFlags)

	textViewFlex := tview.NewFlex().SetDirection(tview.FlexRow).AddItem(infoTextView, 0, 2, false).AddItem(debugTextView, 0, 2, false)

	flex := tview.NewFlex().AddItem(table, 92, 3, true).AddItem(textViewFlex, 0, 2, false) // XXX table fixed width, flexible info/debug
	// TODO make focus switch on click

	// start exchange goroutines, apply configs
	updateChannel := make(chan models.UpdateNotification, 100)
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
		go exchange.Subscribe(currencyPairs, debugLogger, updateChannel)
		fmt.Println("Started " + exchange.GetName() + " exchange for currency pairs: " + strings.Join(currencyPairs, ", "))
	}

	// start showing updates
	go analysis.TableUpdater(&allMarkets, exchanges, app, table, debugLogger, updateChannel)

	// start opportunity detector
	go analysis.AbsoluteDetector(&allMarkets, infoLogger)

	// start tview
	if err := app.SetRoot(flex, true).SetFocus(table).Run(); err != nil {
		panic(err)
	}
}
