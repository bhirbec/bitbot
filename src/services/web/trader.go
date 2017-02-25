package main

import (
	"fmt"

	"bitbot/errorutils"

	"github.com/jmoiron/sqlx"
)

func tradedArbitrages(db *sqlx.DB, limit int) interface{} {
	const stmt = `
        select
            a.arbitrage_id,
            a.buy_ex,
            a.sell_ex,
            a.pair,
            a.ts,
            a.buy_price,
            a.sell_price,
            a.vol,
            a.spread,
            buy.price as real_buy_price,
            sell.price as real_sell_price
        from
            arbitrage a
            left join trade as buy on a.arbitrage_id = buy.arbitrage_id and buy.side = 'buy'
            left join trade as sell on a.arbitrage_id = sell.arbitrage_id and sell.side = 'sell'
        order by
            a.ts desc
        limit
            %d
    `
	var rows []*struct {
		ArbitrageId   string   `db:"arbitrage_id"`
		BuyEx         string   `db:"buy_ex"`
		SellEx        string   `db:"sell_ex"`
		Pair          string   `db:"pair"`
		Date          string   `db:"ts"`
		BuyPrice      float64  `db:"buy_price"`
		SellPrice     float64  `db:"sell_price"`
		Vol           float64  `db:"vol"`
		Spread        float64  `db:"spread"`
		RealBuyPrice  *float64 `db:"real_buy_price"`
		RealSellPrice *float64 `db:"real_sell_price"`
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
	var rows []*struct {
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
