package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
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
	side        string
}

type Trade struct {
	tradeId     string
	price       float64
	quantity    float64
	fee         float64
	feeCurrency string
}

var getTradesFuncs = map[string]getTradesFunc{
	"Kraken":   getKrakenTrades,
	"Poloniex": getPoloniexTrades,
	"Hitbtc":   getHitbtcTrades,
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
		log.Printf("syncTrades: start sync of %s trade for arbId %s\n", ack.exchanger, ack.arbitrageId)

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

		err = saveTrades(db, ack, trades)
		if err != nil {
			log.Printf("syncTrades: saveTrades failed - %s\n", err)
			continue
		}

		// throttle queries (we're using the same API keys than the trader)
		log.Printf("syncTrades: completed sync of arbId %s\n", ack.arbitrageId)
		time.Sleep(10 * time.Second)
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

		trades = append(trades, &Trade{
			tradeId:     strconv.FormatInt(idFloat, 10),
			price:       price,
			quantity:    item["execQuantity"].(float64) * lotSize,
			fee:         fee,
			feeCurrency: ack.pair.Quote,
		})
	}

	return trades, nil
}

func getKrakenTrades(conf *Config, ack *OrderAck) ([]*Trade, error) {
	api := kraken.NewClient(conf.Kraken.Key, conf.Kraken.Secret)

	resp, err := api.OrdersInfo(ack.externalId, true)
	if err != nil {
		return nil, err
	}

	tradeIds := []string{}
	ids := resp[ack.externalId].(map[string]interface{})["trades"].([]interface{})
	for _, id := range ids {
		tradeIds = append(tradeIds, id.(string))
	}

	resp, err = api.TradesInfo(strings.Join(tradeIds, ","))
	if err != nil {
		return nil, err
	}

	trades := []*Trade{}
	for _, tradeId := range tradeIds {
		item := resp[tradeId].(map[string]interface{})

		price, err := strconv.ParseFloat(item["price"].(string), 64)
		if err != nil {
			return nil, fmt.Errorf("getKrakenTrades: parsing `price` failed - %s", err)
		}

		vol, err := strconv.ParseFloat(item["vol"].(string), 64)
		if err != nil {
			return nil, fmt.Errorf("getKrakenTrades: parsing `vol` failed - %s", err)
		}

		fee, err := strconv.ParseFloat(item["fee"].(string), 64)
		if err != nil {
			return nil, fmt.Errorf("getKrakenTrades: parsing `fee` failed - %s", err)
		}

		trades = append(trades, &Trade{
			tradeId:     tradeId,
			price:       price,
			quantity:    vol,
			fee:         fee,
			feeCurrency: ack.pair.Quote,
		})
	}

	return trades, nil
}

func getPoloniexTrades(conf *Config, ack *OrderAck) ([]*Trade, error) {
	api := poloniex.NewClient(conf.Poloniex.Key, conf.Poloniex.Secret)

	resp, err := api.OrderTrades(ack.externalId)
	if err != nil {
		return nil, fmt.Errorf("getPoloniexTrades: OrderTrades() failed - %s", err)
	}

	var feeCurrency string
	var feeAmount float64
	trades := []*Trade{}

	for _, item := range resp {
		id := item["tradeID"].(float64)
		tradeId := strconv.FormatFloat(id, 'f', 0, 64)

		rate, err := strconv.ParseFloat(item["rate"].(string), 64)
		if err != nil {
			return nil, fmt.Errorf("getPoloniexTrades: parsing `rate` failed - %s", err)
		}

		amount, err := strconv.ParseFloat(item["amount"].(string), 64)
		if err != nil {
			return nil, fmt.Errorf("getPoloniexTrades: parsing `amount` failed - %s", err)
		}

		feePercent, err := strconv.ParseFloat(item["fee"].(string), 64)
		if err != nil {
			return nil, fmt.Errorf("getPoloniexTrades: parsing `fee` failed - %s", err)
		}

		total, err := strconv.ParseFloat(item["total"].(string), 64)
		if err != nil {
			return nil, fmt.Errorf("getPoloniexTrades: parsing `fee` failed - %s", err)
		}

		// Not so sure about the fee computation.
		// TODO: only take X significant digit
		if ack.side == "buy" {
			feeCurrency = ack.pair.Base
			feeAmount = amount * feePercent
		} else {
			feeCurrency = ack.pair.Quote
			feeAmount = total * feePercent
		}

		trades = append(trades, &Trade{
			tradeId:     tradeId,
			price:       rate,
			quantity:    amount,
			fee:         feeAmount,
			feeCurrency: feeCurrency,
		})
	}

	return trades, nil
}

func getOrderAcks(db *sql.DB) ([]*OrderAck, error) {
	const sql = `
		select
			arbitrage_id,
			external_id,
			exchanger,
			pair,
			side
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

	var arbId, externalId, ex, pair, side string
	acks := []*OrderAck{}

	for rows.Next() {
		err = rows.Scan(&arbId, &externalId, &ex, &pair, &side)
		if err != nil {
			return nil, err
		}

		acks = append(acks, &OrderAck{arbId, externalId, ex, exchanger.NewPair(pair), side})
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return acks, nil
}

func saveTrades(db *sql.DB, ack *OrderAck, trades []*Trade) error {
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
		params := []interface{}{ack.arbitrageId, t.tradeId, t.price, t.quantity, ack.pair.String(), ack.side, t.fee, t.feeCurrency}
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
