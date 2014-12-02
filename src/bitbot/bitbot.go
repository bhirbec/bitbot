// docker build -t bitbot-img . && docker run --rm bitbot-img
package main

import (
	"fmt"
	"log"
	"math"
	"time"

	"exchanger/bitfinex"
	"exchanger/btce"
	"exchanger/bter"
	"exchanger/hitbtc"
	"exchanger/kraken"
	"exchanger/orderbook"
)

type market struct {
	f    func(string) (*orderbook.OrderBook, error)
	pair string
}

func main() {
	log.Println("Starting bitbot...")

	markets := []*market{
		&market{hitbtc.OrderBook, "BTCUSD"},
		&market{bitfinex.OrderBook, "BTCUSD"},
		&market{bter.OrderBook, "BTC_USD"},
		&market{btce.OrderBook, "btc_usd"},
		&market{kraken.OrderBook, "XXBTZUSD"},
	}

	for i := 0; i < 10; i++ {
		orderbooks := fetchOrderbooks(markets)
		detectArbitrage(orderbooks)
		time.Sleep(2 * time.Second)
	}

	log.Println("Stopping bitbot...")
}

func fetchOrderbooks(markets []*market) []*orderbook.OrderBook {
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
			log.Println(p.err)
			continue
		}
		orderbooks = append(orderbooks, p.orderbook)
	}
	return orderbooks
}

func detectArbitrage(orderbooks []*orderbook.OrderBook) {
	// scan orderbooks to detect arbitrage opportunities
	l := len(orderbooks)
	for i := 0; i < l-1; i++ {
		ob1 := orderbooks[i]
		for j := i + 1; j < l; j++ {
			ob2 := orderbooks[j]
			if r := detectOpportunity(ob1, ob2); r != "" {
				log.Println(r)
			}
		}
	}
}

func detectOpportunity(ob1, ob2 *orderbook.OrderBook) string {
	if ask, bid := ob1.Asks[0], ob2.Bids[0]; ask.Price < bid.Price {
		diff := math.Min(ask.Volume, bid.Volume) * (bid.Price - ask.Price)
		profit := 100 * (bid.Price/ask.Price - 1)
		return fmt.Sprintf("%.2f%% %#v | buy %s %#v/%#v | sell %s %#v/%#v", profit, diff, ob1.Exchanger, ask.Price, ask.Volume, ob2.Exchanger, bid.Price, bid.Volume)
	} else if ask, bid := ob2.Asks[0], ob1.Bids[0]; ask.Price < bid.Price {
		diff := math.Min(ask.Volume, bid.Volume) * (bid.Price - ask.Price)
		profit := 100 * (bid.Price/ask.Price - 1)
		return fmt.Sprintf("%.2f%% %#v | buy %s %#v/%#v | sell %s %#v/%#v", profit, diff, ob2.Exchanger, ask.Price, ask.Volume, ob1.Exchanger, bid.Price, bid.Volume)
	} else {
		return ""
	}
}
