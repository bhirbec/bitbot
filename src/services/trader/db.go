package main

import (
	"database/sql"
	"flag"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

var (
	dbName = flag.String("db-name", "bitbot", "MySQL database.")
	dbHost = flag.String("db-host", "localhost", "MySQL host.")
	dbPort = flag.String("db-port", "3306", "MySQL port.")
	dbUser = flag.String("db-user", "bitbot", "MySQL user.")
	dbPwd  = flag.String("db-password", "password", "MySQL user's password.")
)

func OpenMysql() (*sql.DB, error) {
	source := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8", *dbUser, *dbPwd, *dbHost, *dbPort, *dbName)
	return sql.Open("mysql", source)
}

func saveArbitrage(db *sql.DB, arb *arbitrage) error {
	params := []interface{}{}
	params = append(params, arb.id)
	params = append(params, arb.buyEx.Exchanger)
	params = append(params, arb.sellEx.Exchanger)
	params = append(params, arb.pair.String())
	params = append(params, arb.ts)
	params = append(params, arb.buyEx.Asks[0].Price)
	params = append(params, arb.sellEx.Bids[0].Price)
	params = append(params, arb.vol)
	params = append(params, arb.spread)

	const stmt = `
		insert into arbitrage
			(arbitrage_id, buy_ex, sell_ex, pair, ts, buy_price, sell_price, vol, spread)
		values
			(?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := db.Exec(stmt, params...)
	return err
}

func saveOrderAck(db *sql.DB, arbId, externalId, pair, ex, side string) error {
	params := []interface{}{arbId, externalId, pair, ex, side}
	const stmt = "insert into order_ack (arbirage_id, external_id, pair, exchanger, side) values (?, ?, ?, ?, ?)"
	_, err := db.Exec(stmt, params...)
	return err
}
