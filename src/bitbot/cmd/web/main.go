package main

import (
	"encoding/json"
	"flag"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bitbot/database"
)

// TODO: HTTP handler are not panic safe
// TODO: cache static files
// TODO: make the static dir indepedent of the working directory

const (
	staticDir  = "public"
	timeFormat = "2006-01-02 15:04:05"
)

var db *database.DB

var (
	dbPath  = flag.String("d", "./data/dev.sql", "SQLite database path.")
	address = flag.String("h", "localhost:8080", "host:port TCP informations")
)

var pairs = map[string]bool{
	"BTC_USD": true,
	"BTC_EUR": true,
	"LTC_BTC": true,
}

func main() {
	flag.Parse()

	db = database.Open(*dbPath)
	defer db.Close()

	for pair, _ := range pairs {
		http.HandleFunc("/bid_ask/"+pair, BidAskHandler)
		http.HandleFunc("/opportunity/"+pair, OpportunityHandler)
	}

	http.Handle("/", http.FileServer(http.Dir(staticDir)))

	log.Printf("Starting webserver on %s\n", *address)
	err := http.ListenAndServe(*address, nil)
	if err != nil {
		panic(err)
	}
}

func BidAskHandler(w http.ResponseWriter, r *http.Request) {
	pair := parsePairFromURI(r.URL.Path)
	if _, ok := pairs[pair]; !ok {
		http.Error(w, "Page Not Found", http.StatusNotFound)
		return
	}

	records := database.SelectRecords(db, pair, 100)
	JSONResponse(w, records)
}

// TODO: default value for minProfitStr and limit are not correct
func OpportunityHandler(w http.ResponseWriter, r *http.Request) {
	pair := parsePairFromURI(r.URL.Path)
	if _, ok := pairs[pair]; !ok {
		http.Error(w, "Page Not Found", http.StatusNotFound)
		return
	}

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
	if limit == 0 {
		limit = 1000
	}

	records := map[string]*database.Record{}
	opps := []map[string]interface{}{}

	for r1 := range database.StreamRecords(db, pair, limit) {
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

func parsePairFromURI(path string) string {
	return strings.Split(path, "/")[2]
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
