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
                "Exchangers", orderbooks->"$.*.Exchanger",
                "BidPrices",  orderbooks->"$.*.Bids[0].Price",
                "BidVolumes", orderbooks->"$.*.Bids[0].Volume",
                "AskPrices",  orderbooks->"$.*.Asks[0].Price",
                "AskVolumes", orderbooks->"$.*.Asks[0].Volume"
            )
        from
            %s
        order by
            ts desc
        limit
            %d
    `
	fmt.Println(stmt)
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
			BidPrices  []float64
			BidVolumes []float64
			AskPrices  []float64
			AskVolumes []float64
		}
		err = json.Unmarshal(jsonData, &dest)
		panicOnError(err)

		obs := map[string]*orderbook.OrderBook{}

		// TODO: try to simplify/remove this
		for i, ex := range dest.Exchangers {
			obs[ex] = &orderbook.OrderBook{
				Exchanger: ex,
				Bids: []*orderbook.Order{
					&orderbook.Order{dest.BidPrices[i], dest.BidVolumes[i], 0},
				},
				Asks: []*orderbook.Order{
					&orderbook.Order{dest.AskPrices[i], dest.AskVolumes[i], 0},
				},
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
