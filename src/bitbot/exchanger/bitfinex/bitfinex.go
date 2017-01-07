package bitfinex

import (
	"encoding/json"
	"fmt"
	"strconv"

	"bitbot/exchanger"
)

const (
	APIURL        = "https://api.bitfinex.com/v1"
	ExchangerName = "Bitfinex"
)

// Pairs maps standardized currency pairs to Bitfinex pairs as used by the API.
var Pairs = map[exchanger.Pair]string{
	exchanger.BTC_USD: "BTCUSD",
	exchanger.LTC_USD: "LTCUSD",
	exchanger.LTC_BTC: "LTCBTC",
	exchanger.ETH_USD: "ETHUSD",
	exchanger.ETH_BTC: "ETHBTC",
	exchanger.ETC_USD: "ETCUSD",
	exchanger.ETC_BTC: "ETCBTC",
	exchanger.ZEC_BTC: "ZECBTC",
}

func OrderBook(pair exchanger.Pair) (*exchanger.OrderBook, error) {
	p, ok := Pairs[pair]
	if !ok {
		return nil, fmt.Errorf("Bitfinex: OrderBook function doesn't not support %s", pair)
	}

	var result struct {
		Asks orders
		Bids orders
	}

	url := fmt.Sprintf("%s/book/%s", APIURL, p)
	if err := exchanger.FetchOrderBook(url, &result); err != nil {
		return nil, err
	}

	return exchanger.NewOrderbook(ExchangerName, result.Bids, result.Asks)
}

type orders []*exchanger.Order

func (ko *orders) UnmarshalJSON(b []byte) error {
	rows := []map[string]string{}

	if err := json.Unmarshal(b, &rows); err != nil {
		return err
	}

	for _, row := range rows {
		price, err := strconv.ParseFloat(row["price"], 64)
		if err != nil {
			return err
		}

		volume, err := strconv.ParseFloat(row["amount"], 64)
		if err != nil {
			return err
		}

		timestamp, err := strconv.ParseFloat(row["timestamp"], 64)
		if err != nil {
			return err
		}

		*ko = append(*ko, &exchanger.Order{price, volume, timestamp})
	}
	return nil
}
