package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"time"

	"bitbot/exchanger/bitfinex"
	"bitbot/exchanger/btce"
	"bitbot/exchanger/cex"
	"bitbot/exchanger/hitbtc"
	"bitbot/exchanger/kraken"
	"bitbot/exchanger/orderbook"
)

type exchanger struct {
	name string
	f    func(string) (*orderbook.OrderBook, error)
}

type records map[string]*record

type record struct {
	Exchanger string
	StartTime int64
	EndTime   int64
	Bids      []*orderbook.Order
	Asks      []*orderbook.Order
}

func main() {
	log.Println("Fetching orderbooks...")

	flag.Parse()
	dbPath := flag.Arg(0)
	pair := "BTC_USD"

	db := openDb(dbPath)
	defer db.Close()

	exchangers := []*exchanger{
		&exchanger{"hitbtc", hitbtc.OrderBook},
		&exchanger{"bitfinex", bitfinex.OrderBook},
		&exchanger{"btce", btce.OrderBook},
		&exchanger{"kraken", kraken.OrderBook},
		&exchanger{"cex", cex.OrderBook},
	}

	createTable(db, pair)

	for {
		for _, e := range exchangers {
			// TODO: timeout after 2 sec
			go func(e *exchanger) {
				if r := fetchRecord(e, pair); r != nil {
					saveRecord(db, pair, r)
				}
			}(e)
		}

		time.Sleep(2 * time.Second)
	}
}

func createTable(db *sql.DB, pair string) {
	const stmt = `
		create table if not exists %s (
			StartTime int,
			EndTime int,
			Exchanger text,
			Bids text,
			Asks text
		)
	`
	_, err := db.Exec(fmt.Sprintf(stmt, pair))
	if err != nil {
		log.Fatal(err)
	}
}

func openDb(dbPath string) *sql.DB {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func fetchRecord(e *exchanger, pair string) *record {
	start := time.Now().UnixNano()
	book, err := e.f(pair)
	end := time.Now().UnixNano()

	if err != nil {
		log.Println(err)
		return nil
	}

	return &record{
		Exchanger: e.name,
		StartTime: start,
		EndTime:   end,
		Bids:      book.Bids,
		Asks:      book.Asks,
	}
}

func saveRecord(db *sql.DB, pair string, r *record) {
	bids, err := json.Marshal(r.Bids[:10])
	if err != nil {
		log.Fatal(err)
	}

	asks, err := json.Marshal(r.Asks[:10])
	if err != nil {
		log.Fatal(err)
	}

	const stmt = "insert into %s (StartTime, EndTime, Exchanger, Bids, Asks) values (?, ?, ?, ?, ?);"
	_, err = db.Exec(fmt.Sprintf(stmt, pair), r.StartTime, r.EndTime, r.Exchanger, bids, asks)
	if err != nil {
		log.Fatal(err)
	}
}
