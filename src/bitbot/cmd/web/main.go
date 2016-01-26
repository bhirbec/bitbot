package main

import (
	"encoding/json"
	"flag"
	"html/template"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"

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
	// TODO: factorize db related flags
	dbName  = flag.String("d", "bitbot", "MySQL database.")
	dbHost  = flag.String("h", "localhost", "MySQL host.")
	dbPort  = flag.String("p", "3306", "MySQL port.")
	dbUser  = flag.String("u", "bitbot", "MySQL user.")
	dbPwd   = flag.String("w", "password", "MySQL user's password.")
	address = flag.String("b", "localhost:8080", "host:port TCP informations")
)

var pairs = map[string]bool{
	"btc_usd": true,
	"btc_eur": true,
	"ltc_btc": true,
}

func main() {
	flag.Parse()

	db = database.Open(*dbName, *dbHost, *dbPort, *dbUser, *dbPwd)
	defer db.Close()

	for pair, _ := range pairs {
		http.HandleFunc("/bid_ask/"+pair, BidAskHandler)
		http.HandleFunc("/opportunity/"+pair, OpportunityHandler)
	}

	http.Handle("/public/", http.StripPrefix("/public", http.FileServer(http.Dir(staticDir))))
	http.HandleFunc("/", HomeHandler)

	log.Printf("Starting webserver on %s\n", *address)
	err := http.ListenAndServe(*address, nil)
	if err != nil {
		panic(err)
	}
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("client/index.html")
	if err != nil {
		panic(err)
	}
	t.Execute(w, nil)
}

func BidAskHandler(w http.ResponseWriter, r *http.Request) {
	pair := parsePairFromURI(r.URL.Path)
	if _, ok := pairs[pair]; !ok {
		http.Error(w, "Page Not Found", http.StatusNotFound)
		return
	}

	rows := []interface{}{}

	for _, rec := range database.SelectRecords(db, pair, 100) {
		for ex, ob := range rec.Orderbooks {
			rows = append(rows, map[string]interface{}{
				"Exchanger": ex,
				"StartDate": rec.StartDate,
				"Bids":      ob.Bids,
				"Asks":      ob.Asks,
			})
		}
	}

	JSONResponse(w, rows)
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

	opps := []map[string]interface{}{}

	for _, rec := range database.SelectRecords(db, pair, limit) {
		for ex1, buy := range rec.Orderbooks {
			for ex2, sell := range rec.Orderbooks {
				if ex1 == ex2 {
					continue
				}

				if buy.Asks[0].Price >= sell.Bids[0].Price {
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
					"Date":          rec.StartDate.Format(timeFormat),
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
