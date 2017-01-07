package poloniex

import (
	"fmt"
	"strconv"

	"bitbot/exchanger"
)

const (
	APIURL        = "https://poloniex.com/public"
	ExchangerName = "Poloniex"
)

// Pairs maps standardized currency pairs to Poloniex pairs as used by the API.
var Pairs = map[exchanger.Pair]string{
	exchanger.ETH_BTC: "BTC_ETH",
	exchanger.ETC_BTC: "BTC_ETC",
	exchanger.LTC_BTC: "BTC_LTC",
	exchanger.ZEC_BTC: "BTC_ZEC",
}

func OrderBook(pair exchanger.Pair) (*exchanger.OrderBook, error) {
	p, ok := Pairs[pair]
	if !ok {
		return nil, fmt.Errorf("Poloniex: OrderBook function doesn't not support %s", pair)
	}

	var result struct {
		Bids [][]interface{}
		Asks [][]interface{}
	}

	url := fmt.Sprintf("%s?command=returnOrderBook&currencyPair=%s&depth=10", APIURL, p)
	if err := exchanger.FetchOrderBook(url, &result); err != nil {
		return nil, err
	}

	asks, err := parseOrders(result.Asks)
	if err != nil {
		return nil, err
	}

	bids, err := parseOrders(result.Bids)
	if err != nil {
		return nil, err
	}

	return exchanger.NewOrderbook(ExchangerName, bids, asks)
}

func parseOrders(rows [][]interface{}) ([]*exchanger.Order, error) {
	orders := make([]*exchanger.Order, len(rows))
	for i, row := range rows {
		price, err := strconv.ParseFloat(row[0].(string), 64)
		if err != nil {
			return nil, err
		}

		orders[i] = &exchanger.Order{
			Price:  price,
			Volume: row[1].(float64),
		}
	}

	return orders, nil
}
