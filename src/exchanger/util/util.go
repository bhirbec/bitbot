package util

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

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
