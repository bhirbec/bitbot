package main

import (
	"encoding/json"
	"flag"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"bitbot/database"
)

const (
	staticDir  = "public"
	timeFormat = "2006-01-02 15:04:05"
)

var db *database.DB

var (
	dbPath  = flag.String("d", "./data/dev.sql", "SQLite database path.")
	address = flag.String("h", "localhost:8080", "host:port TCP informations")
)

func main() {
	flag.Parse()

	db = database.Open(*dbPath)
	defer db.Close()

	http.HandleFunc("/bid_ask", BidAskHandler)
	http.HandleFunc("/opportunity", OpportunityHandler)

	// TODO: cache static files
	// TODO: make the static dir indepedent of the working directory
	http.Handle("/", http.FileServer(http.Dir(staticDir)))

	log.Printf("Starting webserver on %s\n", *address)
	err := http.ListenAndServe(*address, nil)
	if err != nil {
		panic(err)
	}
}

func BidAskHandler(w http.ResponseWriter, r *http.Request) {
	records := database.SelectRecords(db, "BTC_USD", 100)
	JSONResponse(w, records)
}

func OpportunityHandler(w http.ResponseWriter, r *http.Request) {
	minProfitStr := r.FormValue("min_profit")
	if minProfitStr == "" {
		minProfitStr = "0"
	}

	minProfit, err := strconv.ParseFloat(minProfitStr, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	limitParam := r.FormValue("limit")
	limit, _ := strconv.ParseInt(limitParam, 10, 64)

	records := map[string]*database.Record{}
	opps := []map[string]interface{}{}

	for r1 := range database.StreamRecords(db, "BTC_USD", limit) {
		records[r1.Exchanger] = r1

		for ex, r2 := range records {
			if ex == r1.Exchanger {
				continue
			}

			// skip record that are not on the same time range
			if r1.StartDate.Sub(r2.StartDate) > time.Duration(1*time.Second) {
				continue
			}

			var buy, sell *database.Record

			if r1.Asks[0].Price < r2.Bids[0].Price {
				buy, sell = r1, r2
			} else if r2.Asks[0].Price < r1.Bids[0].Price {
				buy, sell = r2, r1
			} else {
				continue
			}

			ask := buy.Asks[0]
			bid := sell.Bids[0]
			vol := math.Min(ask.Volume, bid.Volume)
			spread := bid.Price/ask.Price - 1

			if spread < minProfit {
				continue
			}

			opp := map[string]interface{}{
				"Date":          r1.StartDate.Format(timeFormat),
				"Ask":           ask,
				"BuyExchanger":  buy.Exchanger,
				"Bid":           bid,
				"SellExchanger": sell.Exchanger,
				"Profit":        vol * (bid.Price - ask.Price),
				"Spread":        100 * spread,
			}

			opps = append(opps, opp)
		}
	}

	JSONResponse(w, opps)
}

func JSONResponse(w http.ResponseWriter, input interface{}) {
	out, err := json.Marshal(input)
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
