package main

import (
	"database/sql"
	"flag"
	"log"
	"sync"
	"time"

	"bitbot/exchanger"

	"bitbot/exchanger/hitbtc"
	"bitbot/exchanger/kraken"
	"bitbot/exchanger/poloniex"
)

type Exchanger struct {
	name  string
	pairs map[exchanger.Pair]string
	f     func(exchanger.Pair) (*exchanger.OrderBook, error)
}

var (
	p          = flag.String("p", "zec_btc", "Currency pair.")
	configPath = flag.String("config", "ansible/secrets/trader.json", "JSON file that stores exchanger credentials.")
)

var pairs = map[string]exchanger.Pair{
	"eth_btc": exchanger.ETH_BTC,
	"zec_btc": exchanger.ZEC_BTC,
	"ltc_btc": exchanger.LTC_BTC,
}

const (
	periodicity = 20
	minSpread   = 0.8
	minVol      = 0.1
)

func main() {
	log.Println("Start trader...")
	flag.Parse()

	config, err := LoadConfig(*configPath)
	if err != nil {
		log.Panic(err)
	}

	pair, ok := pairs[*p]
	if !ok {
		log.Panicf("Unsupported pair %s\n", *p)
	}

	var exchangers = []*Exchanger{
		&Exchanger{hitbtc.ExchangerName, hitbtc.Pairs, hitbtc.OrderBook},
		&Exchanger{poloniex.ExchangerName, poloniex.Pairs, poloniex.OrderBook},
		&Exchanger{kraken.ExchangerName, kraken.Pairs, kraken.OrderBook},
	}

	traders := map[string]Trader{
		"Hitbtc":   NewHitbtcTrader(config.Hitbtc),
		"Poloniex": NewPoloniexTrader(config.Poloniex),
		"Kraken":   NewKrakenTrader(config.Kraken),
	}

	withdrawers := map[string]Withdrawer{
		"Hitbtc":   NewHitbtcWithdrawer(config.Hitbtc),
		"Poloniex": NewPoloniexWithdrawer(config.Poloniex),
		"Kraken":   NewKrakenWithdrawer(config.Kraken),
	}

	go startSyncTrades(config)

	balances, err := getBalances(withdrawers)
	if err != nil {
		log.Printf("Cannot retrieve balances: %s", err)
	} else {
		printBalances(balances, pair)
	}

	for {
		for arb := range findArbitages(pair, exchangers) {
			if arb.spread < minSpread || arb.vol < minVol {
				break
			}

			log.Println(arb)
			availableSellVol := balances[arb.sellEx.Exchanger][pair.Base]
			availableBuyVol := 0.95 * (balances[arb.buyEx.Exchanger][pair.Quote] / arb.buyEx.Asks[0].Price)
			arb.vol = minFloat64(arb.vol, availableSellVol, availableBuyVol)
			arbitre(traders, arb)

			// TODO: arbitre() should block?
			time.Sleep(1 * time.Minute)

			wg := sync.WaitGroup{}
			wg.Add(2)

			go func() {
				defer wg.Done()
				ExecRebalanceTransactions(withdrawers, pair.Base)
			}()

			// ExecRebalanceTransactions triggers several API requests. With latency issues, the exchanger
			// could receive requests in a different order than what we sent. This involves that the nounce
			// will be invalid and a "Kraken errors: [EAPI:Invalid nonce]" can occur. To fix this quickly
			// we just wait 10 seconds here...
			time.Sleep(10 * time.Second)

			go func() {
				defer wg.Done()
				ExecRebalanceTransactions(withdrawers, pair.Quote)
			}()

			wg.Wait()

			balances, err = getBalances(withdrawers)
			if err != nil {
				log.Printf("Cannot retrieve balances: %s", err)
			}

			break
		}

		log.Printf("Waiting %d seconds before fetching orderbooks...\n", periodicity)
		time.Sleep(time.Duration(periodicity) * time.Second)
		printBalances(balances, pair)
	}
}

func arbitre(traders map[string]Trader, arb *arbitrage) {
	db, err := OpenMysql()
	if err != nil {
		log.Printf("executeOrder: cannot open db %s\n", err)
		return
	}

	ex1 := arb.buyEx.Exchanger
	go executeOrder(db, traders[ex1], arb.id, ex1, "buy", arb.pair, arb.buyEx.Asks[0].Price, arb.vol)

	ex2 := arb.sellEx.Exchanger
	go executeOrder(db, traders[ex2], arb.id, ex2, "sell", arb.pair, arb.sellEx.Bids[0].Price, arb.vol)

	err = saveArbitrage(db, arb)
	if err != nil {
		log.Printf("saveArbitrage failed - %s\n", err)
	}
}

func executeOrder(db *sql.DB, t Trader, arbId, ex, side string, pair exchanger.Pair, price, vol float64) {
	log.Printf("%s: side: %s | pair: %s | price: %f | vol: %f\n", ex, side, pair, price, vol)

	ids, err := t.PlaceOrder(side, pair, price, vol)
	if err != nil {
		log.Printf("Cannot execute %s order on %s: %s\n", side, ex, err)
		return
	} else {
		log.Printf("Order sent successfully on %s\n", ex)
	}

	// TODO: batch this operation with one insert
	for _, id := range ids {
		err = saveOrderAck(db, arbId, id, pair.String(), ex, side)
		if err != nil {
			log.Printf("saveOrderAck failed - %s\n", err)
		}
	}
}

func minFloat64(vols ...float64) float64 {
	var m float64 = vols[0]
	for _, v := range vols[1:] {
		if v < m {
			m = v
		}
	}
	return m
}
