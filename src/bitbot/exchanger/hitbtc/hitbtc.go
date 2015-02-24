package hitbtc

import (
	"fmt"
	"strconv"

	"bitbot/exchanger/orderbook"
)

const (
	APIURL        = "https://api.hitbtc.com/api/1"
	ExchangerName = "Hitbtc"
)

var pairs = map[string]string{
	"BTC_USD": "BTCUSD",
	"LTC_BTC": "LTCBTC",
}

func OrderBook(pair string) (*orderbook.OrderBook, error) {
	pair = pairs[pair]
	url := fmt.Sprintf("%s/public/%s/orderbook", APIURL, pair)

	var result struct {
		Asks [][]string
		Bids [][]string
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

func parseOrders(rows [][]string) ([]*orderbook.Order, error) {
	orders := make([]*orderbook.Order, len(rows))
	for i, row := range rows {
		price, err := strconv.ParseFloat(row[0], 64)
		if err != nil {
			return nil, err
		}

		volume, err := strconv.ParseFloat(row[1], 64)
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