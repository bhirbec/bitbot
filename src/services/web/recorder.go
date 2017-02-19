package main

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"bitbot/errorutils"
)

// TODO: can we do time formatting on the client instead?
const (
	timeFormat        = "2006-01-02 15:04:05.000"
	displayTimeFormat = "2006-01-02 15:04"
)

func recordedArbitrages(db *sqlx.DB, pair, buyExchanger, sellExchanger string, minProfit, minVol float64, limit int64) interface{} {
	const stmt = `
        select
            ts,
            buy_ex,
            sell_ex,
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

	var rows []*struct {
		Date          string  `db:"ts"`
		BuyPrice      float64 `db:"buy_price"`
		BuyExchanger  string  `db:"buy_ex"`
		SellPrice     float64 `db:"sell_price"`
		SellExchanger string  `db:"sell_ex"`
		Volume        float64 `db:"vol"`
		Spread        float64 `db:"spread"`
	}

	sql := fmt.Sprintf(stmt, limit)
	err := db.Select(&rows, sql, pair, buyExchanger, buyExchanger, sellExchanger, sellExchanger, minProfit, minVol)
	errorutils.PanicOnError(err)

	for _, row := range rows {
		date, err := time.Parse(timeFormat, row.Date)
		errorutils.PanicOnError(err)
		row.Date = date.Format(displayTimeFormat)
	}

	return rows
}

func recordedBidAsk(db *sqlx.DB, pair string, limit int64) interface{} {
	const stmt = `
        select
            ts,
            exchanger,
            bids->'$[0].Price' as bid_price,
            asks->'$[0].Price' as ask_price,
            bids->'$[0].Volume' as bid_vol,
            asks->'$[0].Volume' as ask_vol
        from
            orderbooks
        where
            pair = ?
        order by
            ts desc
        limit
            %d
    `

	var rows []*struct {
		Date      string  `db:"ts"`
		Exchanger string  `db:"exchanger"`
		BidPrice  float64 `db:"bid_price"`
		AskPrice  float64 `db:"ask_price"`
		BidVol    float64 `db:"bid_vol"`
		AskVol    float64 `db:"ask_vol"`
	}

	err := db.Select(&rows, fmt.Sprintf(stmt, limit), pair)
	errorutils.PanicOnError(err)

	for _, row := range rows {
		date, err := time.Parse(timeFormat, row.Date)
		errorutils.PanicOnError(err)
		row.Date = date.Format(displayTimeFormat)
	}

	return rows
}
