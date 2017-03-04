package main

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"bitbot/exchanger"
)

type transaction struct {
	orig   string
	dest   string
	amount float64
}

const threshold = 0.05

func rebalance(withdrawers map[string]Withdrawer, pair exchanger.Pair) {
	wg := sync.WaitGroup{}

	f := func(cur string) {
		wg.Add(1)
		defer wg.Done()
		execRebalanceTransactions(withdrawers, cur)
	}

	// execRebalanceTransactions triggers several API requests. With latency issues, the exchanger
	// could receive requests in a different order than what we sent. This involves that the nounce
	// will be invalid and a "Kraken errors: [EAPI:Invalid nonce]" can occur. To fix this quickly
	// we just wait 10 seconds here...
	go f(pair.Base)
	time.Sleep(10 * time.Second)
	go f(pair.Quote)
	wg.Wait()
}

func execRebalanceTransactions(withdrawers map[string]Withdrawer, cur string) {
	curBal, err := getCurrencyBalances(cur, withdrawers)
	if err != nil {
		log.Printf("execRebalanceTransactions: call to getCurrencyBalances() failed - %s (%s)\n", err, cur)
		return
	}

	rebalanced := map[string]bool{}

	for _, t := range findRebalanceTransactions(curBal) {
		err := execTransaction(withdrawers[t.orig], withdrawers[t.dest], cur, t.amount)
		if err != nil {
			log.Printf("execRebalanceTransactions: call to execTransaction() failed - %s (%s)\n", err, cur)
		} else {
			rebalanced[t.dest] = true
		}
	}

	for len(rebalanced) > 0 {
		log.Printf("execRebalanceTransactions: waiting for %s transfer to complete\n", cur)
		time.Sleep(1 * time.Minute)

		curBal, err := getCurrencyBalances(cur, withdrawers)
		if err != nil {
			log.Printf("execRebalanceTransactions: call to getCurrencyBalances() failed - %s (%s)\n", err, cur)
			continue
		}

		total := sumBalance(curBal)
		for ex, _ := range rebalanced {
			err := withdrawers[ex].AfterWithdraw(cur)
			if err != nil {
				log.Printf("execRebalanceTransactions: call to AfterWithdraw() failed - %s (%s)\n", err, cur)
				continue
			}

			alloc := curBal[ex] / total
			if alloc >= threshold {
				delete(rebalanced, ex)
			}
		}
	}
}

func findRebalanceTransactions(balances map[string]float64) []*transaction {
	total := sumBalance(balances)
	targetBal := total / float64(len(balances))
	positives := map[string]float64{}
	negatives := map[string]float64{}

	for exchanger, balance := range balances {
		alloc := balance / total
		delta := balance - targetBal

		if alloc < threshold {
			negatives[exchanger] = -delta
		} else if delta > 0 {
			positives[exchanger] = delta
		}
	}

	var amount float64
	transactions := []*transaction{}

	for dest, negDelta := range negatives {
		for orig, posDelta := range positives {
			if posDelta <= 0 || negDelta == 0 {
				continue
			} else if posDelta > negDelta {
				amount = negDelta
				positives[orig] -= amount
			} else {
				amount = posDelta
				negDelta -= posDelta
				delete(positives, orig)
			}

			t := &transaction{orig, dest, amount}
			transactions = append(transactions, t)
		}
	}

	return transactions
}

func sumBalance(balances map[string]float64) float64 {
	var total float64
	for _, balance := range balances {
		total += balance
	}
	return total
}

func execTransaction(org, dest Withdrawer, cur string, vol float64) error {
	log.Printf("Starting transfert of %f %s from %s to %s\n", vol, cur, org.Exchanger(), dest.Exchanger())

	var address string
	var err error

	if org.Exchanger() == "Kraken" {
		// Kraken requires to input the withdrawal addresses in the UI and to
		// give them unique name. The convention is ExchangerName + "-" + cur.
		// Example: Poloniex-ZEC
		exName := strings.Replace(dest.Exchanger(), " ", "-", -1)
		address = fmt.Sprintf("%s-%s", exName, cur)
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

	return nil
}

func getCurrencyBalances(cur string, withdrawers map[string]Withdrawer) (map[string]float64, error) {
	masterBal, err := getBalances(withdrawers)
	if err != nil {
		return nil, fmt.Errorf("getCurrencyBalances: call to getBalances() failed - %s (%s)", err, cur)
	}

	curBal := map[string]float64{}
	for ex, bal := range masterBal {
		curBal[ex] = bal[cur]
	}

	return curBal, nil
}

func getBalances(withdrawers map[string]Withdrawer) (map[string]map[string]float64, error) {
	out := map[string]map[string]float64{}

	for _, w := range withdrawers {
		b, err := w.TradingBalances()
		if err != nil {
			return nil, err
		}
		out[w.Exchanger()] = b
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
