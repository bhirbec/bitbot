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

func main() {
	for i := 0; i < 10; i++ {
		fmt.Println("*********")
		detect()
		time.Sleep(2 * time.Second)
	}
}

func detect() {
	type partial struct {
		orderbook *orderbook.OrderBook
		err       error
	}

	partials := make(chan *partial)

	// fetch orderbooks concurrently
	go func() {
		book, err := hitbtc.OrderBook("LTCBTC")
		partials <- &partial{book, err}
	}()

	go func() {
		book, err := bitfinex.OrderBook("LTCBTC")
		partials <- &partial{book, err}
	}()

	go func() {
		book, err := bter.OrderBook("LTC_BTC")
		partials <- &partial{book, err}
	}()

	go func() {
		book, err := btce.OrderBook("ltc_btc")
		partials <- &partial{book, err}
	}()

	// get orderbooks when they're ready
	orderbooks := []*orderbook.OrderBook{}
	for i := 0; i < 4; i++ {
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
