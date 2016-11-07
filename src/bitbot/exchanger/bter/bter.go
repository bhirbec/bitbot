package bter

import (
	"fmt"
	"strconv"

	"bitbot/exchanger"
)

const (
	APIURL        = "http://data.bter.com/api/1"
	ExchangerName = "Bter"
)

// Pairs maps standardized currency pairs to Bter pairs as used by the API.
var Pairs = map[exchanger.Pair]string{
	exchanger.BTC_USD: "BTC_USD",
	exchanger.LTC_BTC: "LTC_BTC",
	exchanger.ETH_BTC: "ETH_BTC",
	exchanger.ETC_BTC: "ETC_BTC",
}

func OrderBook(pair exchanger.Pair) (*exchanger.OrderBook, error) {
	p := Pairs[pair]
	url := fmt.Sprintf("%s/depth/%s", APIURL, p)

	var result struct {
		Result string
		Asks   [][]interface{}
		Bids   [][]interface{}
	}

	err := exchanger.FetchOrderBook(url, &result)
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

	// asks orders come in decreasing order
	asks = reverseOrders(asks)

	return exchanger.NewOrderbook(ExchangerName, bids, asks)
}

func parseOrders(rows [][]interface{}) ([]*exchanger.Order, error) {
	orders := make([]*exchanger.Order, len(rows))
	for i, row := range rows {

		price, err := parseValue(row[0])
		if err != nil {
			return nil, err
		}

		volume, err := parseValue(row[1])
		if err != nil {
			return nil, err
		}

		orders[i] = &exchanger.Order{
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

func reverseOrders(orders []*exchanger.Order) []*exchanger.Order {
	n := len(orders)
	output := make([]*exchanger.Order, n)
	for i := 0; i < n; i++ {
		output[i] = orders[n-1-i]
	}
	return output
}
