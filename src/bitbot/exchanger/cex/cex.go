package cex

import (
	"fmt"

	"bitbot/exchanger/orderbook"
)

const (
	APIURL        = "https://cex.io/api/"
	ExchangerName = "Cex"
)

var pairs = map[string]string{
	"BTC_USD": "BTC/USD",
	"LTC_BTC": "LTC/BTC",
}

func OrderBook(pair string) (*orderbook.OrderBook, error) {
	pair = pairs[pair]
	url := fmt.Sprintf("%sorder_book/%s", APIURL, pair)

	var result struct {
		Timestamp int64
		Asks      [][]interface{}
		Bids      [][]interface{}
	}

	if err := orderbook.FetchOrderBook(url, &result); err != nil {
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

	return orderbook.NewOrderbook(ExchangerName, bids, asks)
}

func parseOrders(rows [][]interface{}) ([]*orderbook.Order, error) {
	orders := make([]*orderbook.Order, len(rows))
	for i, row := range rows {
		orders[i] = &orderbook.Order{
			Price:  row[0].(float64),
			Volume: row[1].(float64),
		}
	}

	return orders, nil
}
