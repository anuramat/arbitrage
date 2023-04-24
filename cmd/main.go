package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/anuramat/arbitrage/internal/exchanges/gate"
	"github.com/anuramat/arbitrage/internal/exchanges/okx"
	"github.com/anuramat/arbitrage/internal/models"
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

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	// read configs, start exchange goroutines
	exchanges := map[string]models.Exchange{"gate": &gate.Gate{}, "okx": &okx.Okx{}}
	for name, exchange := range exchanges {
		currencyPairs := viper.GetStringSlice(name + ".currencyPairs")
		if len(currencyPairs) == 0 {
			continue
		}
		exchange.MakeMarkets(currencyPairs, &allMarkets)
		wg.Add(1)
		go exchange.Subscribe(ctx, wg, currencyPairs)
		fmt.Println("Started " + name + " exchange for currency pairs: " + strings.Join(currencyPairs, ", "))
	}

	fmt.Println("Exchanges started")

	// arbitrage goes here

	// Ctrl-C will close the program gracefully
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	cancel()
	wg.Wait()
}
