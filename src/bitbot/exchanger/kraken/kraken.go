package kraken

import (
	"fmt"
	"strconv"

	"bitbot/exchanger"
	"bitbot/httpreq"
)

const (
	// APIURL is the official Kraken API Endpoint
	APIURL = "https://api.kraken.com"

	// APIVersion is the official Kraken API Version Number
	APIVersion = "0"

	// "Kraken"
	ExchangerName = "Kraken"
)

// Pairs maps standardized currency pairs to Kraken pairs as used by the API.
var Pairs = map[exchanger.Pair]string{
	exchanger.BTC_EUR: "XXBTZEUR",
	exchanger.BTC_USD: "XXBTZUSD",
	exchanger.LTC_USD: "XLTCZUSD",
	exchanger.LTC_BTC: "XLTCXXBT",
	exchanger.ETH_USD: "XETHZUSD",
	exchanger.ETH_BTC: "XETHXXBT",
	exchanger.ETC_USD: "XETCZUSD",
	exchanger.ETC_BTC: "XETCXXBT",
	exchanger.ZEC_BTC: "XZECXXBT",
}

func OrderBook(pair exchanger.Pair) (*exchanger.OrderBook, error) {
	p := Pairs[pair]
	url := fmt.Sprintf("%s/%s/public/Depth?pair=%s", APIURL, APIVersion, p)

	var result struct {
		Error  []string
		Result map[string]struct {
			Bids [][]interface{}
			Asks [][]interface{}
		}
	}

	if err := httpreq.Get(url, nil, &result); err != nil {
		return nil, err
	}

	if len(result.Error) > 0 {
		return nil, fmt.Errorf("Kraken returned an error. %s", result.Error[0])
	}

	asks, err := parseOrders(result.Result[p].Asks)
	if err != nil {
		return nil, err
	}

	bids, err := parseOrders(result.Result[p].Bids)
	if err != nil {
		return nil, err
	}

	return exchanger.NewOrderbook(ExchangerName, bids, asks)
}

func parseOrders(rows [][]interface{}) ([]*exchanger.Order, error) {
	orders := make([]*exchanger.Order, len(rows))
	for i, row := range rows {
		price, err := strconv.ParseFloat(row[0].(string), 64)
		if err != nil {
			return nil, err
		}

		volume, err := strconv.ParseFloat(row[1].(string), 64)
		if err != nil {
			return nil, err
		}

		orders[i] = &exchanger.Order{
			Price:     price,
			Volume:    volume,
			Timestamp: row[2].(float64),
		}
	}

	return orders, nil
}
