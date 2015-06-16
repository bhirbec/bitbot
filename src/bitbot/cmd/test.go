package main

import (
	"fmt"

	"bitbot/config"
	"bitbot/exchanger/hitbtc"
)

func main() {
	err := config.Load("conf.ini")
	if err != nil {
		panic(err)
	}

	// r, err := hitbtc.OrderBook("BTC_USD")
	// r, err := hitbtc.TradingBalance()
	// r, err := hitbtc.TransfertToTradingAccount(0.01, "BTC")
	// r, err := hitbtc.TransfertToMainAccount(0.01, "BTC")
	// r, err := hitbtc.PaymentAddress("BTC")
	// r, err := hitbtc.CreateAddress("BTC")
	// r, err := hitbtc.NewOrder("sell", "BTCUSD", 280, 0.01, "limit")
	r, err := hitbtc.CancelOrder("hitbtc-1434427745401321", "BTCUSD", "sell")
	// r, err := hitbtc.Withdraw(0.001, "BTC", "15zNpn6BKJ4omtNeWpPxK34y2u8y4X1hmP")
	fmt.Println(r, err)
}
