package hitbtc

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"bitbot/httpreq"
)

// TODO: defined struct to be returned instead of interface{}
// TODO: check use of fmt.Printf() makes amounts have the right number of decimals
// TODO: clarify naming between currency and symbol
// TODO: generate unique clientOrderId for order creation

type Client struct {
	ApiKey    string
	ApiSecret string
}

func NewClient(apiKey, apiSecret string) *Client {
	return &Client{apiKey, apiSecret}
}

// TradingBalance returns trading balance.
func (c *Client) TradingBalance() (interface{}, error) {
	const path = "/api/1/trading/balance"
	v := map[string]interface{}{}
	err := c.authGet(path, &v)
	return v, err
}

// PlaceOrder places a new order.
func (c *Client) PlaceOrder(side, symbol string, price, quantity float64, orderType string) (interface{}, error) {
	const path = "/api/1/trading/new_order"

	// 1 lot equals 0.01 BTC
	qtyInLots := fmt.Sprintf("%.12f", quantity*100)

	data := &url.Values{
		"clientOrderId": []string{fmt.Sprintf("hitbtc-%d", makeTimestamp())},
		"symbol":        []string{symbol},
		"side":          []string{side},
		"price":         []string{fmt.Sprint(price)},
		"quantity":      []string{qtyInLots},
		"type":          []string{orderType},
		"timeInForce":   []string{"GTC"},
	}

	v := map[string]interface{}{}
	err := c.authPost(path, data, &v)
	return v, err
}

// CancelOrder cancels an order.
func (c *Client) CancelOrder(clientOrderId, symbol, side string) (interface{}, error) {
	const path = "/api/1/trading/cancel_order"

	cancelRequestClientOrderId := fmt.Sprintf("cancel-order-%d", makeTimestamp())

	data := &url.Values{
		"clientOrderId":              []string{clientOrderId},
		"cancelRequestClientOrderId": []string{cancelRequestClientOrderId},
		"symbol":                     []string{symbol},
		"side":                       []string{side},
	}

	v := map[string]interface{}{}
	err := c.authPost(path, data, &v)
	return v, err
}

// TransfertToTradingAccount transfers funds from main and to trading accounts.
// It returns a transaction ID.
func (c *Client) TransfertToTradingAccount(amount float64, currencyCode string) (string, error) {
	const path = "/api/1/payment/transfer_to_trading"
	return c.transfert(path, amount, currencyCode)
}

// TransfertToMainAccount transfers funds from trading accounts to main.
// It returns a transaction ID
func (c *Client) TransfertToMainAccount(amount float64, currencyCode string) (string, error) {
	const path = "/api/1/payment/transfer_to_main"
	return c.transfert(path, amount, currencyCode)
}

func (c *Client) transfert(path string, amount float64, currencyCode string) (string, error) {
	data := &url.Values{
		"amount":        []string{fmt.Sprint(amount)},
		"currency_code": []string{currencyCode},
	}

	var v struct{ Transaction string }
	err := c.authPost(path, data, &v)
	return v.Transaction, err
}

// PaymentAddress returns the last created incoming cryptocurrency address that
// can be used to deposit.
func (c *Client) PaymentAddress(currency string) (string, error) {
	const path = "/api/1/payment/address/"
	var v struct{ Address string }
	err := c.authGet(path+currency, &v)
	return v.Address, err
}

// CreateAddress creates an address that can be used to deposit cryptocurrency to your account.
// It returns a new cryptocurrency address.
func (c *Client) CreateAddress(currency string) (string, error) {
	const path = "/api/1/payment/address/"
	var v struct{ Address string }
	err := c.authPost(path+currency, nil, &v)
	return v.Address, err
}

// Withdraw withdraws money and creates an outgoing crypotocurrency transaction. It returns
// a transaction ID. Withdraw operates on the main account (not the trading account).
func (c *Client) Withdraw(amount float64, currencyCode, address string) (string, error) {
	const path = "/api/1/payment/payout"

	data := &url.Values{
		"amount":        []string{fmt.Sprint(amount)},
		"currency_code": []string{currencyCode},
		"address":       []string{address},
	}

	var v struct{ Transaction string }
	err := c.authPost(path, data, &v)
	return v.Transaction, err
}

func (c *Client) authGet(path string, v interface{}) error {
	uri := authURI(path, c.ApiKey)
	headers := authHeader(uri, "", c.ApiSecret)
	return httpreq.Get(host+uri, headers, v)
}

func (c *Client) authPost(path string, data *url.Values, v interface{}) error {
	var body string
	if data != nil {
		body = data.Encode()
	}
	uri := authURI(path, c.ApiKey)
	headers := authHeader(uri, body, c.ApiSecret)
	return httpreq.Post(host+uri, headers, body, v)
}

func authURI(path, apiKey string) string {
	return fmt.Sprintf("%s?nonce=%d&apikey=%s", path, makeTimestamp(), apiKey)
}

func authHeader(uri, body, apiSecret string) http.Header {
	signature := sign(uri+body, apiSecret)
	headers := http.Header{}
	headers.Add("X-Signature", signature)
	return headers
}

func sign(msg, apiSecret string) string {
	h := hmac.New(sha512.New, []byte(apiSecret))
	h.Write([]byte(msg))
	return strings.ToLower(hex.EncodeToString(h.Sum(nil)))
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Microsecond)
}
