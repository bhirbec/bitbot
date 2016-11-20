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
	var hitbtcClient = hitbtc.NewClient(cred.Key, cred.Secret)

	cred = config["poloniex"]
	var poloniexClient = poloniex.NewClient(cred.Key, cred.Secret)

	balances, err := getBalances(hitbtcClient, poloniexClient)
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
			arbitre(hitbtcClient, poloniexClient, arb, pair)

			// TODO: arbitre() should block
			time.Sleep(1 * time.Minute)

			balances, err = rebalance(hitbtcClient, poloniexClient, arb, pair)
			if err != nil {
				log.Println(err)
			}
		}

		log.Printf("Waiting %d seconds before fetching orderbooks...\n", periodicity)
		time.Sleep(time.Duration(periodicity) * time.Second)
		printBalances(balances, pair)
	}
}

func arbitre(h *hitbtc.Client, p *poloniex.Client, arb *arbitrage, pair exchanger.Pair) {
	if arb.buyEx.Exchanger == "Hitbtc" {
		go executeHitbtc(h, "buy", pair, 0, arb.vol)
		go executePoloniex(p, "sell", pair, arb.sellEx.Bids[0].Price, arb.vol)
	} else {
		go executePoloniex(p, "buy", pair, arb.buyEx.Asks[0].Price, arb.vol)
		go executeHitbtc(h, "sell", pair, 0, arb.vol)
	}
}

func executeHitbtc(h *hitbtc.Client, side string, pair exchanger.Pair, price, vol float64) {
	log.Printf("Sending %s order on Hitbtc: %f %s\n", side, vol, exchanger.ZEC_BTC)
	ack, err := h.PlaceOrder(side, pair, 0, vol, "market")
	if err != nil {
		// { ExecutionReport: {
		//     orderStatus:rejected
		//     side:sell
		//     userId:user_142834
		//     symbol:ZECBTC
		//     timeInForce:IOC
		//     lastPrice:
		//     orderRejectReason:badQuantity
		//     orderId:N/A
		//     averagePrice:0
		//     execReportType:rejected
		//     type:market
		//     leavesQuantity:0
		//     lastQuantity:0
		//     cumQuantity:0
		//     clientOrderId:hitbtc-1479089075799175
		//     quantity:0}
		// }

		// {"code":"InvalidArgument","message":"Fields are not valid: quantity"}
		log.Printf("Cannot execute `%s` order on Hitbtc: %s, %s\n", side, ack, err)
	} else {
		// {ExecutionReport: {
		//     orderStatus:new
		//     leavesQuantity:63
		//     timestamp:1.479186944972e+12
		//     side:buy
		//     type:market
		//     timeInForce:IOC
		//     created:1.479186944972e+12
		//     execReportType:new
		//     averagePrice:0
		//     orderId:564643943
		//     userId:user_142834
		//     symbol:ZECBTC
		//     lastQuantity:0
		//     lastPrice:
		//     cumQuantity:0
		//     clientOrderId:hitbtc-1479186945249209
		//     quantity:63}
		// }
		log.Println("Hitbtc order successed", ack)
	}
}

func executePoloniex(p *poloniex.Client, side string, pair exchanger.Pair, price, vol float64) {
	var ack interface{}
	var err error

	log.Printf("Sending %s order on Poloniex: %f %s\n", side, vol, exchanger.ZEC_BTC)
	if side == "buy" {
		ack, err = p.Buy(pair, price, vol)
	} else {
		ack, err = p.Sell(pair, price, vol)
	}

	if err != nil {
		// {error: Total must be at least 0.0001.}
		log.Printf("Cannot executfe `%s` order on Poloniex: %s", err, side)
	} else {
		// {
		//  orderNumber:3073419682
		// 	resultingTrades: {
		// 		amount:0.06300000
		// 		date:2016-11-15 05:15:45
		// 		rate:0.15445400
		// 		total:0.00973060
		// 		tradeID:620559
		// 		type:sell
		// 	 }
		// }
		log.Println("Poloniex order successed", ack)
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
