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
            t.real_buy_price,
            t.real_sell_price,
            t.real_buy_vol,
            t.real_sell_vol,
            100 * (t.real_sell_price / t.real_buy_price - 1) as real_spread
        from
            arbitrage a
            left join (
                select
                    arbitrage_id,
                    sum(case when side = 'buy' then price * quantity else null end) /
                    sum(case when side = 'buy' then quantity else null end) as real_buy_price,

                    sum(case when side = 'sell' then price * quantity else null end) /
                    sum(case when side = 'sell' then quantity else null end) as real_sell_price,

                    sum(case when side = 'buy' then quantity else null end) as real_buy_vol,
                    sum(case when side = 'sell' then quantity else null end) as real_sell_vol
                from
                    trade
                group by
                    arbitrage_id
            ) as t using(arbitrage_id)
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
		RealBuyVol    *float64 `db:"real_buy_vol"`
		RealSellVol   *float64 `db:"real_sell_vol"`
		RealSpread    *float64 `db:"real_spread"`
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
