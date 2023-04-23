package main

import (
	"context"
	"fmt"
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
	currencyPairsCounts := map[string]int{}

	viper.SetConfigFile("config.toml")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	// allowed exchanges
	exchanges := map[string]models.Exchange{"gate": gate.New()}

	for _, section := range viper.AllKeys() {
		if _, ok := exchanges[section]; !ok {
			fmt.Printf("\"%s\" is not a valid exchange name", section)
			continue
		}
		subViper := viper.Sub(section)
		currencyPairs := subViper.GetStringSlice("currencyPairs")
		for _, currencyPair := range currencyPairs {
			if _, ok := currencyPairsCounts[currencyPair]; !ok {
				currencyPairsCounts[currencyPair] = 1
			} else {
				currencyPairsCounts[currencyPair]++
			}
		}
	}

	allowedCurrencyPairs := map[string]struct{}{}
	for currencyPair, count := range currencyPairsCounts {
		if count > 1 {
			allowedCurrencyPairs[currencyPair] = struct{}{}
		}
	}

	for name, exchange := range exchanges {
		currencyPairs := viper.GetStringSlice(name + ".currencyPairs")
		exchangeAllowedCurrencyPairs := []string{}
		for _, currencyPair := range currencyPairs {
			if _, ok := allowedCurrencyPairs[currencyPair]; ok {
				exchangeAllowedCurrencyPairs = append(exchangeAllowedCurrencyPairs, currencyPair)
			}
		}
		if len(exchangeAllowedCurrencyPairs) == 0 {
			continue
		}
		for _, currencyPair := range exchangeAllowedCurrencyPairs {
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
