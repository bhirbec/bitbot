package hitbtc

import (
	"fmt"
	"strconv"

	"bitbot/exchanger"
	"bitbot/httpreq"
)

const (
	host          = "https://api.hitbtc.com"
	ExchangerName = "Hitbtc"
)

var Pairs = map[string]string{
	"btc_eur": "BTCEUR",
	"btc_usd": "BTCUSD",
	"ltc_btc": "LTCBTC",
	"ltc_usd": "LTCUSD",
	"eth_btc": "ETHBTC",
	"zec_btc": "ZECBTC",
}

func OrderBook(pair string) (*exchanger.OrderBook, error) {
	pair = Pairs[pair]
	url := fmt.Sprintf("%s/api/1/public/%s/orderbook", host, pair)

	var result struct {
		Asks [][]string
		Bids [][]string
	}

	if err := httpreq.Get(url, nil, &result); err != nil {
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

func parseOrders(rows [][]string) ([]*exchanger.Order, error) {
	orders := make([]*exchanger.Order, len(rows))
	for i, row := range rows {
		price, err := strconv.ParseFloat(row[0], 64)
		if err != nil {
			return nil, err
		}

		volume, err := strconv.ParseFloat(row[1], 64)
		if err != nil {
			return nil, err
		}

		orders[i] = &exchanger.Order{
			Price:  price,
			Volume: volume,
		}
	}

	return orders, nil
}
