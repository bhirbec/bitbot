package main

import (
	"flag"
	"log"
	"runtime"
	"time"

	"bitbot/database"
	"bitbot/exchanger/bitfinex"
	"bitbot/exchanger/btce"
	"bitbot/exchanger/cex"
	"bitbot/exchanger/hitbtc"
	"bitbot/exchanger/kraken"
	"bitbot/exchanger/orderbook"
)

var (
	dbPath      = flag.String("d", "./data/dev.sql", "SQLite database path.")
	periodicity = flag.Int64("t", 5, "Periodicity expressed in seconds.")
)

type exchanger struct {
	name  string
	pairs map[string]string
	f     func(string) (*orderbook.OrderBook, error)
}

var exchangers = []*exchanger{
	&exchanger{"hitbtc", hitbtc.Pairs, hitbtc.OrderBook},
	&exchanger{"bitfinex", bitfinex.Pairs, bitfinex.OrderBook},
	&exchanger{"btce", btce.Pairs, btce.OrderBook},
	&exchanger{"kraken", kraken.Pairs, kraken.OrderBook},
	&exchanger{"cex", cex.Pairs, cex.OrderBook},
}

var pairs = []string{
	"BTC_USD",
	"BTC_EUR",
	"LTC_BTC",
}

func main() {
	log.Println("Start recording...")
	flag.Parse()

	db := database.Open(*dbPath)
	defer db.Close()

	for _, pair := range pairs {
		database.CreateTable(db, pair)
	}

	for {
		for _, e := range exchangers {
			for _, pair := range pairs {
				if _, ok := e.pairs[pair]; ok {
					go work(db, e, pair)
				}
			}
		}

		time.Sleep(time.Duration(*periodicity) * time.Second)
	}
}

func work(db *database.DB, e *exchanger, pair string) {
	defer logPanic()

	log.Printf("Fetching %s for pair %s...", e.name, pair)
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

	// created_at := start.Format("2006-1-2 15:04:05")
	// duration := int64(time.Since(start) / time.Microsecond)
	database.SaveRecord(db, pair, r)
}

// logPanic logs a formatted stack trace of the panicing goroutine. The stack trace is truncated
// at 4096 bytes (https://groups.google.com/d/topic/golang-nuts/JGraQ_Cp2Es/discussion)
func logPanic() {
	if err := recover(); err != nil {
		const size = 4096
		buf := make([]byte, size)
		stack := buf[:runtime.Stack(buf, false)]
		log.Printf("Error: %v\n%s", err, stack)
	}
}
