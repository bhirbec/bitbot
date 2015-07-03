package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/szferi/gomdb"

	"bitbot/exchanger/bitfinex"
	"bitbot/exchanger/btce"
	"bitbot/exchanger/cex"
	"bitbot/exchanger/hitbtc"
	"bitbot/exchanger/kraken"
	"bitbot/exchanger/orderbook"
)

// TODO: this code is not panic safe

type exchanger struct {
	name string
	f    func(string) (*orderbook.OrderBook, error)
}

// TODO: factorize this type with Orderbook type?
type record struct {
	StartTime int64
	EndTime   int64
	Bids      []*orderbook.Order
	Asks      []*orderbook.Order
}

func main() {
	log.Println("Fetching orderbooks...")

	flag.Parse()
	path := flag.Arg(0)
	pair := "BTC_USD"

	exchangers := []*exchanger{
		&exchanger{"hitbtc", hitbtc.OrderBook},
		&exchanger{"bitfinex", bitfinex.OrderBook},
		&exchanger{"btce", btce.OrderBook},
		&exchanger{"kraken", kraken.OrderBook},
		&exchanger{"cex", cex.OrderBook},
	}

	env := NewEnv(path)

	for _, e := range exchangers {
		go func(e *exchanger) {
			recordBook(env, e, pair)
		}(e)
	}

	// wait indefinetely
	<-make(chan struct{})
}

func recordBook(env *mdb.Env, e *exchanger, pair string) {
	c := time.Tick(1 * time.Second)
	dbi := fmt.Sprintf("%s-%s", e.name, pair)

	for now := range c {
		log.Printf("Fetching %s %s\n", e.name, pair)
		start := now.UnixNano()

		book, err := e.f(pair)
		if err != nil {
			log.Println(err)
			continue
		}

		r := &record{
			StartTime: start,
			EndTime:   time.Now().UnixNano(),
			Bids:      book.Bids,
			Asks:      book.Asks,
		}

		write(env, dbi, r)
	}
}

func NewEnv(path string) *mdb.Env {
	env, err := mdb.NewEnv()
	panicOnError(err)

	env.SetMaxDBs(10)
	env.SetMapSize(1 << 22) // max file size
	env.Open(path, 0, 0664)
	return env
}

func write(env *mdb.Env, dbiName string, r *record) {
	txn, err := env.BeginTxn(nil, 0)
	panicOnError(err)

	dbi, err := txn.DBIOpen(&dbiName, 0x40000)
	panicOnError(err)

	defer env.DBIClose(dbi)

	key := fmt.Sprintf("%d", r.EndTime)
	val, err := json.Marshal(r)
	panicOnError(err)

	err = txn.Put(dbi, []byte(key), val, 0)
	panicOnError(err)

	err = txn.Commit()
	panicOnError(err)

	err = env.Sync(1)
	panicOnError(err)
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
