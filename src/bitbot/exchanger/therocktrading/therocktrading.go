package therocktrading

import (
	"fmt"

	"bitbot/exchanger"
)

const (
	APIURL        = "https://api.therocktrading.com/v1"
	ExchangerName = "The Rock Trading"
)

var Pairs = map[string]string{
	"ltc_btc": "LTCBTC",
	"eth_btc": "ETHBTC",
}

func OrderBook(pair string) (*exchanger.OrderBook, error) {
	pair = Pairs[pair]
	url := fmt.Sprintf("%s/funds/%s/orderbook", APIURL, pair)

	var result struct {
		Asks []map[string]float64
		Bids []map[string]float64
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

func parseOrders(rows []map[string]float64) ([]*exchanger.Order, error) {
	orders := make([]*exchanger.Order, len(rows))
	for i, row := range rows {
		orders[i] = &exchanger.Order{
			Price:  row["price"],
			Volume: row["amount"],
		}
	}

	return orders, nil
}
