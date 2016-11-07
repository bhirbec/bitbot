package cex

import (
	"fmt"

	"bitbot/exchanger"
)

const (
	APIURL        = "https://cex.io/api/"
	ExchangerName = "Cex"
)

// Pairs maps standardized currency pairs to Cex pairs as used by the API.
var Pairs = map[exchanger.Pair]string{
	exchanger.BTC_EUR: "BTC/EUR",
	exchanger.BTC_USD: "BTC/USD",
	exchanger.LTC_BTC: "LTC/BTC",
	exchanger.ETH_BTC: "ETH/BTC",
}

func OrderBook(pair exchanger.Pair) (*exchanger.OrderBook, error) {
	p := Pairs[pair]
	url := fmt.Sprintf("%sorder_book/%s", APIURL, p)

	var result struct {
		Timestamp int64
		Asks      [][]interface{}
		Bids      [][]interface{}
	}

	if err := exchanger.FetchOrderBook(url, &result); err != nil {
		return nil, err
	}

	bids, err := parseOrders(result.Bids)
	if err != nil {
		return nil, err
	}

	asks, err := parseOrders(result.Asks)
	if err != nil {
		return nil, err
	}

	return exchanger.NewOrderbook(ExchangerName, bids, asks)
}

func parseOrders(rows [][]interface{}) ([]*exchanger.Order, error) {
	orders := make([]*exchanger.Order, len(rows))
	for i, row := range rows {
		orders[i] = &exchanger.Order{
			Price:  row[0].(float64),
			Volume: row[1].(float64),
		}
	}

	return orders, nil
}
