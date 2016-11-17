package exchanger

import (
	"strings"
)

// Pair is used to represent pairs of currency symnbol.
type Pair struct {
	Label string
	Base  string
	Quote string
}

func NewPair(pair string) Pair {
	currencies := strings.Split(pair, "_")
	return Pair{pair, currencies[0], currencies[1]}
}

func (p Pair) String() string {
	return p.Label
}

var (
	BTC_USD   = NewPair("BTC_USD")
	BTC_EUR   = NewPair("BTC_EUR")
	LTC_BTC   = NewPair("LTC_BTC")
	LTC_USD   = NewPair("LTC_USD")
	LTC_EUR   = NewPair("LTC_EUR")
	DSH_BTC   = NewPair("DSH_BTC")
	ETH_BTC   = NewPair("ETH_BTC")
	ETH_USD   = NewPair("ETH_USD")
	ETC_BTC   = NewPair("ETC_BTC")
	ETC_USD   = NewPair("ETC_USD")
	QCN_BTC   = NewPair("QCN_BTC")
	FCN_BTC   = NewPair("FCN_BTC")
	LSK_BTC   = NewPair("LSK_BTC")
	LSK_EUR   = NewPair("LSK_EUR")
	STEEM_BTC = NewPair("STEEM_BTC")
	STEEM_EUR = NewPair("STEEM_EUR")
	SBD_BTC   = NewPair("SBD_BTC")
	DASH_BTC  = NewPair("DASH_BTC")
	XEM_BTC   = NewPair("XEM_BTC")
	XEM_EUR   = NewPair("XEM_EUR")
	EMC_BTC   = NewPair("EMC_BTC")
	EMC_EUR   = NewPair("EMC_EUR")
	SC_BTC    = NewPair("SC_BTC")
	SC_USD    = NewPair("SC_USD")
	ARDR_BTC  = NewPair("ARDR_BTC")
	ZEC_BTC   = NewPair("ZEC_BTC")
)
