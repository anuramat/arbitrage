// This package provides implementations of Exchange interface for
// different exchanges. One exchange per file.
package exchanges

import "github.com/anuramat/arbitrage/internal/models"

var allMarkets models.AllMarkets
var Exchanges []models.Exchange

func init() {
	allMarkets = make(models.AllMarkets)
	Exchanges = []models.Exchange{}
}
