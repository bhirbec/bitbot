package gemini

import (
	"fmt"
	"strconv"

	"bitbot/exchanger"
)

const (
	APIURL        = "https://api.gemini.com/v1"
	ExchangerName = "Gemini"
)

var Pairs = map[string]string{
	"eth_btc": "ethbtc",
}

func OrderBook(pair string) (*exchanger.OrderBook, error) {
	pair = Pairs[pair]
	url := fmt.Sprintf("%s/book/%s", APIURL, pair)

	var result struct {
		Asks []map[string]string
		Bids []map[string]string
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

func parseOrders(rows []map[string]string) ([]*exchanger.Order, error) {
	orders := make([]*exchanger.Order, len(rows))
	for i, row := range rows {

		price, err := strconv.ParseFloat(row["price"], 64)
		if err != nil {
			return nil, err
		}

		volume, err := strconv.ParseFloat(row["amount"], 64)
		if err != nil {
			return nil, err
		}

		timestamp, err := strconv.ParseFloat(row["timestamp"], 64)
		if err != nil {
			return nil, err
		}

		orders[i] = &exchanger.Order{
			Price:     price,
			Volume:    volume,
			Timestamp: timestamp,
		}
	}

	return orders, nil
}
