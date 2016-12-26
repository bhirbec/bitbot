package main

import (
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
	// note: this is not a "market" order so it may not be filled.
	resp, err := t.Client.PlaceOrder(side, pair, price, vol)
	if err != nil {
		return nil, err
	}

	orderNumber := resp["orderNumber"].(string)
	return []string{orderNumber}, nil
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
