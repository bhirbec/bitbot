package main

import (
	"flag"
	"log"
	"strings"
	"time"

	"github.com/toorop/go-bittrex"

	"bitbot/database"
)

var (
	configPath = flag.String("config", "src/services/bittrex_trader/config.json", "JSON file that stores bittrex credentials.")
)

func main() {
	log.Println("Start Bittrex trader...")
	flag.Parse()

	config, err := loadConfig(*configPath)
	if err != nil {
		log.Panic(err)
	}

	// Bittrex client
	bittrex := bittrex.New(config.Bittrex.Key, config.Bittrex.Secret)

	// open db
	creds := config.Mysql
	db := database.Open(creds.Db, creds.Host, creds.Port, creds.User, creds.Pwd)
	defer db.Close()

	// Get markets
	for {
		log.Println("Fetching market summaries...")

		if summaries, err := bittrex.GetMarketSummaries(); err != nil {
			log.Println("ERROR: ", err)
		} else if err = saveMarketSummaries(db, summaries); err != nil {
			log.Println("ERROR: ", err)
		}

		time.Sleep(1 * time.Minute)
	}
}

func saveMarketSummaries(db *database.DB, summaries []bittrex.MarketSummary) error {
	placeholders := []string{}
	params := []interface{}{}

	for _, s := range summaries {
		params = append(params, s.MarketName)
		params = append(params, s.High)
		params = append(params, s.Low)
		params = append(params, s.Ask)
		params = append(params, s.Bid)
		params = append(params, s.OpenBuyOrders)
		params = append(params, s.OpenSellOrders)
		params = append(params, s.Volume)
		params = append(params, s.Last)
		params = append(params, s.BaseVolume)
		params = append(params, s.PrevDay)
		params = append(params, s.TimeStamp)
		placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	}

	// Bittrex returns the last known price with its timestamp. If the price
	// hasn't moved since the last fetch then we got a `duplicated key error`
	// on (timestamp, price). We use `ignore` to skip this.
	stmt := `
        insert ignore into market_summary
            (market_name, high, low, ask, bid, open_buy_orders, open_sell_orders,
            volume, last, base_volume, prev_day, creation_date)
        values
    ` + strings.Join(placeholders, ",")

	_, err := db.Exec(stmt, params...)
	return err
}
