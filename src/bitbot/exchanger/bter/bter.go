package bter

import (
	"fmt"
	"strconv"

	"bitbot/exchanger/orderbook"
)

const (
	APIURL        = "http://data.bter.com/api/1"
	ExchangerName = "Bter"
)

var Pairs = map[string]string{
	"btc_usd": "BTC_USD",
	"ltc_btc": "LTC_BTC",
	"eth_btc": "ETH_BTC",
	"etx_btc": "ETC_BTC",
}

func OrderBook(pair string) (*orderbook.OrderBook, error) {
	pair = Pairs[pair]
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

	// asks orders come in decreasing order
	asks = reverse(asks)

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

func reverse(orders []*orderbook.Order) []*orderbook.Order {
	n := len(orders)
	output := make([]*orderbook.Order, n)
	for i := 0; i < n; i++ {
		output[i] = orders[n-1-i]
	}
	return output
}
