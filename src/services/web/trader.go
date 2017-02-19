package main

import (
	"fmt"

	"bitbot/errorutils"

	"github.com/jmoiron/sqlx"
)

func tradedArbitrages(db *sqlx.DB, limit int) interface{} {
	const stmt = `
        select
            arbitrage_id,
            buy_ex,
            sell_ex,
            pair,
            ts,
            buy_price,
            sell_price,
            vol,
            spread
        from
            arbitrage
        order by
            ts desc
        limit
            %d
    `
	var rows []struct {
		ArbitrageId string  `db:"arbitrage_id"`
		BuyEx       string  `db:"buy_ex"`
		SellEx      string  `db:"sell_ex"`
		Pair        string  `db:"pair"`
		Date        string  `db:"ts"`
		BuyPrice    float64 `db:"buy_price"`
		SellPrice   float64 `db:"sell_price"`
		Vol         float64 `db:"vol"`
		Spread      float64 `db:"spread"`
	}

	err := db.Select(&rows, fmt.Sprintf(stmt, limit))
	errorutils.PanicOnError(err)
	return rows
}

func trades(db *sqlx.DB, limit int) interface{} {
	const stmt = `
        select
		    arbitrage_id,
		    trade_id,
		    price,
		    quantity,
		    pair,
		    side,
		    fee,
		    fee_currency
        from
            trade
        order by
            arbitrage_id desc
        limit
            %d
    `
	var rows []struct {
		ArbitrageId string  `db:"arbitrage_id"`
		TradeId     string  `db:"trade_id"`
		Price       float64 `db:"price"`
		Quantity    float64 `db:"quantity"`
		Pair        string  `db:"pair"`
		Side        string  `db:"side"`
		Fee         float64 `db:"fee"`
		FeeCurrency string  `db:"fee_currency"`
	}

	err := db.Select(&rows, fmt.Sprintf(stmt, limit))
	errorutils.PanicOnError(err)
	return rows
}
