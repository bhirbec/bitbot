package kraken

import (
	"fmt"
	"strconv"

	"bitbot/exchanger/orderbook"
)

const (
	APIURL        = "https://api.kraken.com/0"
	ExchangerName = "Kraken"
)

var Pairs = map[string]string{
	"btc_eur": "XXBTZEUR",
	"btc_usd": "XXBTZUSD",
	"ltc_usd": "XLTCZUSD",
	"ltc_btc": "XLTCXXBT",
	// "eth_usd": "XETHZUSD",
	"eth_btc": "XETHXXBT",
	// "etc_usd": "XETCZUSD",
	"etc_btc": "XETCXXBT",
}

func OrderBook(pair string) (*orderbook.OrderBook, error) {
	pair = Pairs[pair]
	url := fmt.Sprintf("%s/public/Depth?pair=%s", APIURL, pair)

	var result struct {
		Error  []string
		Result map[string]struct {
			Bids [][]interface{}
			Asks [][]interface{}
		}
	}

	if err := orderbook.FetchOrderBook(url, &result); err != nil {
		return nil, err
	}

	if len(result.Error) > 0 {
		return nil, fmt.Errorf("Kraken returned an error. %s", result.Error[0])
	}

	asks, err := parseOrders(result.Result[pair].Asks)
	if err != nil {
		return nil, err
	}

	bids, err := parseOrders(result.Result[pair].Bids)
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

		volume, err := strconv.ParseFloat(row[1].(string), 64)
		if err != nil {
			return nil, err
		}

		orders[i] = &orderbook.Order{
			Price:     price,
			Volume:    volume,
			Timestamp: row[2].(float64),
		}
	}

	return orders, nil
}
