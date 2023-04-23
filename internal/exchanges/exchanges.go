// This package provides implementations of Exchange interface for
// different exchanges. One exchange per file.
package exchanges

import (
	"github.com/anuramat/arbitrage/internal/models"
)

var AllMarkets models.AllMarkets
var Exchanges []models.Exchange

func init() {
	AllMarkets = make(models.AllMarkets)
	Exchanges = []models.Exchange{}
}
