package main

import (
	"fmt"
	"strconv"
	"time"

	"bitbot/exchanger"
	"bitbot/exchanger/hitbtc"
	"bitbot/exchanger/kraken"
	"bitbot/exchanger/poloniex"
)

type Trader interface {
	PlaceOrder(side string, pair exchanger.Pair, price, vol float64) ([]string, error)
}

// Hitbtc trader
type HitbtcTrader struct {
	*hitbtc.Client
}

func NewHitbtcTrader(cred credential) *HitbtcTrader {
	c := hitbtc.NewClient(cred.Key, cred.Secret)
	return &HitbtcTrader{c}
}

func (t *HitbtcTrader) PlaceOrder(side string, pair exchanger.Pair, price, vol float64) ([]string, error) {
	resp1, err := t.Client.PlaceOrder(side, pair, 0, vol, "market")
	if err != nil {
		return nil, err
	}

	// TradesByOrder sometime returns an "Unknown order" error. It seems Hitbtc database
	// needs sometime to propagate the information...
	time.Sleep(5 * time.Second)

	clientOrderId := resp1["clientOrderId"].(string)
	resp2, err := t.Client.TradesByOrder(clientOrderId)
	if err != nil {
		return nil, err
	}

	ids := []string{}
	for _, m := range resp2 {
		i := int64(m["tradeId"].(float64))
		ids = append(ids, strconv.FormatInt(i, 10))
	}

	return ids, nil
}

// Poloniex trader
type PoloniexTrader struct {
	*poloniex.Client
}

func NewPoloniexTrader(cred credential) *PoloniexTrader {
	c := poloniex.NewClient(cred.Key, cred.Secret)
	return &PoloniexTrader{c}
}

func (t *PoloniexTrader) PlaceOrder(side string, pair exchanger.Pair, price, vol float64) ([]string, error) {
	resp, err := t.Client.PlaceOrder(side, pair, price, vol)
	if err != nil {
		return nil, err
	}

	// note: this is not a "market" order so it may not be filled.
	trades := resp["resultingTrades"].([]interface{})
	if len(trades) == 0 {
		// TODO: cancel order?
		return nil, fmt.Errorf("Kraken: order has been filled (0 resulting trade)")
	}

	ids := []string{}
	for _, m := range trades {
		item := m.(map[string]interface{})
		ids = append(ids, item["tradeID"].(string))
	}

	return ids, nil
}

// Kraken Trader
type KrakenTrader struct {
	*kraken.Client
}

func NewKrakenTrader(cred credential) *KrakenTrader {
	c := kraken.NewClient(cred.Key, cred.Secret)
	return &KrakenTrader{c}
}

func (t *KrakenTrader) PlaceOrder(side string, pair exchanger.Pair, price, vol float64) ([]string, error) {
	resp, err := t.Client.AddOrder(side, pair, price, vol, "market")
	if err != nil {
		return []string{}, err
	}

	tradeIds := resp["txid"].([]interface{})

	ids := []string{}
	for _, id := range tradeIds {
		ids = append(ids, id.(string))
	}

	return ids, nil
}
