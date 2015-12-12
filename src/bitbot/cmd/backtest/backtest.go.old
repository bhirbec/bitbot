package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"time"

	"bitbot/database"
)

// TODO: this code is not panic safe

const (
	timeFormat         = "2006-01-02 15:04:05.000"
	maxRequestDuration = time.Duration(500 * time.Millisecond)
)

func main() {
	log.Println("Testing strat...")

	flag.Parse()
	dbPath := flag.Arg(0)

	db := database.Open(dbPath)
	exchangers := []string{"bitfinex", "kraken", "hitbtc", "btce", "cex"}
	strategy(db, "BTC_USD", 500, exchangers)
}

type PairPosition struct {
	USD float64
	BTC float64
}

func strategy(db *database.DB, pair string, investment int64, exchangers []string) {

	n := len(exchangers)
	invest := float64(investment) / float64(n)
	recordStream := database.SelectRecords(db, pair)

	// initialize positions with the initial investment
	pairPositions := map[string]*PairPosition{}

	for r := range recordStream {
		_, ok := pairPositions[r.Exchanger]

		if !ok {
			// TODO: compute a more accurate buying price than r.Asks[0].Price
			usd := invest * 0.5
			btc := usd / r.Asks[0].Price
			pairPositions[r.Exchanger] = &PairPosition{USD: usd, BTC: btc}
		}

		// TODO: log something if this condition is not met after x iterations
		if len(pairPositions) == n {
			break
		}
	}

	// scan records to find opportunities
	records := map[string]*database.Record{}

	for r1 := range recordStream {
		oldRec, ok := records[r1.Exchanger]

		// skip records with older startime (latency)
		if ok && r1.StartTime < oldRec.StartTime {
			continue
		}

		records[r1.Exchanger] = r1

		for ex, r2 := range records {
			if ex == r1.Exchanger {
				continue
			}

			start := minTime(r1.StartDate, r2.StartDate)
			end := maxTime(r1.EndDate, r2.EndDate)

			// skip request not on the same time range
			if end.Sub(start) > maxRequestDuration {
				continue
			}

			var buy, sell *database.Record

			if r1.Asks[0].Price < r2.Bids[0].Price {
				buy, sell = r1, r2
			} else if r2.Asks[0].Price < r1.Bids[0].Price {
				buy, sell = r2, r1
			} else {
				continue
			}

			ask := buy.Asks[0]
			bid := sell.Bids[0]
			vol := math.Min(ask.Volume, bid.Volume)
			spread := bid.Price/ask.Price - 1

			if spread < 0.04 || vol < 0.1 {
				continue
			}

			date := start.Format(timeFormat)
			pAndl := vol * (bid.Price - ask.Price)
			fmt.Printf("%s | %6.2f%% %6.2f$ | buy %-10s %6.2f %6.2f | sell %-10s %6.2f %6.2f\n", date, 100*spread, pAndl, buy.Exchanger, ask.Price, ask.Volume, sell.Exchanger, bid.Price, bid.Volume)
		}
	}
}

func minTime(t1, t2 time.Time) time.Time {
	if t1.Before(t2) {
		return t1
	}
	return t2
}

func maxTime(t1, t2 time.Time) time.Time {
	if t1.After(t2) {
		return t1
	}
	return t2
}
