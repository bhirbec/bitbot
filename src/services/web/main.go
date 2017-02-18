package main

import (
	"encoding/json"
	"flag"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bitbot/database"
	"bitbot/errorutils"
)

// TODO: HTTP handler are not panic safe
// TODO: cache static files
// TODO: make the static dir indepedent of the working directory

const (
	staticDir = "public"
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

// TODO: use exchanger.Pair?
var pairs = map[string]bool{
	"btc_usd": true,
	"btc_eur": true,
	"ltc_btc": true,
	"eth_btc": true,
	"etc_btc": true,
	"zec_btc": true,
}

func main() {
	flag.Parse()

	db = database.Open(*dbName, *dbHost, *dbPort, *dbUser, *dbPwd)
	defer db.Close()

	m := http.NewServeMux()

	for pair, _ := range pairs {
		m.HandleFunc("/bid_ask/"+pair, BidAskHandler)
		m.HandleFunc("/opportunity/"+pair, OpportunityHandler)
	}

	m.Handle("/public/", http.StripPrefix("/public", http.FileServer(http.Dir(staticDir))))
	m.HandleFunc("/", HomeHandler)

	log.Printf("Starting webserver on %s\n", *address)
	http.HandleFunc("/", timeItWrapper(m))
	err := http.ListenAndServe(*address, nil)
	errorutils.PanicOnError(err)
}

func timeItWrapper(h http.Handler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h.ServeHTTP(w, r)
		duration := time.Since(start) / time.Millisecond
		log.Printf("%s took %dms", r.URL.RequestURI(), duration)
	}
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("client/index.html")
	errorutils.PanicOnError(err)
	t.Execute(w, nil)
}

func BidAskHandler(w http.ResponseWriter, r *http.Request) {
	pair := parsePairFromURI(r.URL.Path)
	if _, ok := pairs[pair]; !ok {
		http.Error(w, "Page Not Found", http.StatusNotFound)
		return
	}

	rows := recordedBidAsk(db, pair, 100)
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
		limit = 100
	}

	minVolStr := r.FormValue("min_vol")
	if minVolStr == "" {
		minVolStr = "0"
	}

	minVol, err := strconv.ParseFloat(minVolStr, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// TODO: validate?
	buyExchanger := r.FormValue("buy_ex")
	sellExchanger := r.FormValue("sell_ex")

	rows := recordedArbitrages(db, pair, buyExchanger, sellExchanger, minProfit, minVol, limit)
	JSONResponse(w, rows)
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
