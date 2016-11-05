package cex

import (
	"fmt"

	"bitbot/exchanger"
)

const (
	APIURL        = "https://cex.io/api/"
	ExchangerName = "Cex"
)

var Pairs = map[string]string{
	"btc_eur": "BTC/EUR",
	"btc_usd": "BTC/USD",
	"ltc_btc": "LTC/BTC",
	"eth_btc": "ETH/BTC",
}

func OrderBook(pair string) (*exchanger.OrderBook, error) {
	pair = Pairs[pair]
	url := fmt.Sprintf("%sorder_book/%s", APIURL, pair)

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
