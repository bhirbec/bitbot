package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"bitbot/errorutils"
	"bitbot/exchanger"
)

type bookFunc func(exchanger.Pair) (*exchanger.OrderBook, error)

type arbitrage struct {
	id     string
	ts     time.Time
	pair   exchanger.Pair
	buyEx  *exchanger.OrderBook
	sellEx *exchanger.OrderBook
	vol    float64
	spread float64
}

func (a *arbitrage) String() string {
	return fmt.Sprintf("Buy %s at %2f and Sell %s at %2f | pair: %s | spread: %2.2f%% | vol: %f",
		a.buyEx.Exchanger,
		a.buyEx.Asks[0].Price,
		a.sellEx.Exchanger,
		a.sellEx.Bids[0].Price,
		a.pair,
		a.spread,
		a.vol,
	)
}

func findArbitages(pair exchanger.Pair, bookFuncs map[string]bookFunc) chan *arbitrage {
	c := make(chan *arbitrage)

	go func() {
		defer close(c)
		obs := []*exchanger.OrderBook{}

		for b := range getBooks(pair, bookFuncs) {
			for _, ob := range obs {
				arbitrage := computeArbitrage(pair, b, ob)
				if arbitrage != nil {
					c <- arbitrage
				}

				arbitrage = computeArbitrage(pair, ob, b)
				if arbitrage != nil {
					c <- arbitrage
				}
			}

			obs = append(obs, b)
		}
	}()

	return c
}

func getBooks(pair exchanger.Pair, bookFuncs map[string]bookFunc) chan *exchanger.OrderBook {
	var wg sync.WaitGroup
	c := make(chan *exchanger.OrderBook)

	var getBook = func(pair exchanger.Pair, ex string, f bookFunc) {
		defer wg.Done()
		defer errorutils.LogPanic()

		log.Printf("Fetching %s orderbook for pair %s...", ex, pair)
		book, err := f(pair)

		if err == nil {
			c <- book
		} else {
			log.Println("getBooks: failed to retrieve %s orderbook for pair %s - %s", ex, pair, err)
		}
	}

	for ex, f := range bookFuncs {
		wg.Add(1)
		go getBook(pair, ex, f)
	}

	go func() {
		wg.Wait()
		close(c)
	}()

	return c
}

func computeArbitrage(pair exchanger.Pair, buyEx, sellEx *exchanger.OrderBook) *arbitrage {
	buyOrder := buyEx.Asks[0]
	sellOrder := sellEx.Bids[0]
	ts := time.Now()

	if buyOrder.Price >= sellOrder.Price {
		return nil
	}

	return &arbitrage{
		id:     arbId(ts, pair, buyEx.Exchanger, sellEx.Exchanger),
		ts:     ts,
		pair:   pair,
		buyEx:  buyEx,
		sellEx: sellEx,
		vol:    math.Min(buyOrder.Volume, sellOrder.Volume),
		spread: 100 * (sellOrder.Price/buyOrder.Price - 1),
	}
}

func arbId(ts time.Time, pair exchanger.Pair, buyEx, sellEx string) string {
	key := fmt.Sprintf("%d-%s-%s-%s", ts.UnixNano(), pair.String(), buyEx, sellEx)
	h := md5.New()
	h.Write([]byte(key))
	b := h.Sum(nil)
	return hex.EncodeToString(b)
}
