package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"time"

	"bitbot/exchanger/orderbook"
)

// TODO: should panic instead of log.Fatal

type DB struct {
	*sql.DB
}

// TODO: factorize this type with Orderbook type?
// TODO: should only expose StartDate and EndDate (not StartTime and EndTime)
type Record struct {
	Exchanger string
	StartTime int64
	StartDate time.Time
	EndTime   int64
	EndDate   time.Time
	Bids      []*orderbook.Order
	Asks      []*orderbook.Order
}

func Open(dbPath string) *DB {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	return &DB{db}
}

func CreateTable(db *DB, pair string) {
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

func SaveRecord(db *DB, pair string, r *Record) {
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
