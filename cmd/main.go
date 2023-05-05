package main

import (
	"fmt"
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
	fmt.Println("Starting application, loading config...")

	allMarkets := make(models.AllMarkets)

	viper.SetConfigFile("config.toml")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	fmt.Println("Config loaded, starting exchanges...")

	// read configs, start exchange goroutines
	exchanges := []models.Exchange{gate.New(), okx.New(), whitebit.New()}
	for _, exchange := range exchanges {
		currencyPairs := viper.GetStringSlice(exchange.GetName() + ".currencyPairs")
		if len(currencyPairs) == 0 {
			continue
		}
		exchange.MakeMarkets(currencyPairs, &allMarkets)
		go exchange.Subscribe(currencyPairs)
		fmt.Println("Started " + exchange.GetName() + " exchange for currency pairs: " + strings.Join(currencyPairs, ", "))
	}

	fmt.Println("Exchanges started")

	// arbitrage goes here
	box := tview.NewBox().SetBorder(true).SetTitle("Hello, world!")
	if err := tview.NewApplication().SetRoot(box, true).Run(); err != nil {
		panic(err)
	}
	strategy.DetectArbitrage(&allMarkets)
}
