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

type Trader interface {
	Exchanger() string
	TradingBalances() (map[string]float64, error)
	PlaceOrder(side string, pair exchanger.Pair, price, vol float64) (map[string]interface{}, error)
	Withdraw(vol float64, cur, address string) (string, error)
	WaitBalance(cur string, amount float64) error
	PaymentAddress(cur string) (string, error)
}

var balanceWaintingDuration = 2 * time.Minute

// ***************** Hitbtc *****************

type HitbtcTrader struct {
	*hitbtc.Client
}

func NewHitbtcTrader(cred credential) *HitbtcTrader {
	c := hitbtc.NewClient(cred.Key, cred.Secret)
	return &HitbtcTrader{c}
}

func (t *HitbtcTrader) Exchanger() string {
	return hitbtc.ExchangerName
}

func (t *HitbtcTrader) TradingBalances() (map[string]float64, error) {
	return t.Client.TradingBalances()
}

func (t *HitbtcTrader) PlaceOrder(side string, pair exchanger.Pair, price, vol float64) (map[string]interface{}, error) {
	return t.Client.PlaceOrder(side, pair, 0, vol, "market")
}

func (t *HitbtcTrader) Withdraw(vol float64, cur, address string) (string, error) {
	result, err := t.Client.TransfertToMainAccount(vol, cur)
	if err != nil {
		return "", fmt.Errorf("Hitbtc: cannot transfert from `%s` trading account to main account: %s", cur, err)
	} else {
		log.Printf("Hitbtc: transfert from trading to main account successed: %s\n", result)
	}

	_, err = t.Client.Withdraw(vol, cur, address)
	return "ok", err
}

func (t *HitbtcTrader) WaitBalance(cur string, amount float64) error {
	var vol float64

	for {
		bal, err := t.Client.MainBalances()
		if err != nil {
			log.Println(err)
		} else if bal[cur] > amount {
			vol = bal[cur]
			break
		} else {
			log.Printf("Hitbtc: Wait until %s transfer is complete\n", cur)
		}

		time.Sleep(balanceWaintingDuration)
	}

	result, err := t.Client.TransfertToTradingAccount(vol, cur)
	if err != nil {
		return fmt.Errorf("Hitbtc: Cannot transfert `%s` from main trading account to trading account: %s", cur, err)
	}

	log.Printf("Hitbtc: transfert from main to trading account successed: %s\n", result)
	return nil
}

// ***************** Poloniex *****************

type PoloniexTrader struct {
	*poloniex.Client
	addresses map[string]string
}

func NewPoloniexTrader(cred credential) *PoloniexTrader {
	c := poloniex.NewClient(cred.Key, cred.Secret)
	return &PoloniexTrader{c, map[string]string{}}
}

func (t *PoloniexTrader) Exchanger() string {
	return poloniex.ExchangerName
}

func (t *PoloniexTrader) TradingBalances() (map[string]float64, error) {
	return t.Client.TradingBalances()
}

func (t *PoloniexTrader) WaitBalance(cur string, amount float64) error {
	for {
		bal, err := t.Client.TradingBalances()

		if err != nil {
			return err
		} else if bal[cur] >= amount {
			break
		} else {
			log.Printf("Poloniex: Wait until %s transfer is complete\n", cur)
		}

		time.Sleep(balanceWaintingDuration)
	}

	return nil
}

func (t *PoloniexTrader) PaymentAddress(cur string) (string, error) {
	// we first load the addresses and then cache them
	if len(t.addresses) == 0 {
		addresses, err := t.Client.DepositAddresses()
		if err != nil {
			return "", fmt.Errorf("Poloniex: cannot retrieve address for %s: %s", cur, err)
		} else {
			t.addresses = addresses
		}
	}

	address, ok := t.addresses[cur]
	if !ok {
		return "", fmt.Errorf("Poloniex: missing %s address", cur)
	} else {
		return address, nil
	}
}

// ***************** Kraken *****************

type KrakenTrader struct {
	*kraken.Client
}

func NewKrakenTrader(cred credential) *KrakenTrader {
	c := kraken.NewClient(cred.Key, cred.Secret)
	return &KrakenTrader{c}
}

func (t *KrakenTrader) Exchanger() string {
	return kraken.ExchangerName
}

func (t *KrakenTrader) TradingBalances() (map[string]float64, error) {
	return t.Client.AccountBalance()
}

func (t *KrakenTrader) PlaceOrder(side string, pair exchanger.Pair, price, vol float64) (map[string]interface{}, error) {
	return t.Client.AddOrder(side, pair, price, vol)
}

// Withdraw withdraws some fund from the registered account.
//
// Withdraw status:
// - initiated: the withdraw was received by Kraken
// - on hold: email confirmation was sent and transaction is waiting for approval
// - pending: confirmation link was clicked
// - sending: sending transaction
// - success:
//
// Fees:
// - BTC: ฿0.00050
// - ZEC: ⓩ0.00010
func (t *KrakenTrader) Withdraw(vol float64, cur, account string) (string, error) {
	// After some testing it appears that the currencies doesn't need to be translated to
	// kraken symbol. It works with ZEC, XZEC, BTC and XBT.
	data := map[string]string{
		"asset":  cur,
		"key":    account,
		"amount": fmt.Sprint(vol),
	}

	resp := map[string]string{}
	err := t.Client.Query("Withdraw", data, &resp)
	if err != nil {
		return "", fmt.Errorf("Kraken: %s withdraw failed - %s", cur, err)
	}

	// TODO: need to fetch gmail to get confirmation code and then
	// curl -X POST --data "code=4yR4ope7X32I1fTrVpSm1xuou69VrYpCRp4KjTYJI9XBDT0D" https://www.kraken.com/withdrawal-approve
	return resp["refid"], nil
}

func (t *KrakenTrader) WaitBalance(cur string, amount float64) error {
	for {
		bal, err := t.Client.TradeBalance(cur)

		if err != nil {
			return err
		} else if bal >= amount {
			break
		} else {
			log.Printf("Kraken: Wait until %s transfer is complete\n", cur)
		}

		time.Sleep(balanceWaintingDuration)
	}

	return nil
}

// PaymentAddress retrieve the first payment address for the given currency.
func (t *KrakenTrader) PaymentAddress(cur string) (string, error) {
	// Apparently kraken does the translation from "BTC" to "XBT"
	data := map[string]string{"asset": cur}
	resp := []map[string]interface{}{}

	err := t.Client.Query("DepositMethods", data, &resp)
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
	err = t.Client.Query("DepositAddresses", data, &resp)
	if err != nil {
		return "", fmt.Errorf("Kraken: call to DepositAddresses failed - %s", err)
	} else if len(resp) == 0 {
		return "", fmt.Errorf("Kraken: missing address for currency %s", cur)
	}

	return resp[0]["address"].(string), nil
}
