package btce

import (
	"fmt"

	"bitbot/exchanger"
)

const (
	APIURL        = "https://btc-e.com/api/3"
	ExchangerName = "Btce"
)

// Pairs maps standardized currency pairs to Btce pairs as used by the API.
var Pairs = map[exchanger.Pair]string{
	exchanger.BTC_EUR: "btc_eur",
	exchanger.BTC_USD: "btc_usd",
	exchanger.LTC_BTC: "ltc_btc",
	exchanger.ETH_BTC: "eth_btc",
}

func OrderBook(pair exchanger.Pair) (*exchanger.OrderBook, error) {
	p, ok := Pairs[pair]
	if !ok {
		return nil, fmt.Errorf("Btce: OrderBook function doesn't not support %s", pair)
	}

	var result map[string]struct {
		Asks [][]float64
		Bids [][]float64
	}

	url := fmt.Sprintf("%s/depth/%s", APIURL, p)
	if err := exchanger.FetchOrderBook(url, &result); err != nil {
		return nil, err
	}

	bids := makeOrders(result[p].Bids)
	asks := makeOrders(result[p].Asks)
	return exchanger.NewOrderbook(ExchangerName, bids, asks)
}

func makeOrders(rows [][]float64) []*exchanger.Order {
	orders := make([]*exchanger.Order, len(rows))
	for i, row := range rows {
		orders[i] = &exchanger.Order{
			Price:  row[0],
			Volume: row[1],
		}
	}
	return orders
}
