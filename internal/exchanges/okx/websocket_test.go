package okx

import (
	"fmt"
	"testing"

	"github.com/anuramat/arbitrage/internal/models"
	"github.com/shopspring/decimal"
)

func dec(s string) decimal.Decimal {
	res, _ := decimal.NewFromString(s)
	return res
}

func book(orders [][2]string) []models.OrderBookEntry {
	entries := make([]models.OrderBookEntry, len(orders))
	for i, order := range orders {
		price := dec(order[0])
		amount := dec(order[1])
		entries[i] = models.OrderBookEntry{Price: price, Amount: amount}
	}
	return entries
}

func Test_mergeBooks(t *testing.T) {

	// all in one: tests deletion, replacement, appending in the middle, beginning, and the end
	asks_updates := book([][2]string{{"0.1", "0.42"}, {"0.2", "0.12894"}, {"5.1", "21341"}, {"6", "124"}, {"200", "12312"}, {"250", "23"}})
	asks_book := book([][2]string{{"1", "0.34534"}, {"5.1", "2.4123"}, {"150", "0.34"}})
	comparator := func(a, b decimal.Decimal) bool { return a.LessThan(b) }
	res := mergeBooks(asks_updates, asks_book, comparator)
	want := "[{0.1 0.42} {0.2 0.12894} {1 0.34534} {5.1 21341} {6 124} {150 0.34} {200 12312} {250 23}]"
	have := fmt.Sprintf("%v", res)
	if want != have {
		t.Errorf("want != have: %v != %v", want, have)
	}

}
