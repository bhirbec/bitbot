package main

import (
	"fmt"
	"log"
	"time"

	"bitbot/exchanger"
	"bitbot/exchanger/hitbtc"
	"bitbot/exchanger/poloniex"
)

type Client interface {
	Exchanger() string
	TradingBalances() (map[string]float64, error)
	PlaceOrder(side string, pair exchanger.Pair, price, vol float64) (map[string]interface{}, error)
	Withdraw(vol float64, cur, address string) (string, error)
	WaitBalance(cur string) error
	PaymentAddress(cur string) (string, error)
}

// Hitbtc
type HitbtcClient struct {
	*hitbtc.Client
}

func NewHitbtcClient(cred credential) *HitbtcClient {
	c := hitbtc.NewClient(cred.Key, cred.Secret)
	return &HitbtcClient{c}
}

func (c *HitbtcClient) Withdraw(vol float64, cur, address string) (string, error) {
	result, err := c.Client.TransfertToMainAccount(vol, cur)
	if err != nil {
		return "", fmt.Errorf("Hitbtc: cannot transfert from `%s` trading account to main account: %s\n", cur, err)
	} else {
		log.Printf("Hitbtc: transfert from trading to main account successed: %s\n", result)
	}

	_, err = c.Client.Withdraw(vol, cur, address)
	return "ok", err
}

func (c *HitbtcClient) WaitBalance(cur string) error {
	var vol float64

	for {
		bal, err := c.Client.MainBalances()
		if err != nil {
			log.Println(err)
		} else if bal[cur] >= minBalance[cur] {
			vol = bal[cur]
			break
		} else {
			log.Printf("Hitbtc: Wait until %s transfer is complete\n", cur)
		}

		time.Sleep(2 * time.Minute)
	}

	result, err := c.Client.TransfertToTradingAccount(vol, cur)
	if err != nil {
		return fmt.Errorf("Cannot transfert `%s` from main trading account to trading account: %s\n", cur, err)
	}

	log.Printf("Transfert from main to trading account successed: %s\n", result)
	return nil
}

// Poloniex
type PoloniexClient struct {
	*poloniex.Client
	addresses map[string]string
}

func NewPoloniexClient(cred credential) *PoloniexClient {
	c := poloniex.NewClient(cred.Key, cred.Secret)
	return &PoloniexClient{c, map[string]string{}}
}

func (c *PoloniexClient) WaitBalance(cur string) error {
	for {
		bal, err := c.Client.TradingBalances()

		if err != nil {
			return err
		} else if bal[cur] >= minBalance[cur] {
			break
		} else {
			log.Printf("Poloniex: Wait until %s transfer is complete\n", cur)
		}

		time.Sleep(2 * time.Minute)
	}

	return nil
}

func (c *PoloniexClient) PaymentAddress(cur string) (string, error) {
	// we first load the addresses and then cache them
	if len(c.addresses) == 0 {
		addresses, err := c.Client.DepositAddresses()
		if err != nil {
			return "", fmt.Errorf("Poloniex: cannot retrieve address for %s: %s\n", cur, err)
		} else {
			c.addresses = addresses
		}
	}

	address, ok := c.addresses[cur]
	if !ok {
		return "", fmt.Errorf("Poloniex: missing %s address\n", cur)
	} else {
		return address, nil
	}
}
