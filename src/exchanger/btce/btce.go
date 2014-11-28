package btce

import (
	"exchanger/orderbook"
	"fmt"
)

const (
	APIURL        = "https://btc-e.com/api/3"
	ExchangerName = "btce"
)

func OrderBook(pair string) (*orderbook.OrderBook, error) {
	url := fmt.Sprintf("%s/depth/%s", APIURL, pair)

	var result map[string]struct {
		Asks [][]float64
		Bids [][]float64
	}

	if err := orderbook.FetchOrderBook(url, &result); err != nil {
		return nil, err
	}

	bids := makeOrders(result[pair].Bids)
	asks := makeOrders(result[pair].Asks)
	return &orderbook.OrderBook{ExchangerName, bids, asks}, nil
}

func makeOrders(rows [][]float64) []*orderbook.Order {
	orders := make([]*orderbook.Order, len(rows))
	for i, row := range rows {
		orders[i] = &orderbook.Order{
			Price:  row[0],
			Volume: row[1],
		}
	}
	return orders
}
