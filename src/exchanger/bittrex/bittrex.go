package bittrex

import (
	"exchanger/orderbook"
	"fmt"
)

const (
	APIURL        = "https://bittrex.com/api/v1.1/public"
	ExchangerName = "Bittrex"
)

type order struct {
	Quantity, Rate float64
}

func OrderBook(pair string) (*orderbook.OrderBook, error) {
	url := fmt.Sprintf("%s/getorderbook?market=%s&type=both", APIURL, pair)

	var result struct {
		Success bool
		Message string
		Result  struct {
			Buy  []*order
			Sell []*order
		}
	}

	if err := orderbook.FetchOrderBook(url, &result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf("Bittrex returned an error. %s", result.Message)
	}

	bids := makeOrders(result.Result.Buy)
	asks := makeOrders(result.Result.Sell)
	return orderbook.NewOrderbook(ExchangerName, bids, asks)
}

func makeOrders(rows []*order) []*orderbook.Order {
	orders := make([]*orderbook.Order, len(rows))
	for i, row := range rows {
		orders[i] = &orderbook.Order{
			Price:  row.Rate,
			Volume: row.Quantity,
		}
	}

	return orders
}
