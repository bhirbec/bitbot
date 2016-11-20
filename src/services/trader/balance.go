package main

import (
	"fmt"
	"log"
	"time"

	"bitbot/exchanger"
	"bitbot/exchanger/hitbtc"
	"bitbot/exchanger/poloniex"
)

var minBalance = map[string]float64{
	"BTC": 0.0005,
	"ZEC": 0.001,
}

type transfertFunc func(*hitbtc.Client, *poloniex.Client, string, float64, string) error

var transfertFunctions = map[string]transfertFunc{
	"Hitbtc->Poloniex": moveFromHitbtcToPoloniex,
	"Poloniex->Hitbtc": moveFromPoloniexToHitbtc,
}

func rebalance(h *hitbtc.Client, p *poloniex.Client, arb *arbitrage, pair exchanger.Pair) error {
	// TODO: client should passed as a function parameter
	clients := map[string]Client{
		h.Exchanger(): h,
		p.Exchanger(): p,
	}

	balances, err := getBalances(clients)
	if err != nil {
		return err
	}

	addresses, err := getAddresses(h, p)
	if err != nil {
		return err
	}

	availableSellVol := balances[arb.sellEx.Exchanger][pair.Base]
	if availableSellVol <= minBalance[pair.Base] {
		key := arb.buyEx.Exchanger + "->" + arb.sellEx.Exchanger
		transfert, ok := transfertFunctions[key]
		if !ok {
			log.Panicf("No transfert function found for: %s", key)
		}

		printBalances(balances, pair)

		err := transfert(
			h,
			p,
			pair.Base,
			balances[arb.buyEx.Exchanger][pair.Base],
			addresses[arb.sellEx.Exchanger][pair.Base],
		)

		if err != nil {
			return err
		}
	}

	availableBuyVol := balances[arb.buyEx.Exchanger][pair.Quote] / arb.buyEx.Asks[0].Price
	if availableBuyVol <= minBalance[pair.Quote] {
		key := arb.sellEx.Exchanger + "->" + arb.buyEx.Exchanger
		transfert, ok := transfertFunctions[key]
		if !ok {
			log.Panicf("No transfert function found for: %s", key)
		}

		printBalances(balances, pair)

		err := transfert(
			h,
			p,
			pair.Quote,
			balances[arb.sellEx.Exchanger][pair.Quote],
			addresses[arb.buyEx.Exchanger][pair.Quote],
		)

		if err != nil {
			return err
		}
	}

	return nil
}

func moveFromHitbtcToPoloniex(h *hitbtc.Client, p *poloniex.Client, cur string, vol float64, address string) error {
	log.Printf("Starting transfert of %f %s from %s to %s\n", vol, cur, "Hitbtc", "Poloniex")

	result, err := h.TransfertToMainAccount(vol, cur)
	if err != nil {
		return fmt.Errorf("Cannot transfert from `%s` trading account to main account: %s\n", cur, err)
	}
	log.Printf("Transfert from trading to main account successed: %s\n", result)

	ack, err := h.Withdraw(vol, cur, address)
	if err != nil {
		return fmt.Errorf("Cannot withdraw `%s` from Hitbtc: %s\n", cur, err)
	}
	log.Printf("Transfered %f %s from %s to %s: %s\n", vol, cur, "Hitbtc", "Poloniex", ack)

	// Wait until balance reaches the min
	for {
		bal, err := p.TradingBalances()

		if err != nil {
			log.Println(err)
		} else if bal[cur] >= minBalance[cur] {
			break
		} else {
			log.Printf("Wait until %s transfer is complete (%s -> %s)\n", cur, "Hitbtc", "Poloniex")
		}

		time.Sleep(2 * time.Minute)
	}

	return nil
}

func moveFromPoloniexToHitbtc(h *hitbtc.Client, p *poloniex.Client, cur string, vol float64, address string) error {
	log.Printf("Starting transfert of %f %s from %s to %s\n", vol, cur, "Poloniex", "Hitbtc")

	ack, err := p.Withdraw(vol, cur, address)
	if err != nil {
		return fmt.Errorf("Cannot withdraw `%s` from Poloniex: %s\n", cur, err)
	}
	log.Printf("Transfered %f %s from %s to %s: %s\n", vol, cur, "Poloniex", "Hitbtc", ack)

	// Wait until we see the amout on Hitbtc main account
	for {
		bal, err := h.MainBalances()
		if err != nil {
			log.Println(err)
		} else if bal[cur] >= minBalance[cur] {
			vol = bal[cur]
			break
		} else {
			log.Printf("Wait until %s transfer is complete (%s -> %s)\n", cur, "Poloniex", "Hitbtc")
		}

		time.Sleep(2 * time.Minute)
	}

	result, err := h.TransfertToTradingAccount(vol, cur)
	if err != nil {
		return fmt.Errorf("Cannot transfert `%s` from main trading account to trading account: %s\n", cur, err)
	}
	log.Printf("Transfert from main to trading account successed: %s\n", result)

	return nil
}

func getAddresses(h *hitbtc.Client, p *poloniex.Client) (map[string]map[string]string, error) {
	out := map[string]map[string]string{
		"Hitbtc":   map[string]string{},
		"Poloniex": map[string]string{},
	}

	poloniexAddresses, err := p.DepositAddresses()
	if err != nil {
		return nil, err
	}

	for _, cur := range []string{"BTC", "ZEC"} {
		add, ok := poloniexAddresses[cur]
		if !ok {
			return nil, fmt.Errorf("Missing %s deposit address for Poloniex\n", cur)
		} else {
			out["Poloniex"][cur] = add
		}

		add, err := h.PaymentAddress(cur)
		if err != nil {
			return nil, fmt.Errorf("Cannot retrieve %s address for Hitbtc: %s\n", cur, err)
		} else if add == "" {
			return nil, fmt.Errorf("Missing %s address for Hitbtc: %s\n", cur)
		} else {
			out["Hitbtc"][cur] = add
		}
	}

	return out, nil

}

func getBalances(clients map[string]Client) (map[string]map[string]float64, error) {
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
	log.Printf("Balance: Hitbtc %s: %f, %s %f\n", pair.Base, balances["Hitbtc"][pair.Base], pair.Quote, balances["Hitbtc"][pair.Quote])
	log.Printf("Balance: Poloniex %s: %f, %s %f\n", pair.Base, balances["Poloniex"][pair.Base], pair.Quote, balances["Poloniex"][pair.Quote])

	totalBase := balances["Hitbtc"][pair.Base] + balances["Poloniex"][pair.Base]
	totalQuote := balances["Hitbtc"][pair.Quote] + balances["Poloniex"][pair.Quote]
	log.Printf("Balance: Total %s: %f, %s %f\n", pair.Base, totalBase, pair.Quote, totalQuote)
}
