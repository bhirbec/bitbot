package main

import (
	"fmt"
	"log"
	"math"
	"sync"

	"bitbot/errorutils"
	"bitbot/exchanger"
)

type arbitrage struct {
	buyEx  *exchanger.OrderBook
	sellEx *exchanger.OrderBook
	vol    float64
	spread float64
}

func (a *arbitrage) String() string {
	return fmt.Sprintf("Buy %s at %2f and Sell %s at %2f | spread: %2.2f%% | vol: %f",
		a.buyEx.Exchanger,
		a.buyEx.Asks[0].Price,
		a.sellEx.Exchanger,
		a.sellEx.Bids[0].Price,
		a.spread,
		a.vol,
	)
}

func findArbitages(pair exchanger.Pair, exchangers []*Exchanger) chan *arbitrage {
	c := make(chan *arbitrage)

	go func() {
		defer close(c)
		obs := []*exchanger.OrderBook{}

		for b := range getBooks(pair, exchangers) {
			for _, ob := range obs {
				arbitrage := computeArbitrage(b, ob)
				if arbitrage != nil {
					c <- arbitrage
				}

				arbitrage = computeArbitrage(ob, b)
				if arbitrage != nil {
					c <- arbitrage
				}
			}

			obs = append(obs, b)
		}
	}()

	return c
}

func getBooks(pair exchanger.Pair, exchangers []*Exchanger) chan *exchanger.OrderBook {
	var wg sync.WaitGroup
	c := make(chan *exchanger.OrderBook)

	var getBook = func(pair exchanger.Pair, e *Exchanger) {
		defer wg.Done()
		defer errorutils.LogPanic()

		log.Printf("Fetching %s orderbook for pair %s...", e.name, pair)
		book, err := e.f(pair)

		if err != nil {
			log.Println(err)
			return
		}

		c <- book
	}

	for _, e := range exchangers {
		if _, ok := e.pairs[pair]; ok {
			wg.Add(1)
			go getBook(pair, e)
		}
	}

	go func() {
		wg.Wait()
		close(c)
	}()

	return c
}

func computeArbitrage(buyEx, sellEx *exchanger.OrderBook) *arbitrage {
	buyOrder := buyEx.Asks[0]
	sellOrder := sellEx.Bids[0]

	if buyOrder.Price >= sellOrder.Price {
		return nil
	}

	return &arbitrage{
		buyEx:  buyEx,
		sellEx: sellEx,
		vol:    math.Min(buyOrder.Volume, sellOrder.Volume),
		spread: 100 * (sellOrder.Price/buyOrder.Price - 1),
	}
}
