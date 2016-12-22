package main

import (
	"bitbot/exchanger"
	"bitbot/exchanger/hitbtc"
	"bitbot/exchanger/kraken"
	"bitbot/exchanger/poloniex"
)

type Trader interface {
	PlaceOrder(side string, pair exchanger.Pair, price, vol float64) (map[string]interface{}, error)
}

// ***************** Hitbtc *****************

type HitbtcTrader struct {
	*hitbtc.Client
}

func NewHitbtcTrader(cred credential) *HitbtcTrader {
	c := hitbtc.NewClient(cred.Key, cred.Secret)
	return &HitbtcTrader{c}
}

func (t *HitbtcTrader) PlaceOrder(side string, pair exchanger.Pair, price, vol float64) (map[string]interface{}, error) {
	return t.Client.PlaceOrder(side, pair, 0, vol, "market")
}

// ***************** Poloniex *****************

type PoloniexTrader struct {
	*poloniex.Client
}

func NewPoloniexTrader(cred credential) *PoloniexTrader {
	c := poloniex.NewClient(cred.Key, cred.Secret)
	return &PoloniexTrader{c}
}

// ***************** Kraken *****************

type KrakenTrader struct {
	*kraken.Client
}

func NewKrakenTrader(cred credential) *KrakenTrader {
	c := kraken.NewClient(cred.Key, cred.Secret)
	return &KrakenTrader{c}
}

func (t *KrakenTrader) PlaceOrder(side string, pair exchanger.Pair, price, vol float64) (map[string]interface{}, error) {
	return t.Client.AddOrder(side, pair, price, vol)
}
