package main

import (
	"fmt"
	"log"
	"time"

	"bitbot/exchanger"
	"bitbot/exchanger/hitbtc"
	"bitbot/exchanger/kraken"
	"bitbot/exchanger/poloniex"
)

type Client interface {
	Exchanger() string
	TradingBalances(currencies ...string) (map[string]float64, error)
	PlaceOrder(side string, pair exchanger.Pair, price, vol float64) (map[string]interface{}, error)
	Withdraw(vol float64, cur, address string) (string, error)
	WaitBalance(cur string) error
	PaymentAddress(cur string) (string, error)
}

// ***************** Hitbtc *****************

type HitbtcClient struct {
	*hitbtc.Client
}

func NewHitbtcClient(cred credential) *HitbtcClient {
	c := hitbtc.NewClient(cred.Key, cred.Secret)
	return &HitbtcClient{c}
}

func (c *HitbtcClient) Exchanger() string {
	return hitbtc.ExchangerName
}

func (c *HitbtcClient) TradingBalances(currencies ...string) (map[string]float64, error) {
	return c.Client.TradingBalances()
}

func (c *HitbtcClient) PlaceOrder(side string, pair exchanger.Pair, price, vol float64) (map[string]interface{}, error) {
	return c.Client.PlaceOrder(side, pair, price, vol, "market")
}

func (c *HitbtcClient) Withdraw(vol float64, cur, address string) (string, error) {
	result, err := c.Client.TransfertToMainAccount(vol, cur)
	if err != nil {
		return "", fmt.Errorf("Hitbtc: cannot transfert from `%s` trading account to main account: %s", cur, err)
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
		return fmt.Errorf("Hitbtc: Cannot transfert `%s` from main trading account to trading account: %s", cur, err)
	}

	log.Printf("Hitbtc: transfert from main to trading account successed: %s\n", result)
	return nil
}

// ***************** Poloniex *****************

type PoloniexClient struct {
	*poloniex.Client
	addresses map[string]string
}

func NewPoloniexClient(cred credential) *PoloniexClient {
	c := poloniex.NewClient(cred.Key, cred.Secret)
	return &PoloniexClient{c, map[string]string{}}
}

func (c *PoloniexClient) Exchanger() string {
	return poloniex.ExchangerName
}

func (c *PoloniexClient) TradingBalances(currencies ...string) (map[string]float64, error) {
	return c.Client.TradingBalances()
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
			return "", fmt.Errorf("Poloniex: cannot retrieve address for %s: %s", cur, err)
		} else {
			c.addresses = addresses
		}
	}

	address, ok := c.addresses[cur]
	if !ok {
		return "", fmt.Errorf("Poloniex: missing %s address", cur)
	} else {
		return address, nil
	}
}

// ***************** Kraken *****************

type KrakenClient struct {
	*kraken.Client
}

func NewKrakenClient(cred credential) *KrakenClient {
	c := kraken.NewClient(cred.Key, cred.Secret)
	return &KrakenClient{c}
}

func (c *KrakenClient) Exchanger() string {
	return kraken.ExchangerName
}

func (c *KrakenClient) TradingBalances(currencies ...string) (map[string]float64, error) {
	out := map[string]float64{}

	for _, cur := range currencies {
		bal, err := c.Client.TradeBalance(cur)
		if err != nil {
			return nil, fmt.Errorf("Kraken: missing balance for currency %s", cur)
		}

		out[cur] = bal
	}

	return out, nil
}

func (c *KrakenClient) PlaceOrder(side string, pair exchanger.Pair, price, vol float64) (map[string]interface{}, error) {
	return c.Client.AddOrder(side, pair, price, vol)
}

func (c *KrakenClient) Withdraw(vol float64, cur, key string) (string, error) {
	cur, ok := kraken.Currencies[cur]
	if !ok {
		return "", fmt.Errorf("Kraken: currency not supported %s", cur)
	}

	data := map[string]string{
		"asset":  cur,
		"key":    key,
		"amount": fmt.Sprint(vol),
	}

	resp := map[string]string{}
	err := c.Client.Query("Withdraw", data, &resp)
	if err != nil {
		return "", fmt.Errorf("Kraken: call to Withdraw failed - %s", err)
	}

	return resp["refid"], nil
}

func (c *KrakenClient) WaitBalance(cur string) error {
	for {
		bal, err := c.Client.TradeBalance(cur)

		if err != nil {
			return err
		} else if bal >= minBalance[cur] {
			break
		} else {
			log.Printf("Kraken: Wait until %s transfer is complete\n", cur)
		}

		time.Sleep(2 * time.Minute)
	}

	return nil
}

// PaymentAddress retrieve the first payment address for the given currency.
func (c *KrakenClient) PaymentAddress(cur string) (string, error) {
	// Apparently kraken does the translation from "BTC" to "XBT"
	data := map[string]string{"asset": cur}
	resp := []map[string]interface{}{}

	err := c.Client.Query("DepositMethods", data, &resp)
	if err != nil {
		return "", fmt.Errorf("Kraken: call to DepositMethods failed - %s", err)
	} else if len(resp) == 0 {
		return "", fmt.Errorf("Kraken: call to DepositMethods failed - empty list")
	}

	data = map[string]string{
		"asset":  cur,
		"method": resp[0]["method"].(string),
	}

	resp = []map[string]interface{}{}
	err = c.Client.Query("DepositAddresses", data, &resp)
	if err != nil {
		return "", fmt.Errorf("Kraken: call to DepositAddresses failed - %s", err)
	} else if len(resp) == 0 {
		return "", fmt.Errorf("Kraken: missing address for currency %s", cur)
	}

	return resp[0]["address"].(string), nil
}
