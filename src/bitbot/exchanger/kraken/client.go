package kraken

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"bitbot/exchanger"
	"bitbot/httpreq"
)

var Currencies = map[string]string{
	"BTC": "XXBT",
	"ZEC": "XZEC",
}

var reversedCurrencies = map[string]string{}

func init() {
	for k, v := range Currencies {
		reversedCurrencies[v] = k
	}
}

type Client struct {
	ApiKey    string
	ApiSecret string
}

func NewClient(apiKey, apiSecret string) *Client {
	return &Client{apiKey, apiSecret}
}

// AccountBalance returns account balance.
func (c *Client) AccountBalance() (map[string]float64, error) {
	resp := map[string]string{}
	data := map[string]string{}
	err := c.Query("Balance", data, &resp)
	if err != nil {
		return nil, err
	}

	out := map[string]float64{}
	for k, s := range resp {
		cur, ok := reversedCurrencies[k]
		if !ok {
			return nil, fmt.Errorf("Kraken: missing currency translation %s", k)
		}

		value, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, fmt.Errorf("Kraken: float parsing of `%s` failed - %s", s, err)
		}

		out[cur] = value
	}

	return out, err
}

// TradingBalance returns trading balance.
func (c *Client) TradeBalance(cur string) (float64, error) {
	convCur, ok := Currencies[cur]
	if ok {
		cur = convCur
	}

	var resp struct{ Tb string }
	data := map[string]string{"asset": cur}

	err := c.Query("TradeBalance", data, &resp)
	if err != nil {
		return 0, err
	}

	return strconv.ParseFloat(resp.Tb, 64)
}

func (c *Client) AddOrder(side string, pair exchanger.Pair, price, vol float64, ordertype string) (map[string]interface{}, error) {
	p, ok := Pairs[pair]
	if !ok {
		return nil, fmt.Errorf("Kraken: pair not supported %s", pair)
	}

	// TODO: use https://api.kraken.com/0/public/AssetPairs to retrieve lot multplier
	// amount to multiply lot volume by to get currency volume (1 for ZEC)
	var lotMult float64 = 1

	data := map[string]string{
		"type":      side,
		"pair":      p,
		"volume":    fmt.Sprint(vol / lotMult),
		"price":     fmt.Sprint(price),
		"ordertype": ordertype,
	}

	// map[descr:map[order:buy 0.01000000 ZECXBT @ market] txid:[XXX]]
	resp := map[string]interface{}{}
	err := c.Query("AddOrder", data, &resp)
	return resp, err
}

func (c *Client) Query(method string, data map[string]string, typ interface{}) error {
	values := url.Values{}
	for key, value := range data {
		values.Set(key, value)
	}

	urlPath := fmt.Sprintf("/%s/private/%s", APIVersion, method)
	reqURL := fmt.Sprintf("%s%s", APIURL, urlPath)
	secret, _ := base64.StdEncoding.DecodeString(c.ApiSecret)
	values.Set("nonce", fmt.Sprintf("%d", time.Now().UnixNano()))

	// Create signature
	signature := createSignature(urlPath, values, secret)

	// Add Key and signature to request headers
	headers := http.Header{}
	headers.Add("API-Key", c.ApiKey)
	headers.Add("API-Sign", signature)

	var resp struct {
		Error  interface{}
		Result interface{}
	}

	if typ != nil {
		resp.Result = typ
	}

	err := httpreq.Post(reqURL, headers, values.Encode(), &resp)
	if err != nil {
		return err
	}

	switch t := resp.Error.(type) {
	case string:
		if t != "" {
			return fmt.Errorf("Kraken error: %s", t)
		}
	case []string:
		if len(t) > 0 {
			return fmt.Errorf("Kraken errors: %s", t[0])
		}
	case []interface{}:
		if len(t) > 0 {
			return fmt.Errorf("Kraken errors: %s", t)
		}
	}

	return nil
}

// getSha256 creates a sha256 hash for given []byte
func getSha256(input []byte) []byte {
	sha := sha256.New()
	sha.Write(input)
	return sha.Sum(nil)
}

// getHMacSha512 creates a hmac hash with sha512
func getHMacSha512(message, secret []byte) []byte {
	mac := hmac.New(sha512.New, secret)
	mac.Write(message)
	return mac.Sum(nil)
}

func createSignature(urlPath string, values url.Values, secret []byte) string {
	// See https://www.kraken.com/help/api#general-usage for more information
	shaSum := getSha256([]byte(values.Get("nonce") + values.Encode()))
	macSum := getHMacSha512(append([]byte(urlPath), shaSum...), secret)
	return base64.StdEncoding.EncodeToString(macSum)
}
