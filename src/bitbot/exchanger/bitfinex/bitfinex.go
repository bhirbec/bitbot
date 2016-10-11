package bitfinex

import (
	"encoding/json"
	"fmt"
	"strconv"

	"bitbot/exchanger/orderbook"
)

const (
	APIURL        = "https://api.bitfinex.com/v1"
	ExchangerName = "Bitfinex"
)

var Pairs = map[string]string{
	"btc_usd": "BTCUSD",
	"ltc_usd": "LTCUSD",
	"ltc_btc": "LTCBTC",
	"eth_usd": "ETHUSD",
	"eth_btc": "ETHBTC",
	"etc_usd": "ETCUSD",
	"etc_btc": "ETCBTC",
}

func OrderBook(pair string) (*orderbook.OrderBook, error) {
	pair = Pairs[pair]
	url := fmt.Sprintf("%s/book/%s", APIURL, pair)

	var result struct {
		Asks orders
		Bids orders
	}

	if err := orderbook.FetchOrderBook(url, &result); err != nil {
		return nil, err
	}

	return orderbook.NewOrderbook(ExchangerName, result.Bids, result.Asks)
}

type orders []*orderbook.Order

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

		*ko = append(*ko, &orderbook.Order{price, volume, timestamp})
	}
	return nil
}
