package kraken

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

const APIURL = "https://api.kraken.com/0"

type OrderBook struct {
	Bids []*Order
	Asks []*Order
}

type Order struct {
	Price     float64
	Volume    float64
	Timestamp float64
}

func FetchOrderBook(pair string) (*OrderBook, error) {
	url := fmt.Sprintf("%s/public/Depth?pair=%s", APIURL, pair)

	// create the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// execute the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	// read the response body
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Error  []string
		Result map[string]struct {
			Bids [][]interface{}
			Asks [][]interface{}
		}
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	if len(result.Error) > 0 {
		return nil, fmt.Errorf("Kraken returned an error. %s", result.Error[0])
	}

	asks, err := parseOrders(result.Result[pair].Asks)
	if err != nil {
		return nil, err
	}

	bids, err := parseOrders(result.Result[pair].Bids)
	if err != nil {
		return nil, err
	}

	return &OrderBook{Bids: bids, Asks: asks}, nil
}

func parseOrders(rows [][]interface{}) ([]*Order, error) {
	orders := make([]*Order, len(rows))
	for i, row := range rows {
		price, err := strconv.ParseFloat(row[0].(string), 64)
		if err != nil {
			return nil, err
		}

		volume, err := strconv.ParseFloat(row[1].(string), 64)
		if err != nil {
			return nil, err
		}

		orders[i] = &Order{
			Price:     price,
			Volume:    volume,
			Timestamp: row[2].(float64),
		}
	}

	return orders, nil
}
