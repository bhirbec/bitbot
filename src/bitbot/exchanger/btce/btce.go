package btce

import (
	"fmt"

	"bitbot/exchanger/orderbook"
)

const (
	APIURL        = "https://btc-e.com/api/3"
	ExchangerName = "Btce"
)

var Pairs = map[string]string{
	"btc_eur": "btc_eur",
	"btc_usd": "btc_usd",
	"ltc_btc": "ltc_btc",
}

func OrderBook(pair string) (*orderbook.OrderBook, error) {
	pair = Pairs[pair]
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
	return orderbook.NewOrderbook(ExchangerName, bids, asks)
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
