package main

import (
	"flag"
	"log"
	"sync"
	"time"

	"bitbot/database"
	"bitbot/errorutils"
	"bitbot/exchanger"

	"bitbot/exchanger/bitfinex"
	"bitbot/exchanger/btce"
	"bitbot/exchanger/cex"
	"bitbot/exchanger/gemini"
	"bitbot/exchanger/hitbtc"
	"bitbot/exchanger/kraken"
	"bitbot/exchanger/poloniex"
	"bitbot/exchanger/therocktrading"
)

// NOTE: use Ticker endpoint to retrieve bid/ask info for several pairs at the same time?
// - https://api.kraken.com/0/public/Ticker?pair=XXBTZUSD,XXBTZEUR,XXBTXLTC
// - https://cex.io/api/tickers/BTC/BTC/USD/EUR
// - https://api.hitbtc.com/api/1/public/ticker

var (
	dbName      = flag.String("d", "bitbot", "MySQL database.")
	dbHost      = flag.String("h", "localhost", "MySQL host.")
	dbPort      = flag.String("p", "3306", "MySQL port.")
	dbUser      = flag.String("u", "bitbot", "MySQL user.")
	dbPwd       = flag.String("w", "password", "MySQL user's password.")
	periodicity = flag.Int64("t", 10, "Wait t seconds between each pair.")
)

type Exchanger struct {
	name  string
	pairs map[exchanger.Pair]string
	f     func(exchanger.Pair) (*exchanger.OrderBook, error)
}

var exchangers = []*Exchanger{
	&Exchanger{bitfinex.ExchangerName, bitfinex.Pairs, bitfinex.OrderBook},
	&Exchanger{btce.ExchangerName, btce.Pairs, btce.OrderBook},
	&Exchanger{kraken.ExchangerName, kraken.Pairs, kraken.OrderBook},
	&Exchanger{cex.ExchangerName, cex.Pairs, cex.OrderBook},
	&Exchanger{gemini.ExchangerName, gemini.Pairs, gemini.OrderBook},
	&Exchanger{hitbtc.ExchangerName, hitbtc.Pairs, hitbtc.OrderBook},
	&Exchanger{poloniex.ExchangerName, poloniex.Pairs, poloniex.OrderBook},
	&Exchanger{therocktrading.ExchangerName, therocktrading.Pairs, therocktrading.OrderBook},
}

var pairs = []exchanger.Pair{
	exchanger.BTC_USD,
	exchanger.BTC_EUR,
	exchanger.LTC_BTC,
	exchanger.ETH_BTC,
	exchanger.ETC_BTC,
	exchanger.ZEC_BTC,
}

func main() {
	log.Println("Start recording...")
	flag.Parse()

	db := database.Open(*dbName, *dbHost, *dbPort, *dbUser, *dbPwd)
	defer db.Close()

	for {
		for _, pair := range pairs {
			go work(db, pair)
			time.Sleep(time.Duration(*periodicity) * time.Second)
		}
	}
}

func work(db *database.DB, pair exchanger.Pair) {
	defer errorutils.LogPanic()

	var wg sync.WaitGroup
	obs := []*exchanger.OrderBook{}
	start := time.Now()

	var _work = func(pair exchanger.Pair, e *Exchanger) {
		defer wg.Done()

		log.Printf("Fetching %s for pair %s...", e.name, pair)
		book, err := e.f(pair)
		// end := time.Now()
		// duration := int64(time.Since(start) / time.Microsecond)

		if err != nil {
			log.Println(err)
			return
		}

		obs = append(obs, book)
	}

	for _, e := range exchangers {
		if _, ok := e.pairs[pair]; ok {
			wg.Add(1)
			go _work(pair, e)
		}
	}

	wg.Wait()

	if len(obs) == 0 {
		return
	}

	database.SaveOrderbooks(db, pair, start, obs)
	database.ComputeAndSaveArbitrage(db, pair, start, obs)
}
