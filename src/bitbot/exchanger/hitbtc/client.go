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

	"bitbot/exchanger"
	"bitbot/httpreq"
)

// TODO: defined struct to be returned instead of interface{}
// TODO: check use of fmt.Printf() makes amounts have the right number of decimals
// TODO: clarify naming between currency and symbol
// TODO: generate unique clientOrderId for order creation

// lot size as defined on https://hitbtc.com/api under "Currency symbols" section
var lotSizes = map[exchanger.Pair]float64{
	exchanger.BTC_USD:   0.01,
	exchanger.BTC_EUR:   0.01,
	exchanger.LTC_BTC:   0.1,
	exchanger.LTC_USD:   0.1,
	exchanger.LTC_EUR:   0.1,
	exchanger.DSH_BTC:   1,
	exchanger.ETH_BTC:   0.001,
	exchanger.QCN_BTC:   0.01,
	exchanger.FCN_BTC:   0.01,
	exchanger.LSK_BTC:   1,
	exchanger.LSK_EUR:   1,
	exchanger.STEEM_BTC: 0.001,
	exchanger.STEEM_EUR: 0.001,
	exchanger.SBD_BTC:   0.001,
	exchanger.DASH_BTC:  0.001,
	exchanger.XEM_BTC:   1,
	exchanger.XEM_EUR:   1,
	exchanger.EMC_BTC:   0.1,
	exchanger.EMC_EUR:   0.01,
	exchanger.SC_BTC:    100,
	exchanger.SC_USD:    1100,
	exchanger.ARDR_BTC:  1,
	exchanger.ZEC_BTC:   0.001,
}

type Client struct {
	ApiKey    string
	ApiSecret string
}

func NewClient(apiKey, apiSecret string) *Client {
	return &Client{apiKey, apiSecret}
}

// TradingBalance returns trading balance.
func (c *Client) TradingBalances() (map[string]float64, error) {
	var v struct {
		Balance []struct {
			Currency_code string
			Cash          float64
			Reserved      float64
		}
	}

	const path = "/api/1/trading/balance"
	if err := c.authGet(path, &v); err != nil {
		return nil, err
	}

	balances := make(map[string]float64)
	for _, row := range v.Balance {
		// the "cash" entry is the available trading balance and there's no need to decrease this value
		// using the "reserved" value.
		balances[row.Currency_code] = row.Cash
	}

	return balances, nil
}

// PlaceOrder places a new order.
func (c *Client) PlaceOrder(side string, pair exchanger.Pair, price, quantity float64, orderType string) (interface{}, error) {
	const path = "/api/1/trading/new_order"

	size, ok := lotSizes[pair]
	if !ok {
		return nil, fmt.Errorf("%s: No lot size for this currency pair %s", ExchangerName, pair)
	}

	p, ok := Pairs[pair]
	if !ok {
		return nil, fmt.Errorf("%s: Pair not traded on this market %s", ExchangerName, pair)
	}

	lots := fmt.Sprintf("%f", quantity/size)

	data := &url.Values{
		"clientOrderId": []string{fmt.Sprintf("hitbtc-%d", makeTimestamp())},
		"symbol":        []string{p},
		"side":          []string{side},
		"quantity":      []string{lots},
		"type":          []string{orderType},
	}

	// TODO: use decimal type?
	// TODO: what about stopLimit type?
	// TODO: Price, in currency units, consider price steps?
	if orderType == "limit" {
		data.Add("price", fmt.Sprint(price))
		data.Add("timeInForce", "GTC")
	} else {
		data.Add("timeInForce", "IOC")
	}

	v := map[string]interface{}{}
	err := c.authPost(path, data, &v)
	return v, err
}

// CancelOrder cancels an order.
func (c *Client) CancelOrder(clientOrderId string, pair exchanger.Pair, side string) (interface{}, error) {
	const path = "/api/1/trading/cancel_order"

	p, ok := Pairs[pair]
	if !ok {
		return nil, fmt.Errorf("%s: Pair not traded on this market %s", ExchangerName, pair)
	}

	cancelRequestClientOrderId := fmt.Sprintf("cancel-order-%d", makeTimestamp())

	data := &url.Values{
		"clientOrderId":              []string{clientOrderId},
		"cancelRequestClientOrderId": []string{cancelRequestClientOrderId},
		"symbol":                     []string{p},
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
