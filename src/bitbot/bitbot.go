// docker build -t bitbot-img . && docker run --rm bitbot-img
package main

import (
	"fmt"
	"time"

	"exchanger/bitfinex"
	"exchanger/btce"
	"exchanger/bter"
	"exchanger/hitbtc"
	"exchanger/orderbook"
)

type market struct {
	f    func(string) (*orderbook.OrderBook, error)
	pair string
}

func main() {
	markets := []*market{
		&market{hitbtc.OrderBook, "LTCBTC"},
		&market{bitfinex.OrderBook, "LTCBTC"},
		&market{bter.OrderBook, "LTC_BTC"},
		&market{btce.OrderBook, "ltc_btc"},
	}

	for i := 0; i < 10; i++ {
		fmt.Println("*********")
		detect(markets)
		time.Sleep(2 * time.Second)
	}
}

func detect(markets []*market) {
	// fetch orderbooks concurrently
	type partial struct {
		orderbook *orderbook.OrderBook
		err       error
	}

	partials := make(chan *partial)

	for _, m := range markets {
		go func(m *market) {
			book, err := m.f(m.pair)
			partials <- &partial{book, err}
		}(m)
	}

	// get orderbooks when they're ready
	orderbooks := []*orderbook.OrderBook{}
	for i := 0; i < len(markets); i++ {
		p := <-partials
		if p.err != nil {
			fmt.Println(p.err)
			continue
		}
		orderbooks = append(orderbooks, p.orderbook)
	}

	// scan orderbooks to detect arbitrage opportunities
	l := len(orderbooks)
	for i := 0; i < l-1; i++ {
		ob1 := orderbooks[i]
		for j := i + 1; j < l; j++ {
			ob2 := orderbooks[j]
			r := detectArbitrage(ob1, ob2)
			fmt.Println(r)
		}
	}
}

func detectArbitrage(ob1, ob2 *orderbook.OrderBook) string {
	if ask, bid := ob1.Asks[0], ob2.Bids[0]; ask.Price < bid.Price {
		return fmt.Sprintf("Buy %s %#v/%#v | sell %s %#v/%#v", ob1.Exchanger, ask.Price, ask.Volume, ob2.Exchanger, bid.Price, bid.Volume)
	} else if ask, bid := ob2.Asks[0], ob1.Bids[0]; ask.Price < bid.Price {
		return fmt.Sprintf("Buy %s %#v/%#v | sell %s %#v/%#v", ob2.Exchanger, ask.Price, ask.Volume, ob1.Exchanger, bid.Price, bid.Volume)
	} else {
		return fmt.Sprintf("No arbitrage between %s and %s", ob1.Exchanger, ob2.Exchanger)
	}
}
