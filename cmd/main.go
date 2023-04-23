package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/anuramat/arbitrage/internal/exchanges"
	_ "github.com/anuramat/arbitrage/internal/exchanges/gate"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	// this channel will receive a signal when the program is interrupted
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// start update goroutines
	for _, exchange := range exchanges.Exchanges {
		wg.Add(1)
		go exchange.Subscribe(ctx, wg)
	}

	// TODO start goroutine for arbitrage

	// wait for a termination signal, then close goroutines
	<-c
	cancel()
	wg.Wait()
}
