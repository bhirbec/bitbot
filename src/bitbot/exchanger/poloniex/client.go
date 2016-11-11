package poloniex

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
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
// {"BTC":"0.59098578","LTC":"3.31117268", ... }
func (c *Client) TradingBalances() (map[string]string, error) {
	v := map[string]string{}
	err := c.post("returnBalances", nil, &v)
	return v, err
}

// Returns all of your deposit addresses
func (c *Client) DepositAddresses() (map[string]string, error) {
	v := map[string]string{}
	err := c.post("returnDepositAddresses", nil, &v)
	return v, err
}

// Places a limit buy order in a given market
func (c *Client) Buy(pair exchanger.Pair, rate, amount float64) (map[string]interface{}, error) {
	return c.placeOrder("buy", pair, rate, amount)
}

// Places a sell order in a given market
func (c *Client) Sell(pair exchanger.Pair, rate, amount float64) (map[string]interface{}, error) {
	return c.placeOrder("sell", pair, rate, amount)
}

// Places a limit buy order in a given market
func (c *Client) placeOrder(cmd string, pair exchanger.Pair, rate, amount float64) (map[string]interface{}, error) {
	p, ok := Pairs[pair]
	if !ok {
		return nil, fmt.Errorf("Pair not supported %s", pair)
	}

	v := map[string]interface{}{}
	data := &url.Values{}
	data.Add("currencyPair", p)
	data.Add("rate", fmt.Sprint(rate))
	data.Add("amount", fmt.Sprint(amount))
	err := c.post(cmd, data, &v)
	return v, err
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
