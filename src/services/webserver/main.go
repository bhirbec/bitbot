package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"

	"bitbot/database"
	"bitbot/errorutils"
	"services"
)

// TODO: HTTP handler are not panic safe
// TODO: cache static files
// TODO: make the static dir indepedent of the working directory

const (
	staticDir = "public"
)

var (
	configPath = flag.String("config", "src/services/config.json", "JSON file that stores credentials.")
	address    = flag.String("b", "localhost:8080", "host:port TCP informations")
)

var dbx *sqlx.DB

func main() {
	flag.Parse()

	// parse config
	config, err := services.LoadConfig(*configPath)
	if err != nil {
		log.Panic(err)
	}

	// open db
	creds := config.Mysql
	dbx = database.Openx(creds.Db, creds.Host, creds.Port, creds.User, creds.Pwd)
	defer dbx.Close()

	m := http.NewServeMux()

	m.Handle("/public/", http.StripPrefix("/public", http.FileServer(http.Dir(staticDir))))

	m.HandleFunc("/api/v1/bittrex", func(w http.ResponseWriter, r *http.Request) {
		JSONResponse(w, BittrexData(dbx))
	})

	m.HandleFunc("/", HomeHandler)
	http.HandleFunc("/", timeItWrapper(m))

	log.Printf("Starting webserver on %s\n", *address)
	if err := http.ListenAndServe(*address, nil); err != nil {
		panic(err)
	}
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

func BittrexData(db *sqlx.DB) interface{} {
	const stmt = `
        select
            m.market_name,
            m.volume,
            m.last,
            m.prev_day
        from
            market_summary m
            inner join (
                select
                    market_name,
                    max(creation_date) as creation_date
                from
                    market_summary
                group by
                    1
            )
            as d on d.market_name = m.market_name and d.creation_date = m.creation_date
        order by
            market_name
    `
	var rows []*struct {
		MarketName string          `db:"market_name"`
		Volume     decimal.Decimal `db:"volume"`
		Last       decimal.Decimal `db:"last"`
		PrevDay    decimal.Decimal `db:"prev_day"`
	}

	err := db.Select(&rows, fmt.Sprintf(stmt))
	errorutils.PanicOnError(err)
	return rows
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
