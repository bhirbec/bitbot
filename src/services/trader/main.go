package main

import (
	"flag"
	"log"
	"time"

	"bitbot/exchanger"

	"bitbot/exchanger/hitbtc"
	"bitbot/exchanger/poloniex"
)

// external_data:76c7a3678b69011963976a814782b7e3b724a24fb247ead6c9845882e707547d
// type:payout created:1.479274795e+09
// currency_code_to:ZEC
// destination_data:t1aUX5GXFFpxtDugJpfY2Zabv2mGvxKQ9a5
// bitcoin_address:
// bitcoin_return_address:
// id:d317d2e9-77b9-4750-bde3-29487c06592e

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
		&Exchanger{poloniex.ExchangerName, poloniex.Pairs, poloniex.OrderBook},
	}

	cred := config["hitbtc"]
	var hitbtcClient = NewHitbtcClient(cred)

	cred = config["poloniex"]
	var poloniexClient = NewPoloniexClient(cred)

	clients := map[string]Client{
		"Hitbtc":   hitbtcClient,
		"Poloniex": poloniexClient,
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

			err := rebalance(hitbtcClient, poloniexClient, arb, pair)
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
