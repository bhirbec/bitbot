package orderbook

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type OrderBook struct {
	// Name of the exchanger
	Exchanger string
	Bids      []*Order
	Asks      []*Order
}

type Order struct {
	Price     float64
	Volume    float64
	Timestamp float64
}

func FetchOrderBook(url string, v interface{}) error {
	// create the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil
	}

	// execute the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil
	}

	// read the response body
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	return json.Unmarshal(body, v)
}
