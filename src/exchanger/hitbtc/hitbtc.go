package hitbtc

import (
	"exchanger/util"
	"fmt"
	"strconv"
)

const APIURL = "https://api.hitbtc.com/api/1"

type OrderBook struct {
	Bids []*Order
	Asks []*Order
}

type Order struct {
	Price     float64
	Volume    float64
	Timestamp float64
}

func FetchOrderBook(pair string) (*OrderBook, error) {
	url := fmt.Sprintf("%s/public/%s/orderbook", APIURL, pair)

	var result struct {
		Asks [][]string
		Bids [][]string
	}

	if err := util.FetchOrderBook(url, &result); err != nil {
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

	return &OrderBook{bids, asks}, nil
}

func parseOrders(rows [][]string) ([]*Order, error) {
	orders := make([]*Order, len(rows))
	for i, row := range rows {
		price, err := strconv.ParseFloat(row[0], 64)
		if err != nil {
			return nil, err
		}

		volume, err := strconv.ParseFloat(row[1], 64)
		if err != nil {
			return nil, err
		}

		orders[i] = &Order{
			Price:  price,
			Volume: volume,
		}
	}

	return orders, nil
}
