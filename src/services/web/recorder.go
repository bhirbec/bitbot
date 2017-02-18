package main

import (
	"fmt"
	"time"

	"bitbot/database"
	"bitbot/errorutils"
)

// TODO: can we do time formatting on the client instead?
const (
	timeFormat        = "2006-01-02 15:04:05.000"
	displayTimeFormat = "2006-01-02 15:04"
)

func recordedArbitrages(db *database.DB, pair, buyExchanger, sellExchanger string, minProfit, minVol float64, limit int64) []map[string]interface{} {
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

func recordedBidAsk(db *database.DB, pair string, limit int64) []map[string]interface{} {
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
