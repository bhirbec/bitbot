package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"bitbot/errorutils"
	"bitbot/exchanger"
)

// TODO: can we do time formatting on the client instead?

const (
	timeFormat        = "2006-01-02 15:04:05.000"
	displayTimeFormat = "2006-01-02 15:04"
)

// TODO: remove this struct?
type DB struct {
	*sql.DB
}

func Open(name, host, port, user, pwd string) *DB {
	source := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8", user, pwd, host, port, name)
	db, err := sql.Open("mysql", source)
	errorutils.PanicOnError(err)
	return &DB{db}
}

func SaveOrderbooks(db *DB, pair string, start time.Time, obs []*exchanger.OrderBook) {
	placeholders := []string{}
	params := []interface{}{}

	var min = func(a, b int) int {
		if a < b {
			return a
		} else {
			return b
		}
	}

	for _, ob := range obs {
		const limit = 10
		n := min(len(ob.Bids), limit)
		bids, err := json.Marshal(ob.Bids[:n])
		errorutils.PanicOnError(err)

		n = min(len(ob.Asks), limit)
		asks, err := json.Marshal(ob.Asks[:n])
		errorutils.PanicOnError(err)

		params = append(params, start)
		params = append(params, pair)
		params = append(params, ob.Exchanger)
		params = append(params, bids)
		params = append(params, asks)
		placeholders = append(placeholders, "(?, ?, ?, ?, ?)")
	}

	stmt := "insert into orderbooks (ts, pair, exchanger, bids, asks) values " + strings.Join(placeholders, ",")
	_, err := db.Exec(stmt, params...)
	errorutils.PanicOnError(err)
}

func SelectBidAsk(db *DB, pair string, limit int64) []map[string]interface{} {
	const stmt = `
        select
            ts,
            exchanger,
            bids->'$[0].Price',
            asks->'$[0].Price',
            bids->'$[0].Volume',
            asks->'$[0].Volume'
        from
            orderbooks
        where
            pair = ?
        order by
            ts desc
        limit
            %d
    `

	rows, err := db.Query(fmt.Sprintf(stmt, limit), pair)
	errorutils.PanicOnError(err)
	defer rows.Close()

	var ts, ex string
	var bidPrice, bidVol, askPrice, askVol float64
	output := []map[string]interface{}{}

	for rows.Next() {
		err = rows.Scan(&ts, &ex, &bidPrice, &askPrice, &bidVol, &askVol)
		errorutils.PanicOnError(err)

		date, err := time.Parse(timeFormat, ts)
		errorutils.PanicOnError(err)

		output = append(output, map[string]interface{}{
			"Exchanger": ex,
			"Date":      date.Format(displayTimeFormat),
			"BidPrice":  bidPrice,
			"AskPrice":  askPrice,
			"BidVol":    bidVol,
			"AskVol":    askVol,
		})
	}

	err = rows.Err()
	errorutils.PanicOnError(err)

	return output
}

func ComputeAndSaveArbitrage(db *DB, pair string, start time.Time, obs []*exchanger.OrderBook) {
	placeholders := []string{}
	params := []interface{}{}

	for _, buyOb := range obs {
		for _, sellOb := range obs {
			if buyOb.Exchanger == sellOb.Exchanger {
				continue
			}

			buyOrder := buyOb.Asks[0]
			sellOrder := sellOb.Bids[0]

			if buyOrder.Price >= sellOrder.Price {
				continue
			}

			vol := math.Min(buyOrder.Volume, sellOrder.Volume)
			spread := 100 * (sellOrder.Price/buyOrder.Price - 1)

			params = append(params, buyOb.Exchanger)
			params = append(params, sellOb.Exchanger)
			params = append(params, pair)
			params = append(params, start)
			params = append(params, buyOrder.Price)
			params = append(params, sellOrder.Price)
			params = append(params, vol)
			params = append(params, spread)
			placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?, ?, ?)")
		}
	}

	if len(params) == 0 {
		return
	}

	stmt := "insert into arbitrages (buy_ex, sell_ex, pair, ts, buy_price, sell_price, vol, spread) values " + strings.Join(placeholders, ",")
	_, err := db.Exec(stmt, params...)
	errorutils.PanicOnError(err)
}

func SelectArbitrages(db *DB, pair, buyExchanger, sellExchanger string, minProfit, minVol float64, limit int64) []map[string]interface{} {
	const stmt = `
        select
            buy_ex,
            sell_ex,
            ts,
            buy_price,
            sell_price,
            vol,
            spread
        from
            arbitrages
        where
            pair = ?
            and (? = '' or buy_ex = ?)
            and (? = '' or sell_ex = ?)
            and spread >= ?
            and vol >= ?
        order by
            ts desc
        limit
            %d
    `

	rows, err := db.Query(fmt.Sprintf(stmt, limit), pair, buyExchanger, buyExchanger, sellExchanger, sellExchanger, minProfit, minVol)
	errorutils.PanicOnError(err)
	defer rows.Close()

	var buyEx, sellEx, ts string
	var buyPrice, sellPrice, vol, spread float64
	output := []map[string]interface{}{}

	for rows.Next() {
		err = rows.Scan(&buyEx, &sellEx, &ts, &buyPrice, &sellPrice, &vol, &spread)
		errorutils.PanicOnError(err)

		date, err := time.Parse(timeFormat, ts)
		errorutils.PanicOnError(err)

		output = append(output, map[string]interface{}{
			"Date":          date.Format(displayTimeFormat),
			"BuyPrice":      buyPrice,
			"BuyExchanger":  buyEx,
			"SellPrice":     sellPrice,
			"SellExchanger": sellEx,
			"Volume":        vol,
			"Spread":        spread,
		})
	}

	err = rows.Err()
	errorutils.PanicOnError(err)

	return output
}
