package poloniex

import (
	"fmt"
	"strconv"

	"bitbot/exchanger/orderbook"
)

const (
	APIURL        = "https://poloniex.com/public"
	ExchangerName = "Poloniex"
)

var Pairs = map[string]string{
	"eth_btc": "BTC_ETH",
	"etc_btc": "BTC_ETC",
	"ltc_btc": "BTC_LTC",
	"zec_btc": "BTC_ZEC",
}

func OrderBook(pair string) (*orderbook.OrderBook, error) {
	pair = Pairs[pair]
	url := fmt.Sprintf("%s?command=returnOrderBook&currencyPair=%s&depth=10", APIURL, pair)

	var result struct {
		Bids [][]interface{}
		Asks [][]interface{}
	}

	if err := orderbook.FetchOrderBook(url, &result); err != nil {
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

	return orderbook.NewOrderbook(ExchangerName, bids, asks)
}

func parseOrders(rows [][]interface{}) ([]*orderbook.Order, error) {
	orders := make([]*orderbook.Order, len(rows))
	for i, row := range rows {
		price, err := strconv.ParseFloat(row[0].(string), 64)
		if err != nil {
			return nil, err
		}

		orders[i] = &orderbook.Order{
			Price:  price,
			Volume: row[1].(float64),
		}
	}

	return orders, nil
}
