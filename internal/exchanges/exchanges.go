// This package provides implementations of Exchange interface for
// different exchanges. One exchange per file.
package exchanges

import (
	"github.com/anuramat/arbitrage/internal/models"
)

var AllMarkets models.AllMarkets
var Exchanges []models.Exchange
var CurrencyPairs map[string]struct{}

func init() {
	AllMarkets = make(models.AllMarkets)
	Exchanges = []models.Exchange{}
	// currency pairs that have two or more exchanges
	CurrencyPairs = map[string]struct{}{"BTC_USDT": {}} // TODO programmatically figure out from config
}
