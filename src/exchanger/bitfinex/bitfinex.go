package bitfinex

import (
	"exchanger/orderbook"
	"fmt"
	"strconv"
)

const APIURL = "https://api.bitfinex.com/v1"

func OrderBook(pair string) (*orderbook.OrderBook, error) {
	url := fmt.Sprintf("%s/book/%s", APIURL, pair)

	var result struct {
		Asks []map[string]string
		Bids []map[string]string
	}

	err := orderbook.FetchOrderBook(url, &result)
	if err != nil {
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

	return &orderbook.OrderBook{bids, asks}, nil
}

func parseOrders(rows []map[string]string) ([]*orderbook.Order, error) {
	orders := make([]*orderbook.Order, len(rows))
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

		orders[i] = &orderbook.Order{
			Price:     price,
			Volume:    volume,
			Timestamp: timestamp,
		}
	}
	return orders, nil
}
