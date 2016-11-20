package main

import (
	"bitbot/exchanger"
)

type Client interface {
	Exchanger() string
	TradingBalances() (map[string]float64, error)
	PlaceOrder(side string, pair exchanger.Pair, price, vol float64) (map[string]interface{}, error)
}
