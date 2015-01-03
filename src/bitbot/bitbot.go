// docker build -t bitbot-img . && docker run --rm bitbot-img
package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"exchanger/bitfinex"
	"exchanger/btce"
	"exchanger/bter"
	"exchanger/cex"
	"exchanger/hitbtc"
	"exchanger/kraken"
	"exchanger/orderbook"
)

const csvpath = "data/orderbook.csv"

type orderBookFunc func(string) (*orderbook.OrderBook, error)

func main() {
	log.Println("Starting bitbot...")

	funcs := []orderBookFunc{
		hitbtc.OrderBook,
		bitfinex.OrderBook,
		bter.OrderBook,
		btce.OrderBook,
		kraken.OrderBook,
		cex.OrderBook,
	}

	exchangers := []string{
		hitbtc.ExchangerName,
		bitfinex.ExchangerName,
		bter.ExchangerName,
		btce.ExchangerName,
		kraken.ExchangerName,
		cex.ExchangerName,
	}

	f, err := os.Create(csvpath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	w := csv.NewWriter(f)

	writeCSVHeader(w, exchangers)

	for i := 0; i < 1; i++ {
		orderbooks := fetchOrderbooks(funcs, "BTC_USD")
		detectArbitrage(orderbooks)
		writeCSVRow(w, orderbooks)
		time.Sleep(2 * time.Second)
	}

	log.Println("Stopping bitbot...")
}

func fetchOrderbooks(funcs []orderBookFunc, pair string) []*orderbook.OrderBook {
	type partial struct {
		orderbook *orderbook.OrderBook
		err       error
		rank      int
	}

	// fetch orderbooks concurrently
	partials := make(chan *partial)

	for i, f := range funcs {
		go func(i int, f orderBookFunc) {
			book, err := f(pair)
			partials <- &partial{book, err, i}
		}(i, f)
	}

	// get orderbooks when they're ready
	orderbooks := make([]*orderbook.OrderBook, len(funcs))
	for i := 0; i < len(funcs); i++ {
		p := <-partials
		if p.err != nil {
			log.Println(p.err)
			continue
		}
		orderbooks[p.rank] = p.orderbook
	}

	return orderbooks
}

func writeCSVHeader(w *csv.Writer, exchangers []string) {
	header := []string{}
	for _, e := range exchangers {
		header = append(header, fmt.Sprintf("%s bid", e))
		header = append(header, fmt.Sprintf("%s bid vol", e))
		header = append(header, fmt.Sprintf("%s ask", e))
		header = append(header, fmt.Sprintf("%s ask vol", e))
		header = append(header, fmt.Sprintf("%s spread %%", e))
	}
	w.Write(header)
}

func writeCSVRow(w *csv.Writer, orderbooks []*orderbook.OrderBook) {
	n := 5 //
	row := make([]string, n*len(orderbooks))
	for i, ob := range orderbooks {
		if ob != nil {
			bid, ask := ob.Bids[0], ob.Asks[0]
			row[i*n] = fmt.Sprintf("%.6f", bid.Price)
			row[i*n+1] = fmt.Sprintf("%.6f", bid.Volume)
			row[i*n+2] = fmt.Sprintf("%.6f", ask.Price)
			row[i*n+3] = fmt.Sprintf("%.6f", ask.Volume)
			row[i*n+4] = fmt.Sprintf("%.6f", (ask.Price/bid.Price-1.0)*100.0)
		}
	}
	w.Write(row)
	w.Flush()
}

func detectArbitrage(orderbooks []*orderbook.OrderBook) {
	// scan orderbooks to detect arbitrage opportunities
	l := len(orderbooks)
	for i := 0; i < l-1; i++ {
		if ob1 := orderbooks[i]; ob1 != nil {
			for j := i + 1; j < l; j++ {
				if ob2 := orderbooks[j]; ob2 != nil {
					if r := detectOpportunity(ob1, ob2); r != "" {
						log.Println(r)
					}
				}
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
