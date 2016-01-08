package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"bitbot/exchanger/orderbook"
)

const timeFormat = "2006-1-2 15:04:05.000"

// TODO: remove this struct?
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

func Open(name, host, port, user, pwd string) *DB {
	source := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8", user, pwd, host, port, name)
	db, err := sql.Open("mysql", source)
	panicOnError(err)
	return &DB{db}
}

func SaveRecord(db *DB, pair string, start time.Time, obs map[string]*orderbook.OrderBook) {
	obsJSON, err := json.Marshal(obs)
	panicOnError(err)

	const stmt = "insert into %s (ts, orderbooks) values (?, ?)"
	ts := start.Format(timeFormat)
	pair = strings.ToLower(pair)
	_, err = db.Exec(fmt.Sprintf(stmt, pair), ts, obsJSON)
	panicOnError(err)
}

func SelectRecords(db *DB, pair string, limit int64) []*Record {
	// TODO: exchanger must be a parameter
	const stmt = `
        select
            ts,
            orderbooks
        from
            %s
        order by
            ts desc
        limit
            %d
    `

	records := []*Record{}
	// TODO: think about the lowercase for pair
	pair = strings.ToLower(pair)
	rows, err := db.Query(fmt.Sprintf(stmt, pair, limit))
	panicOnError(err)

	var ts string
	var jsonData []byte
	var obs map[string]*orderbook.OrderBook

	for rows.Next() {
		err = rows.Scan(&ts, &jsonData)
		panicOnError(err)

		startDate, err := time.Parse(timeFormat, ts)
		panicOnError(err)

		err = json.Unmarshal(jsonData, &obs)
		panicOnError(err)

		for ex, ob := range obs {
			records = append(records, &Record{
				Exchanger: ex,
				StartDate: startDate,
				Bids:      ob.Bids,
				Asks:      ob.Asks,
			})
		}
	}

	return records
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
