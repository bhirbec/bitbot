package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"bitbot/exchanger/orderbook"
)

const timeFormat = "2006-1-2 15:04:05.000"

// TODO: remove this struct?
type DB struct {
	*sql.DB
}

type Record struct {
	StartDate  time.Time
	Orderbooks map[string]*orderbook.OrderBook
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
	_, err = db.Exec(fmt.Sprintf(stmt, pair), ts, obsJSON)
	panicOnError(err)
}

func SelectRecords(db *DB, pair string, limit int64) []*Record {
	stmt := `
        select
            ts,
            json_object(
                "Exchangers", orderbooks->'$.*.Exchanger',
                "Bids", orderbooks->'$.*.Bids[0]',
                "Asks", orderbooks->'$.*.Asks[0]'
            )
        from
            %s
        order by
            ts desc
        limit
            %d
    `
	records := []*Record{}
	rows, err := db.Query(fmt.Sprintf(stmt, pair, limit))
	panicOnError(err)

	var ts string
	var jsonData []byte

	for rows.Next() {
		err = rows.Scan(&ts, &jsonData)
		panicOnError(err)

		startDate, err := time.Parse(timeFormat, ts)
		panicOnError(err)

		var dest struct {
			Exchangers []string
			Bids       []*orderbook.Order
			Asks       []*orderbook.Order
		}
		err = json.Unmarshal(jsonData, &dest)
		panicOnError(err)

		obs := map[string]*orderbook.OrderBook{}

		for i, ex := range dest.Exchangers {
			obs[ex] = &orderbook.OrderBook{
				Exchanger: ex,
				Bids:      []*orderbook.Order{dest.Bids[i]},
				Asks:      []*orderbook.Order{dest.Asks[i]},
			}
		}

		records = append(records, &Record{
			StartDate:  startDate,
			Orderbooks: obs,
		})
	}

	return records
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
