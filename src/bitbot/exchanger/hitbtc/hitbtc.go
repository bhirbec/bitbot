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

// Pairs maps standardized currency pairs to Hitbtc pairs as used by the API.
var Pairs = map[exchanger.Pair]string{
	exchanger.BTC_EUR: "BTCEUR",
	exchanger.BTC_USD: "BTCUSD",
	exchanger.LTC_BTC: "LTCBTC",
	exchanger.LTC_USD: "LTCUSD",
	exchanger.ETH_BTC: "ETHBTC",
	exchanger.ZEC_BTC: "ZECBTC",
}

func OrderBook(pair exchanger.Pair) (*exchanger.OrderBook, error) {
	p := Pairs[pair]
	url := fmt.Sprintf("%s/api/1/public/%s/orderbook", host, p)

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
