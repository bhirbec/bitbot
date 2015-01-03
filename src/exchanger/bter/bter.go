package bter

import (
	"exchanger/orderbook"
	"fmt"
	"strconv"
)

const (
	APIURL        = "http://data.bter.com/api/1"
	ExchangerName = "Bter"
)

var pairs = map[string]string{
	"BTC_USD": "BTC_USD",
	"LTC_BTC": "LTC_BTC",
}

func OrderBook(pair string) (*orderbook.OrderBook, error) {
	pair = pairs[pair]
	url := fmt.Sprintf("%s/depth/%s", APIURL, pair)

	var result struct {
		Result string
		Asks   [][]interface{}
		Bids   [][]interface{}
	}

	err := orderbook.FetchOrderBook(url, &result)
	if err != nil {
		return nil, err
	}

	// TODO: check if this test is correct
	if result.Result != "true" {
		return nil, fmt.Errorf("Bter API error. %s", err)
	}

	asks, err := parseOrders(result.Asks)
	if err != nil {
		return nil, err
	}

	bids, err := parseOrders(result.Bids)
	if err != nil {
		return nil, err
	}

	return orderbook.NewOrderbook(ExchangerName, bids, asks)
}

func parseOrders(rows [][]interface{}) ([]*orderbook.Order, error) {
	orders := make([]*orderbook.Order, len(rows))
	for i, row := range rows {

		price, err := parseValue(row[0])
		if err != nil {
			return nil, err
		}

		volume, err := parseValue(row[1])
		if err != nil {
			return nil, err
		}

		orders[i] = &orderbook.Order{
			Price:  price,
			Volume: volume,
		}
	}
	return orders, nil
}

func parseValue(v interface{}) (float64, error) {
	switch t := v.(type) {
	case string:
		value, err := strconv.ParseFloat(t, 64)
		if err != nil {
			return 0, err
		}
		return value, nil
	case float64:
		return t, nil
	default:
		return 0, fmt.Errorf("Bter API error: cannot parse %s", v)
	}
}
