package main

import (
	"testing"
	"time"
)

type TestWithdrawer struct {
	exchangerName string
	balances      map[string]map[string]float64
}

func (t *TestWithdrawer) Exchanger() string {
	return t.exchangerName
}

func (t *TestWithdrawer) TradingBalances() (map[string]float64, error) {
	return t.balances[t.exchangerName], nil
}

func (t *TestWithdrawer) PaymentAddress(cur string) (string, error) {
	return t.exchangerName, nil
}

func (t *TestWithdrawer) Withdraw(vol float64, cur, address string) (string, error) {
	t.balances[t.exchangerName][cur] -= vol
	go func() {
		time.Sleep(1 * time.Second)
		t.balances[address][cur] += vol
	}()
	return "", nil
}

func (t *TestWithdrawer) AfterWithdraw(cur string) error {
	return nil
}

func TestexecRebalanceTransactions(t *testing.T) {
	const cur = "CUR-1"

	balances := map[string]map[string]float64{
		"market1": map[string]float64{cur: 1},
		"market2": map[string]float64{cur: 19},
		"market3": map[string]float64{cur: 11},
		"market4": map[string]float64{cur: 9},
	}

	w1 := &TestWithdrawer{"market1", balances}
	w2 := &TestWithdrawer{"market2", balances}
	w3 := &TestWithdrawer{"market3", balances}
	w4 := &TestWithdrawer{"market4", balances}

	Withdrawers := map[string]Withdrawer{}
	for _, w := range []Withdrawer{w1, w2, w3, w4} {
		Withdrawers[w.Exchanger()] = w
	}

	execRebalanceTransactions(Withdrawers, cur)

	b1, _ := w1.TradingBalances()
	if amount := b1[cur]; amount != 10 {
		t.Errorf("Trader balance not correct - 10 expected got %f", amount)
	}
}
