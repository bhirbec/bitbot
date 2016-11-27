package main

import (
	"log"
	"testing"
	"time"

	"bitbot/exchanger"
)

type TestTrader struct {
	exchangerName string
	balances      map[string]map[string]float64
}

func (t *TestTrader) Exchanger() string {
	return t.exchangerName
}

func (t *TestTrader) TradingBalances() (map[string]float64, error) {
	return t.balances[t.exchangerName], nil
}

func (t *TestTrader) PlaceOrder(side string, pair exchanger.Pair, price, vol float64) (map[string]interface{}, error) {
	return nil, nil
}

func (t *TestTrader) PaymentAddress(cur string) (string, error) {
	return t.exchangerName, nil
}

func (t *TestTrader) Withdraw(vol float64, cur, address string) (string, error) {
	t.balances[t.exchangerName][cur] -= vol
	go func() {
		time.Sleep(1 * time.Second)
		t.balances[address][cur] += vol
	}()
	return "", nil
}

func (t *TestTrader) WaitBalance(cur string, amount float64) error {
	for {
		bal, err := t.TradingBalances()
		if err != nil {
			log.Printf("waitBalanceChange: call to TradingBalances() failed - %s (%s)\n", err, cur)
		} else if bal[cur] >= amount {
			break
		}

		log.Printf("WaitBalance: Wait until %s %s balance is >= than %f\n", t.Exchanger(), cur, amount)
		time.Sleep(1 * time.Second)
	}

	return nil
}

func TestExecRebalanceTransactions(t *testing.T) {
	const cur = "CUR-1"

	balances := map[string]map[string]float64{
		"market1": map[string]float64{cur: 1},
		"market2": map[string]float64{cur: 19},
		"market3": map[string]float64{cur: 11},
		"market4": map[string]float64{cur: 9},
	}

	t1 := &TestTrader{"market1", balances}
	t2 := &TestTrader{"market2", balances}
	t3 := &TestTrader{"market3", balances}
	t4 := &TestTrader{"market4", balances}

	traders := map[string]Trader{}
	for _, t := range []Trader{t1, t2, t3, t4} {
		traders[t.Exchanger()] = t
	}

	ExecRebalanceTransactions(traders, cur)

	b1, _ := t1.TradingBalances()
	if amount := b1[cur]; amount != 10 {
		t.Errorf("Trader balance not correct - 10 expected got %f", amount)
	}
}
