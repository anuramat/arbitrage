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
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	// this channel will receive a signal when the program is interrupted
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	allMarkets := make(models.AllMarkets)

	viper.SetConfigFile("config.toml")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	exchanges := map[string]models.Exchange{"gate": gate.New()}

	for name, exchange := range exchanges {
		currencyPairs := viper.GetStringSlice(name + ".currencyPairs")
		if len(currencyPairs) == 0 {
			continue
		}
		for _, currencyPair := range currencyPairs {
			newMarket := exchange.NewMarket(currencyPair)
			allMarkets[currencyPair] = append(allMarkets[currencyPair], newMarket)
		}
	}

	// start update goroutines
	for _, exchange := range exchanges {
		wg.Add(1)
		go exchange.Subscribe(ctx, wg)
	}

	// TODO start goroutine for arbitrage

	// wait for a termination signal, then close goroutines
	<-c
	cancel()
	wg.Wait()
}
