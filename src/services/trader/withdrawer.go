package main

import (
	"fmt"
	"log"
	"time"

	"bitbot/exchanger/hitbtc"
	"bitbot/exchanger/kraken"
	"bitbot/exchanger/poloniex"
)

type Withdrawer interface {
	Exchanger() string
	TradingBalances() (map[string]float64, error)
	Withdraw(vol float64, cur, address string) (string, error)
	WaitBalance(cur string, amount float64) error
	PaymentAddress(cur string) (string, error)
}

var balanceWaintingDuration = 2 * time.Minute

// ***************** Hitbtc *****************

type HitbtcWithdrawer struct {
	*hitbtc.Client
}

func NewHitbtcWithdrawer(cred Credential) *HitbtcWithdrawer {
	c := hitbtc.NewClient(cred.Key, cred.Secret)
	return &HitbtcWithdrawer{c}
}

func (w *HitbtcWithdrawer) Exchanger() string {
	return hitbtc.ExchangerName
}

func (w *HitbtcWithdrawer) TradingBalances() (map[string]float64, error) {
	return w.Client.TradingBalances()
}

func (w *HitbtcWithdrawer) Withdraw(vol float64, cur, address string) (string, error) {
	result, err := w.Client.TransfertToMainAccount(vol, cur)
	if err != nil {
		return "", fmt.Errorf("Hitbtc: cannot transfert from `%s` trading account to main account: %s", cur, err)
	} else {
		log.Printf("Hitbtc: transfert from trading to main account successed: %s\n", result)
	}

	_, err = w.Client.Withdraw(vol, cur, address)
	return "ok", err
}

func (w *HitbtcWithdrawer) WaitBalance(cur string, amount float64) error {
	var vol float64

	for {
		bal, err := w.Client.MainBalances()
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

	result, err := w.Client.TransfertToTradingAccount(vol, cur)
	if err != nil {
		return fmt.Errorf("Hitbtc: Cannot transfert `%s` from main trading account to trading account: %s", cur, err)
	}

	log.Printf("Hitbtc: transfert from main to trading account successed: %s\n", result)
	return nil
}

// ***************** Poloniex *****************

type PoloniexWithdrawer struct {
	*poloniex.Client
	addresses map[string]string
}

func NewPoloniexWithdrawer(cred Credential) *PoloniexWithdrawer {
	c := poloniex.NewClient(cred.Key, cred.Secret)
	return &PoloniexWithdrawer{c, map[string]string{}}
}

func (w *PoloniexWithdrawer) Exchanger() string {
	return poloniex.ExchangerName
}

func (w *PoloniexWithdrawer) TradingBalances() (map[string]float64, error) {
	return w.Client.TradingBalances()
}

func (w *PoloniexWithdrawer) WaitBalance(cur string, amount float64) error {
	for {
		bal, err := w.Client.TradingBalances()

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

func (w *PoloniexWithdrawer) PaymentAddress(cur string) (string, error) {
	// we first load the addresses and then cache them
	if len(w.addresses) == 0 {
		addresses, err := w.Client.DepositAddresses()
		if err != nil {
			return "", fmt.Errorf("Poloniex: cannot retrieve address for %s: %s", cur, err)
		} else {
			w.addresses = addresses
		}
	}

	address, ok := w.addresses[cur]
	if !ok {
		return "", fmt.Errorf("Poloniex: missing %s address", cur)
	} else {
		return address, nil
	}
}

// ***************** Kraken *****************

type KrakenWithdrawer struct {
	*kraken.Client
}

func NewKrakenWithdrawer(cred Credential) *KrakenWithdrawer {
	c := kraken.NewClient(cred.Key, cred.Secret)
	return &KrakenWithdrawer{c}
}

func (w *KrakenWithdrawer) Exchanger() string {
	return kraken.ExchangerName
}

func (w *KrakenWithdrawer) TradingBalances() (map[string]float64, error) {
	return w.Client.AccountBalance()
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
func (w *KrakenWithdrawer) Withdraw(vol float64, cur, account string) (string, error) {
	// After some testing it appears that the currencies doesn't need to be translated to
	// kraken symbol. It works with ZEC, XZEC, BTC and XBT.
	data := map[string]string{
		"asset":  cur,
		"key":    account,
		"amount": fmt.Sprint(vol),
	}

	resp := map[string]string{}
	err := w.Client.Query("Withdraw", data, &resp)
	if err != nil {
		return "", fmt.Errorf("Kraken: %s withdraw failed - %s", cur, err)
	}

	return resp["refid"], nil
}

func (w *KrakenWithdrawer) WaitBalance(cur string, amount float64) error {
	for {
		bal, err := w.Client.TradeBalance(cur)

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
func (w *KrakenWithdrawer) PaymentAddress(cur string) (string, error) {
	// Apparently kraken does the translation from "BTC" to "XBT"
	data := map[string]string{"asset": cur}
	resp := []map[string]interface{}{}

	err := w.Client.Query("DepositMethods", data, &resp)
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
	err = w.Client.Query("DepositAddresses", data, &resp)
	if err != nil {
		return "", fmt.Errorf("Kraken: call to DepositAddresses failed - %s", err)
	} else if len(resp) == 0 {
		return "", fmt.Errorf("Kraken: missing address for currency %s", cur)
	}

	return resp[0]["address"].(string), nil
}
