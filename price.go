package twigots

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	"github.com/orsinium-labs/enum"
)

type Price struct {
	Currency Currency `json:"currencyCode"`

	// Amount is the cost in Cents, Pennies etc.
	// Prefer using `Number`
	Amount int `json:"amountInCents"`
}

// Number is the numerical value of the price.
// E.g. Dollars, Pounds, Euros etc.
// Use this over `Amount“.
func (p Price) Number() float64 {
	return float64(p.Amount) / 100
}

// The price as a string.
// e.g. $30.62
func (p Price) String() string {
	return priceString(p.Number(), p.Currency)
}

// Add prices together. Currency will be kept.
// Returns a new price.
func (p Price) Add(other Price) Price {
	return Price{
		Currency: p.Currency,
		Amount:   p.Amount + other.Amount,
	}
}

// Subtract price from another. Currency will be kept.
// Returns a new price.
func (p Price) Subtract(other Price) Price {
	return Price{
		Currency: p.Currency,
		Amount:   p.Amount - other.Amount,
	}
}

// Multiply prices together. Returns a new price.
func (p Price) Multiply(num int) Price {
	return Price{
		Currency: p.Currency,
		Amount:   p.Amount * num,
	}
}

// Divide prices. Currency will be kept.
// Returns a new price.
func (p Price) Divide(num int) Price {
	return Price{
		Currency: p.Currency,
		Amount:   int(math.Round(float64(p.Amount) / float64(num))),
	}
}

func priceString(cost float64, currency Currency) string {
	costString := strconv.FormatFloat(cost, 'f', 2, 64)
	currencyString := currency.Symbol()
	if currencyString == "" {
		return costString + currency.Value
	}

	return currencyString + costString
}

var (
	currency = enum.NewBuilder[string, Currency]()

	CurrencyGBP = currency.Add(Currency{"GBP"})

	Currencies = currency.Enum()
)

type Currency enum.Member[string]

// Symbol is the character that represents the currency
// e.g. $, £, €.
func (c Currency) Symbol() string {
	switch c {
	case CurrencyGBP:
		return "£"
	default:
		return ""
	}
}

func (c *Currency) UnmarshalJSON(data []byte) error {
	var currencyString string
	err := json.Unmarshal(data, &currencyString)
	if err != nil {
		return err
	}

	currency := Currencies.Parse(currencyString)
	if currency == nil {
		return fmt.Errorf("currency '%s' is not valid", currencyString)
	}
	*c = *currency
	return nil
}
