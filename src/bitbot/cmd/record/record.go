package main

import (
	"flag"
	"log"
	"time"

	"bitbot/database"
	"bitbot/exchanger/bitfinex"
	"bitbot/exchanger/btce"
	"bitbot/exchanger/cex"
	"bitbot/exchanger/hitbtc"
	"bitbot/exchanger/kraken"
	"bitbot/exchanger/orderbook"
)

// TODO: this code is not panic safe
var (
	dbPath      = flag.String("d", "./data/book.sql", "SQLite database path.")
	periodicity = flag.Int64("t", 5, "Periodicity expressed in seconds.")
	pair        = flag.String("p", "BTC_USD", "Exchanger pair.")
)

type exchanger struct {
	name string
	f    func(string) (*orderbook.OrderBook, error)
}

func main() {
	log.Println("Start recording...")
	flag.Parse()

	db := database.Open(*dbPath)
	defer db.Close()

	exchangers := []*exchanger{
		&exchanger{"hitbtc", hitbtc.OrderBook},
		&exchanger{"bitfinex", bitfinex.OrderBook},
		&exchanger{"btce", btce.OrderBook},
		&exchanger{"kraken", kraken.OrderBook},
		&exchanger{"cex", cex.OrderBook},
	}

	database.CreateTable(db, *pair)

	for {
		for _, e := range exchangers {
			go work(db, e, *pair)
		}

		time.Sleep(time.Duration(*periodicity) * time.Second)
	}
}

func work(db *database.DB, e *exchanger, pair string) {
	log.Printf("Fetching %s...", e.name)
	start := time.Now().UnixNano()
	// TODO: how the timeout is handled
	book, err := e.f(pair)
	end := time.Now().UnixNano()

	if err != nil {
		log.Println(err)
		return
	}

	r := &database.Record{
		Exchanger: e.name,
		StartTime: start,
		EndTime:   end,
		Bids:      book.Bids,
		Asks:      book.Asks,
	}

	database.SaveRecord(db, pair, r)
}
