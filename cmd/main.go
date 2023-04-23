package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/anuramat/arbitrage/internal/exchanges/gate"
	"github.com/anuramat/arbitrage/internal/models"
	"github.com/spf13/viper"
)

func main() {
	allMarkets := make(models.AllMarkets)

	viper.SetConfigFile("config.toml")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	// read configs, start exchange goroutines
	exchanges := map[string]models.Exchange{"gate": gate.New()}
	for name, exchange := range exchanges {
		currencyPairs := viper.GetStringSlice(name + ".currencyPairs")
		if len(currencyPairs) == 0 {
			continue
		}
		exchange.MakeMarkets(currencyPairs, &allMarkets)
		wg.Add(1)
		go exchange.Subscribe(ctx, wg, currencyPairs)
	}

	// arbitrage goes here

	// Ctrl-C will close the program gracefully
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	cancel()
	wg.Wait()
}
