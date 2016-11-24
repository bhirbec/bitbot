package main

import (
	"flag"
	"log"
	"time"

	"bitbot/exchanger"

	"bitbot/exchanger/hitbtc"
	"bitbot/exchanger/kraken"
	_ "bitbot/exchanger/poloniex"
)

type Exchanger struct {
	name  string
	pairs map[exchanger.Pair]string
	f     func(exchanger.Pair) (*exchanger.OrderBook, error)
}

var (
	p          = flag.String("p", "zec_btc", "Currency pair.")
	configPath = flag.String("config", "", "JSON file that stores exchanger credentials.")
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
		// &Exchanger{poloniex.ExchangerName, poloniex.Pairs, poloniex.OrderBook},
		&Exchanger{kraken.ExchangerName, kraken.Pairs, kraken.OrderBook},
	}

	clients := map[string]Client{
		"Hitbtc": NewHitbtcClient(config["hitbtc"]),
		// "Poloniex": NewPoloniexClient(config["poloniex"]),
		"Kraken": NewKrakenClient(config["kraken"]),
	}

	balances, err := getBalances(clients)
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
			arbitre(clients, arb, pair)

			// TODO: arbitre() should block
			time.Sleep(1 * time.Minute)

			err := rebalance(clients, arb, pair)
			if err != nil {
				log.Println(err)
			}

			balances, err = getBalances(clients)
			if err != nil {
				log.Printf("Cannot retrieve balances: %s", err)
			}
		}

		log.Printf("Waiting %d seconds before fetching orderbooks...\n", periodicity)
		time.Sleep(time.Duration(periodicity) * time.Second)
		printBalances(balances, pair)
	}
}

func arbitre(clients map[string]Client, arb *arbitrage, pair exchanger.Pair) {
	buyClient := clients[arb.buyEx.Exchanger]
	sellClient := clients[arb.sellEx.Exchanger]
	go executeOrder(buyClient, "buy", pair, 0, arb.vol)
	go executeOrder(sellClient, "sell", pair, arb.sellEx.Bids[0].Price, arb.vol)
}

func executeOrder(c Client, side string, pair exchanger.Pair, price, vol float64) {
	log.Printf("Sending %s order on %s: %f %s\n", c.Exchanger(), side, vol, pair)
	ack, err := c.PlaceOrder(side, pair, price, vol)

	if err != nil {
		log.Printf("Cannot execute %s order on %s: %s, %s\n", side, c.Exchanger(), ack, err)
	} else {
		log.Printf("Order sent successfully on %s: %s\n", c.Exchanger(), ack)
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
