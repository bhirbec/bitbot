package main

import (
	"fmt"
	"log"

	"bitbot/exchanger"
)

var minBalance = map[string]float64{
	"BTC": 0.0005,
	"ZEC": 0.001,
}

func rebalance(clients map[string]Client, arb *arbitrage, pair exchanger.Pair) error {
	balances, err := getBalances(clients, pair)
	if err != nil {
		return err
	}

	// ex: sell Poloniex ZEC: 0.000000, BTC 0.12
	vol := balances[arb.buyEx.Exchanger][pair.Base]
	if vol <= minBalance[pair.Base] {
		return transfert(
			clients[arb.buyEx.Exchanger],
			clients[arb.sellEx.Exchanger],
			pair.Base,
			balances[arb.buyEx.Exchanger][pair.Base],
		)
	}

	// ex: buy Poloniex ZEC: 0.9, BTC 0.006104
	vol = balances[arb.buyEx.Exchanger][pair.Quote]
	if vol <= minBalance[pair.Quote] {
		return transfert(
			clients[arb.sellEx.Exchanger],
			clients[arb.buyEx.Exchanger],
			pair.Quote,
			balances[arb.sellEx.Exchanger][pair.Quote],
		)
	}

	return nil
}

func transfert(org, dest Client, cur string, vol float64) error {
	log.Printf("Starting transfert of %f %s from %s to %s\n", vol, cur, org.Exchanger(), dest.Exchanger())

	var address string
	var err error

	if org.Exchanger() == "Kraken" {
		// Kraken requires to input the withdrawal addresses in the UI and to
		// give them unique name. The convention is ExchangerName + "-" + cur.
		// Example: Poloniex-ZEC
		address = fmt.Sprintf("%s-%s", dest.Exchanger(), cur)
	} else {
		address, err = dest.PaymentAddress(cur)
		if err != nil {
			return err
		}
	}

	ack, err := org.Withdraw(vol, cur, address)
	if err != nil {
		return fmt.Errorf("Cannot withdraw `%s` from %s: %s\n", cur, err, org.Exchanger())
	} else {
		log.Printf("Transfer registered: %s\n", ack)
	}

	return dest.WaitBalance(cur)
}

func getBalances(clients map[string]Client, pair exchanger.Pair) (map[string]map[string]float64, error) {
	out := map[string]map[string]float64{}

	for _, c := range clients {
		b, err := c.TradingBalances()
		if err != nil {
			return nil, err
		}
		out[c.Exchanger()] = b
	}

	return out, nil
}

func printBalances(balances map[string]map[string]float64, pair exchanger.Pair) {
	var totalBase float64
	var totalQuote float64

	for ex, bal := range balances {
		totalBase += bal[pair.Base]
		totalQuote += bal[pair.Quote]
		log.Printf("Balance: %s %s: %f, %s %f\n", ex, pair.Base, bal[pair.Base], pair.Quote, bal[pair.Quote])
	}

	log.Printf("Balance: Total %s: %f, %s %f\n", pair.Base, totalBase, pair.Quote, totalQuote)
}
