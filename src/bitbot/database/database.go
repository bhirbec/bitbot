package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"time"

	"bitbot/exchanger/orderbook"
)

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
	panicOnError(err)
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
	panicOnError(err)
}

func SaveRecord(db *DB, pair string, r *Record) {
	bids, err := json.Marshal(r.Bids[:10])
	panicOnError(err)

	asks, err := json.Marshal(r.Asks[:10])
	panicOnError(err)

	const stmt = "insert into %s (StartTime, EndTime, Exchanger, Bids, Asks) values (?, ?, ?, ?, ?);"
	_, err = db.Exec(fmt.Sprintf(stmt, pair), r.StartTime, r.EndTime, r.Exchanger, bids, asks)
	panicOnError(err)
}

func SelectRecords(db *DB, pair string) chan *Record {
	c := make(chan *Record)

	const stmt = `
        select
            StartTime,
            EndTime,
            Exchanger,
            Bids,
            Asks
        from
            %s
        order
            by EndTime
    `

	go func() {
		defer close(c)

		rows, err := db.Query(fmt.Sprintf(stmt, pair))
		panicOnError(err)
		defer rows.Close()

		for rows.Next() {
			var startTime, endTime int64
			var exchanger, bidData, askData string

			err = rows.Scan(&startTime, &endTime, &exchanger, &bidData, &askData)
			panicOnError(err)

			var bids []*orderbook.Order
			err = json.Unmarshal([]byte(bidData), &bids)
			panicOnError(err)

			var asks []*orderbook.Order
			err = json.Unmarshal([]byte(askData), &asks)
			panicOnError(err)

			c <- &Record{
				Exchanger: exchanger,
				StartTime: startTime,
				StartDate: time.Unix(0, startTime),
				EndTime:   endTime,
				EndDate:   time.Unix(0, endTime),
				Bids:      bids,
				Asks:      asks,
			}
		}
	}()

	return c
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
