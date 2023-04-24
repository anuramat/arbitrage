package main

import (
	"fmt"
	"strings"
	"sync"

	"github.com/anuramat/arbitrage/internal/exchanges/gate"
	"github.com/anuramat/arbitrage/internal/exchanges/okx"
	"github.com/anuramat/arbitrage/internal/models"
	"github.com/anuramat/arbitrage/internal/strategy"
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
	exchanges := map[string]models.Exchange{"gate": &gate.Gate{}, "okx": &okx.Okx{}}
	for name, exchange := range exchanges {
		currencyPairs := viper.GetStringSlice(name + ".currencyPairs")
		if len(currencyPairs) == 0 {
			continue
		}
		exchange.MakeMarkets(currencyPairs, &allMarkets)
		go exchange.Subscribe(currencyPairs)
		fmt.Println("Started " + name + " exchange for currency pairs: " + strings.Join(currencyPairs, ", "))
	}

	fmt.Println("Exchanges started")

	// arbitrage goes here
	//wg.Add(1)
	go strategy.DetectArbitrage(&allMarkets)
	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}
