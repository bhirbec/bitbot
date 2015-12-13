package main

import (
	"encoding/json"
	"flag"
	"log"
	"math"
	"net/http"
	"time"

	"bitbot/database"
)

var db *database.DB

var (
	dbPath = flag.String("d", "./data/dev.sql", "SQLite database path.")
)

func main() {
	db = database.Open(*dbPath)
	defer db.Close()

	http.HandleFunc("/bid_ask", BidAskHandler)
	http.HandleFunc("/opportunity", OpportunityHandler)

	// TODO: cache static files
	// TODO: make the static dir indepedent of the working directory
	// TODO: make static dir a flag or init parameter
	http.Handle("/", http.FileServer(http.Dir("client")))

	log.Println("Starting webserver")
	err := http.ListenAndServe("0.0.0.0:8080", nil)
	if err != nil {
		panic(err)
	}
}

func BidAskHandler(w http.ResponseWriter, r *http.Request) {
	records := database.SelectRecords(db, "BTC_USD")
	JSONResponse(w, records)
}

func OpportunityHandler(w http.ResponseWriter, r *http.Request) {
	records := map[string]*database.Record{}
	opps := []map[string]interface{}{}

	for r1 := range database.StreamRecords(db, "BTC_USD") {
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

			// if spread < 0.02 || vol < 0.1 {
			// 	continue
			// }

			const timeFormat = "2006-01-02 15:04:05.000"
			date := r1.StartDate.Format(timeFormat)
			profit := vol * (bid.Price - ask.Price)

			opp := map[string]interface{}{
				"Date":          date,
				"Ask":           ask,
				"BuyExchanger":  buy.Exchanger,
				"Bid":           bid,
				"SellExchanger": sell.Exchanger,
				"Profit":        profit,
				"Spread":        100 * spread,
			}

			opps = append(opps, opp)
		}
		// fmt.Printf("%s | %6.2f%% %6.2f$ | buy %-10s %6.2f %6.2f | sell %-10s %6.2f %6.2f\n", date, 100*spread, pAndl, buy.Exchanger, ask.Price, ask.Volume, sell.Exchanger, bid.Price, bid.Volume)
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
