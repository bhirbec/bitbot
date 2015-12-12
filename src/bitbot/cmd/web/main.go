package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"

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

	log.Println("Starting webserver")
	err := http.ListenAndServe("0.0.0.0:8080", nil)
	if err != nil {
		panic(err)
	}
}

func BidAskHandler(w http.ResponseWriter, r *http.Request) {
	records := database.SelectRecords(db, "BTC_USD")

	out, err := json.Marshal(records)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}
