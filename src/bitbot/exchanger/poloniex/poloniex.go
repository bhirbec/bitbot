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
	"btc_eth": "BTC_ETH",
	"btc_etc": "BTC_ETC",
	"btc_ltc": "BTC_LTC",
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
