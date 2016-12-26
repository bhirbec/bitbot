package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	"bitbot/errorutils"
	"bitbot/exchanger"
	"bitbot/exchanger/hitbtc"
	"bitbot/exchanger/kraken"
	"bitbot/exchanger/poloniex"
)

type getTradesFunc func(*Config, *OrderAck) ([]*Trade, error)

type OrderAck struct {
	arbitrageId string
	externalId  string
	exchanger   string
	pair        exchanger.Pair
}

type Trade struct {
	tradeId     string
	price       float64
	quantity    float64
	side        string
	pair        exchanger.Pair
	fee         float64
	feeCurrency string
}

var getTradesFuncs = map[string]getTradesFunc{
	// "Kraken":   getKrakenTrade,
	// "Poloniex": getPoloniexTrade,
	"Hitbtc": getHitbtcTrades,
}

func startSyncTrades(conf *Config) {
	for {
		syncTrades(conf)
		time.Sleep(10 * time.Minute)
	}
}

func syncTrades(conf *Config) {
	defer errorutils.LogPanic()

	db, err := OpenMysql()
	if err != nil {
		log.Printf("syncTrades: cannot open db %s\n", err)
		return
	}

	acks, err := getOrderAcks(db)
	if err != nil {
		log.Printf("syncTrades: getOrderAcks() failed - %s\n", err)
		return
	}

	for _, ack := range acks {
		log.Printf("syncTrades: start sync of arbId %s\n", ack.arbitrageId)

		f, ok := getTradesFuncs[ack.exchanger]
		if !ok {
			log.Printf("syncTrades: Missing `getTradesFunc` for %s\n", ack.exchanger)
			continue
		}

		trades, err := f(conf, ack)
		if err != nil {
			log.Printf("syncTrades: call getTradesFunc() failed for %s - %s\n", ack.exchanger, err)
			continue
		}

		err = saveTrades(db, ack.arbitrageId, trades)
		if err != nil {
			log.Printf("syncTrades: saveTrades failed - %s\n", err)
			continue
		}

		// throttle queries (we're using the same API keys than the trader)
		log.Printf("syncTrades: completed sync of arbId %s\n", ack.arbitrageId)
		time.Sleep(1 * time.Minute)
	}
}

func getHitbtcTrades(conf *Config, ack *OrderAck) ([]*Trade, error) {
	api := hitbtc.NewClient(conf.Hitbtc.Key, conf.Hitbtc.Secret)
	resp, err := api.TradesByOrder(ack.externalId)
	if err != nil {
		return nil, err
	}

	trades := []*Trade{}
	for _, item := range resp {
		idFloat := int64(item["tradeId"].(float64))

		lotSize, ok := hitbtc.LotSizes[ack.pair]
		if !ok {
			return nil, fmt.Errorf("getHitbtcTrades: Cannot find lot size for pair %s", ack.pair)
		}

		price, err := strconv.ParseFloat(item["execPrice"].(string), 64)
		if err != nil {
			return nil, fmt.Errorf("getHitbtcTrades: parsing `execPrice` failed - %s", err)
		}

		fee, err := strconv.ParseFloat(item["fee"].(string), 64)
		if err != nil {
			return nil, fmt.Errorf("getHitbtcTrades: parsing `fee` failed - %s", err)
		}

		trade := &Trade{
			tradeId:     strconv.FormatInt(idFloat, 10),
			price:       price,
			quantity:    item["execQuantity"].(float64) * lotSize,
			side:        item["side"].(string),
			pair:        ack.pair,
			fee:         fee,
			feeCurrency: ack.pair.Quote,
		}
		trades = append(trades, trade)
	}

	return trades, nil
}

func getKrakenTrade(conf *Config, id string) ([]*Trade, error) {
	_ = kraken.NewClient(conf.Kraken.Key, conf.Kraken.Secret)
	return nil, nil
}

func getPoloniexTrade(conf *Config, id string) ([]*Trade, error) {
	_ = poloniex.NewClient(conf.Poloniex.Key, conf.Poloniex.Secret)
	return nil, nil
}

func getOrderAcks(db *sql.DB) ([]*OrderAck, error) {
	const sql = `
		select
			arbitrage_id,
			external_id,
			exchanger,
			pair
		from
			order_ack
		where
			arbitrage_id not in (select arbitrage_id from trade)
	`

	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var arbId, externalId, ex, pair string
	acks := []*OrderAck{}

	for rows.Next() {
		err = rows.Scan(&arbId, &externalId, &ex, &pair)
		if err != nil {
			return nil, err
		}

		acks = append(acks, &OrderAck{arbId, externalId, ex, exchanger.NewPair(pair)})
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return acks, nil
}

func saveTrades(db *sql.DB, arbId string, trades []*Trade) error {
	const sql = `
		insert into trade
			(arbitrage_id, trade_id, price, quantity, pair, side, fee, fee_currency)
		values
			(?, ?, ?, ?, ?, ?, ?, ?)
	`

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("saveTrades: db.Begin() failed - %s\n", err)
	}

	for _, t := range trades {
		params := []interface{}{arbId, t.tradeId, t.price, t.quantity, t.pair.String(), t.side, t.fee, t.feeCurrency}
		_, err := tx.Exec(sql, params...)

		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("saveTrades: tx.Exec() failed - %s - %s\n", arbId, err)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("syncTrades: tx.Commit() failed - %s\n", err)
	}

	return nil
}
