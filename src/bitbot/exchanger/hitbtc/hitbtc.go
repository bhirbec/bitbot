package hitbtc

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"bitbot/config"
	"bitbot/httpreq"
	"bitbot/orderbook"
)

// TODO: defined struct to be returned instead of interface{}
// TODO: check use of fmt.Printf() makes amounts have the right number of decimals
// TODO: clarify naming between currency and symbol
// TODO: generate unique clientOrderId for order creation

const (
	host          = "https://api.hitbtc.com"
	ExchangerName = "Hitbtc"
)

var (
	apiKey    = config.String("hitbtc", "api_key")
	apiSecret = config.String("hitbtc", "api_secret")
)

var Pairs = map[string]string{
	"btc_eur": "BTCEUR",
	"btc_usd": "BTCUSD",
	"ltc_btc": "LTCBTC",
	"ltc_usd": "LTCUSD",
	"eth_btc": "ETHBTC",
	"zec_btc": "ZECBTC",
}

func OrderBook(pair string) (*orderbook.OrderBook, error) {
	pair = Pairs[pair]
	url := fmt.Sprintf("%s/api/1/public/%s/orderbook", host, pair)

	var result struct {
		Asks [][]string
		Bids [][]string
	}

	if err := httpreq.Get(url, nil, &result); err != nil {
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

	return orderbook.NewOrderbook(ExchangerName, bids, asks)
}

func parseOrders(rows [][]string) ([]*orderbook.Order, error) {
	orders := make([]*orderbook.Order, len(rows))
	for i, row := range rows {
		price, err := strconv.ParseFloat(row[0], 64)
		if err != nil {
			return nil, err
		}

		volume, err := strconv.ParseFloat(row[1], 64)
		if err != nil {
			return nil, err
		}

		orders[i] = &orderbook.Order{
			Price:  price,
			Volume: volume,
		}
	}

	return orders, nil
}

// TradingBalance returns trading balance.
func TradingBalance() (interface{}, error) {
	const path = "/api/1/trading/balance"
	v := map[string]interface{}{}
	err := authGet(path, &v)
	return v, err
}

// NewOrder places a new order.
func NewOrder(side, symbol string, price, quantity float64, orderType string) (interface{}, error) {
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
	err := authPost(path, data, &v)
	return v, err
}

// CancelOrder cancels an order.
func CancelOrder(clientOrderId, symbol, side string) (interface{}, error) {
	const path = "/api/1/trading/cancel_order"

	cancelRequestClientOrderId := fmt.Sprintf("cancel-order-%d", makeTimestamp())

	data := &url.Values{
		"clientOrderId":              []string{clientOrderId},
		"cancelRequestClientOrderId": []string{cancelRequestClientOrderId},
		"symbol":                     []string{symbol},
		"side":                       []string{side},
	}

	v := map[string]interface{}{}
	err := authPost(path, data, &v)
	return v, err
}

// TransfertToTradingAccount transfers funds from main and to trading accounts.
// It returns a transaction ID.
func TransfertToTradingAccount(amount float64, currencyCode string) (string, error) {
	const path = "/api/1/payment/transfer_to_trading"
	return transfert(path, amount, currencyCode)
}

// TransfertToMainAccount transfers funds from trading accounts to main.
// It returns a transaction ID
func TransfertToMainAccount(amount float64, currencyCode string) (string, error) {
	const path = "/api/1/payment/transfer_to_main"
	return transfert(path, amount, currencyCode)
}

func transfert(path string, amount float64, currencyCode string) (string, error) {
	data := &url.Values{
		"amount":        []string{fmt.Sprint(amount)},
		"currency_code": []string{currencyCode},
	}

	var v struct{ Transaction string }
	err := authPost(path, data, &v)
	return v.Transaction, err
}

// PaymentAddress returns the last created incoming cryptocurrency address that
// can be used to deposit.
func PaymentAddress(currency string) (string, error) {
	const path = "/api/1/payment/address/"
	var v struct{ Address string }
	err := authGet(path+currency, &v)
	return v.Address, err
}

// CreateAddress creates an address that can be used to deposit cryptocurrency to your account.
// It returns a new cryptocurrency address.
func CreateAddress(currency string) (string, error) {
	const path = "/api/1/payment/address/"
	var v struct{ Address string }
	err := authPost(path+currency, nil, &v)
	return v.Address, err
}

// Withdraw withdraws money and creates an outgoing crypotocurrency transaction. It returns
// a transaction ID. Withdraw operates on the main account (not the trading account).
func Withdraw(amount float64, currencyCode, address string) (string, error) {
	const path = "/api/1/payment/payout"

	data := &url.Values{
		"amount":        []string{fmt.Sprint(amount)},
		"currency_code": []string{currencyCode},
		"address":       []string{address},
	}

	var v struct{ Transaction string }
	err := authPost(path, data, &v)
	return v.Transaction, err
}

// ========================================================

func authGet(path string, v interface{}) error {
	uri := authURI(path)
	headers := authHeader(uri, "")
	return httpreq.Get(host+uri, headers, v)
}

func authPost(path string, data *url.Values, v interface{}) error {
	var body string
	if data != nil {
		body = data.Encode()
	}
	uri := authURI(path)
	headers := authHeader(uri, body)
	return httpreq.Post(host+uri, headers, body, v)
}

func authURI(path string) string {
	return fmt.Sprintf("%s?nonce=%d&apikey=%s", path, makeTimestamp(), *apiKey)
}

func authHeader(uri, body string) http.Header {
	signature := sign(uri + body)
	headers := http.Header{}
	headers.Add("X-Signature", signature)
	return headers
}

func sign(msg string) string {
	h := hmac.New(sha512.New, []byte(*apiSecret))
	h.Write([]byte(msg))
	return strings.ToLower(hex.EncodeToString(h.Sum(nil)))
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Microsecond)
}
