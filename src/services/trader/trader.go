package main

import (
	"fmt"

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

func NewHitbtcTrader(cred Credential) *HitbtcTrader {
	c := hitbtc.NewClient(cred.Key, cred.Secret)
	return &HitbtcTrader{c}
}

func (t *HitbtcTrader) PlaceOrder(side string, pair exchanger.Pair, price, vol float64) ([]string, error) {
	resp, err := t.Client.PlaceOrder(side, pair, 0, vol, "market")
	if err != nil {
		return nil, err
	}

	clientOrderId := resp["clientOrderId"].(string)
	ids := []string{clientOrderId}
	return ids, nil
}

// Poloniex trader
type PoloniexTrader struct {
	*poloniex.Client
}

func NewPoloniexTrader(cred Credential) *PoloniexTrader {
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

func NewKrakenTrader(cred Credential) *KrakenTrader {
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
