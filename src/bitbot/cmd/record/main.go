package main

import (
	"flag"
	"log"
	"runtime"
	"sync"
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
	dbName      = flag.String("d", "bitbot", "MySQL database.")
	dbHost      = flag.String("h", "localhost", "MySQL host.")
	dbPort      = flag.String("p", "3306", "MySQL port.")
	dbUser      = flag.String("u", "bitbot", "MySQL user.")
	dbPwd       = flag.String("w", "password", "MySQL user's password.")
	periodicity = flag.Int64("t", 10, "Periodicity expressed in seconds.")
)

type exchanger struct {
	name  string
	pairs map[string]string
	f     func(string) (*orderbook.OrderBook, error)
}

var exchangers = []*exchanger{
	&exchanger{hitbtc.ExchangerName, hitbtc.Pairs, hitbtc.OrderBook},
	&exchanger{bitfinex.ExchangerName, bitfinex.Pairs, bitfinex.OrderBook},
	&exchanger{btce.ExchangerName, btce.Pairs, btce.OrderBook},
	&exchanger{kraken.ExchangerName, kraken.Pairs, kraken.OrderBook},
	&exchanger{cex.ExchangerName, cex.Pairs, cex.OrderBook},
}

var pairs = []string{
	"btc_usd",
	"btc_eur",
	"ltc_btc",
}

func main() {
	log.Println("Start recording...")
	flag.Parse()

	db := database.Open(*dbName, *dbHost, *dbPort, *dbUser, *dbPwd)
	defer db.Close()

	for {
		start := time.Now()

		for _, pair := range pairs {
			go work(db, pair, start)
		}

		time.Sleep(time.Duration(*periodicity) * time.Second)
	}
}

func work(db *database.DB, pair string, start time.Time) {
	defer logPanic()

	var wg sync.WaitGroup
	obs := map[string]*orderbook.OrderBook{}

	for _, e := range exchangers {
		if _, ok := e.pairs[pair]; !ok {
			continue
		}

		wg.Add(1)

		go func(e *exchanger) {
			defer wg.Done()

			log.Printf("Fetching %s for pair %s...", e.name, pair)
			book, err := e.f(pair)
			// end := time.Now()
			// duration := int64(time.Since(start) / time.Microsecond)

			if err != nil {
				log.Println(err)
			} else {
				book.Bids = book.Bids[:10]
				book.Asks = book.Asks[:10]
				obs[e.name] = book
			}
		}(e)
	}

	wg.Wait()
	database.SaveRecord(db, pair, start, obs)
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
