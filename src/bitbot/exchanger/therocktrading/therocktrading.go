package therocktrading

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"net/http"
	urlpkg "net/url"
	"strings"
	"time"

	"bitbot/exchanger"
	"bitbot/httpreq"
)

const (
	APIURL        = "https://api.therocktrading.com/v1"
	ExchangerName = "The Rock Trading"
)

// Pairs maps standardized currency pairs to Poloniex pairs as used by the API.
var Pairs = map[exchanger.Pair]string{
	exchanger.LTC_BTC: "LTCBTC",
	exchanger.ETH_BTC: "ETHBTC",
	exchanger.ZEC_BTC: "ZECBTC",
}

func OrderBook(pair exchanger.Pair) (*exchanger.OrderBook, error) {
	p, ok := Pairs[pair]
	if !ok {
		return nil, fmt.Errorf("The Rock Trading: OrderBook function doesn't not support %s", pair)
	}

	var result struct {
		Asks []map[string]float64
		Bids []map[string]float64
	}

	url := fmt.Sprintf("%s/funds/%s/orderbook", APIURL, p)
	if err := exchanger.FetchOrderBook(url, &result); err != nil {
		return nil, err
	}

	bids, err := parseOrders(result.Bids)
	if err != nil {
		return nil, err
	}

	asks, err := parseOrders(result.Asks)
	if err != nil {
		return nil, err
	}

	return exchanger.NewOrderbook(ExchangerName, bids, asks)
}

func parseOrders(rows []map[string]float64) ([]*exchanger.Order, error) {
	orders := make([]*exchanger.Order, len(rows))
	for i, row := range rows {
		orders[i] = &exchanger.Order{
			Price:  row["price"],
			Volume: row["amount"],
		}
	}

	return orders, nil
}

type Client struct {
	ApiKey    string
	ApiSecret string
}

func NewClient(apiKey, apiSecret string) *Client {
	return &Client{apiKey, apiSecret}
}

// Balances returns a list of all your balances in any currency.
func (c *Client) Balances() ([]Balance, error) {
	var v struct{ Balances []Balance }
	url := fmt.Sprintf("%s/balances", APIURL)
	err := c.get(url, &v)
	return v.Balances, err
}

func (c *Client) Transactions() ([]Transaction, error) {
	var v struct {
		Transactions []Transaction
	}
	url := fmt.Sprintf("%s/transactions", APIURL)
	err := c.get(url, &v)
	return v.Transactions, err
}

// PlaceOrder places a limit order in a given market.
func (c *Client) PlaceOrder(side string, pair exchanger.Pair, price, amount float64) (*Order, error) {
	p, ok := Pairs[pair]
	if !ok {
		return nil, fmt.Errorf("The Rock Trading: Pair not supported %s", pair)
	}

	url := fmt.Sprintf("%s/funds/%s/orders", APIURL, p)

	data := &urlpkg.Values{}
	data.Add("fund_id", p)
	data.Add("side", side)
	data.Add("amount", fmt.Sprint(amount))
	data.Add("price", fmt.Sprint(price))

	order := &Order{}
	err := c.post(url, data, order)
	return order, err
}

// Withdraw places a withdrawal for a given currency.
func (c *Client) Withdraw(amount float64, currency, address string) (string, error) {
	url := fmt.Sprintf("%s/atms/withdraw", APIURL)
	data := &urlpkg.Values{}
	data.Add("currency", currency)
	data.Add("amount", fmt.Sprint(amount))
	data.Add("destination_address", address)

	var v struct{ Transaction_id string }
	err := c.post(url, data, &v)
	return v.Transaction_id, err
}

// WithdrawLimit returns a currency related withdraw limit
func (c *Client) WithdrawLimit(currency string) (WithdrawLimit, error) {
	url := fmt.Sprintf("%s/withdraw_limits/%s", APIURL, currency)
	resp := WithdrawLimit{}
	err := c.get(url, &resp)
	return resp, err
}

// WithdrawLimit returns a list of your global and currently available withdraw levels.
func (c *Client) WithdrawLimits() ([]WithdrawLimit, error) {
	url := fmt.Sprintf("%s/withdraw_limits", APIURL)
	var resp struct{ Withdrawlimits []WithdrawLimit }
	err := c.get(url, &resp)
	return resp.Withdrawlimits, err
}

func (c *Client) get(url string, v interface{}) error {
	return httpreq.Get(url, c.authHeader(url), v)
}

func (c *Client) post(url string, data *urlpkg.Values, v interface{}) error {
	return httpreq.Post(url, c.authHeader(url), data.Encode(), v)
}

func (c *Client) authHeader(url string) http.Header {
	nonce := fmt.Sprintf("%d", makeTimestamp())
	message := nonce + url
	signature := sign(message, c.ApiSecret)

	headers := http.Header{}
	headers.Add("X-TRT-KEY", c.ApiKey)
	headers.Add("X-TRT-SIGN", signature)
	headers.Add("X-TRT-NONCE", nonce)
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

// response structs

type Balance struct {
	Currency       string
	Balance        float64
	TradingBalance float64 `json:"trading_balance"`
}

type Order struct {
	// Order ID
	Id int
	// Fund symbol
	FundId string `json:"fund_id"`
	// Order side (sell, buy, close_short or close_long)
	Side string
	// Order type (limit or market)
	Type string
	// Order status: active, conditional, executed or deleted
	Status string
	// Order price
	Price float64
	// Order total amount
	Amount float64
	// Order actual amount
	AmountUnfilled float64 `json:"amount_unfilled"`
	// Order conditional type [stop_loss|take_profit]
	Conditional_type string `json:"conditional_type"`
	// Order submission timestamp (UTC time)
	Date string
	// Order to close at close_on timestamp if present
	CloseOn string `json:"conditional_type"`
	// Order leverage. Default 1 if not a leveraged order
	Leverage float64
	// Position ID present if a closing order (close_short|close_long)
	PositionId int `json:"position_id"`
	// Order resulting trades
	Trades []Trade
}

type Trade struct {
	// Trade id
	Id int
	// Fund symbol
	FundId string `json:"fund_id"`
	// Actual traded amount
	Amount float64
	// Actual traded price
	Price float64
	// Type of order maker
	Side string
	// True if at least maker or taker order involved in this trade, were dark
	Dark bool
	// Trade execution timestamp (UTC time)
	Date string
}

type WithdrawLimit struct {
	// Currency
	Currency string
	// User daily limit
	Initial float64
	// Residual user limit
	Available float64
}

type Transaction struct {
	Id       int
	Type     string
	Price    float64
	Currency string
	Date     string
	// OrderId        interface{} `json:"order_id"`
	// TradeId        interface{} `json:"trade_id"`
	// Note           string
	// TransferDetail interface{} `json:"transfer_detail"`
}
