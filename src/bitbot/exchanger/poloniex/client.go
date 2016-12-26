package poloniex

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"bitbot/exchanger"
	"bitbot/httpreq"
)

const (
	tradingAPI = "https://poloniex.com/tradingApi"
)

// Client that supports trading API Methods.
type Client struct {
	// Your API key
	ApiKey string
	// Your API secret
	ApiSecret string
}

// NewClient creates a new Client
func NewClient(apiKey, apiSecret string) *Client {
	return &Client{apiKey, apiSecret}
}

// TradingBalance returns all of your available balances. Sample output:
// {"BTC": 0.59098578,"LTC": 3.31117268, ... }
func (c *Client) TradingBalances() (map[string]float64, error) {
	v := map[string]string{}
	err := c.post("returnBalances", nil, &v)
	if err != nil {
		return nil, err
	}

	out := map[string]float64{}
	for key, str := range v {
		value, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return nil, err
		}
		out[key] = value
	}

	return out, nil
}

// Returns all of your deposit addresses
func (c *Client) DepositAddresses() (map[string]string, error) {
	v := map[string]string{}
	err := c.post("returnDepositAddresses", nil, &v)
	return v, err
}

// Places a limit buy order in a given market
func (c *Client) Buy(pair exchanger.Pair, rate, amount float64) (map[string]interface{}, error) {
	return c.PlaceOrder("buy", pair, rate, amount)
}

// Places a sell order in a given market
func (c *Client) Sell(pair exchanger.Pair, rate, amount float64) (map[string]interface{}, error) {
	return c.PlaceOrder("sell", pair, rate, amount)
}

// PlaceOrder places a limit order in a given market. The returned value is a map with the following structure:
// - orderNumber:xxx
// - resultingTrades: list of:
//   - amount: 0.19261550
//   - date: 2016-12-23 10:42:09
//   - rate: 0.05736700
//   - total: 0.01104977
//   - tradeID: xxx
//   - type: buy
func (c *Client) PlaceOrder(cmd string, pair exchanger.Pair, rate, amount float64) (map[string]interface{}, error) {
	p, ok := Pairs[pair]
	if !ok {
		return nil, fmt.Errorf("Pair not supported %s", pair)
	}

	v := map[string]interface{}{}
	data := &url.Values{}
	data.Add("currencyPair", p)
	data.Add("rate", fmt.Sprint(rate))
	data.Add("amount", fmt.Sprint(amount))

	// Example of err: {error: Total must be at least 0.0001.}
	err := c.post(cmd, data, &v)
	return v, err
}

// Withdraw places a withdrawal for a given currency, with no email confirmation. In order to use
// this method, the withdrawal privilege must be enabled for your API key.
func (c *Client) Withdraw(amount float64, currency, address string) (string, error) {
	data := &url.Values{}
	data.Add("currency", currency)
	data.Add("amount", fmt.Sprint(amount))
	data.Add("address", address)
	var v struct{ Response string }
	err := c.post("withdraw", data, &v)
	return v.Response, err
}

// OrderTrades returns all trades involving a given order, specified by the "orderNumber" parameter.
func (c *Client) OrderTrades(orderNumber string) ([]map[string]interface{}, error) {
	data := &url.Values{}
	data.Add("orderNumber", orderNumber)
	var dest []map[string]interface{}
	err := c.post("returnOrderTrades", data, &dest)
	return dest, err
}

func (c *Client) post(cmd string, data *url.Values, v interface{}) error {
	if data == nil {
		data = &url.Values{}
	}

	data.Add("command", cmd)
	data.Add("nonce", fmt.Sprint(nonce()))

	body := data.Encode()
	signature := c.sign(body)

	headers := http.Header{}
	headers.Add("Sign", signature)
	headers.Add("Key", c.ApiKey)
	headers.Add("Content-Type", "application/x-www-form-urlencoded")

	return httpreq.Post(tradingAPI, headers, body, v)
}

func (c *Client) sign(body string) string {
	h := hmac.New(sha512.New, []byte(c.ApiSecret))
	h.Write([]byte(body))
	return hex.EncodeToString(h.Sum(nil))
}

func nonce() int64 {
	return time.Now().UnixNano()
}
