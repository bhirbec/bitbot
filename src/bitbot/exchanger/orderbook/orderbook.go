package orderbook

import (
	"fmt"

	"bitbot/httpreq"
)

type OrderBook struct {
	// Name of the exchanger
	Exchanger string
	Bids      []*Order
	Asks      []*Order
}

type Order struct {
	Price     float64
	Volume    float64
	Timestamp float64
}

// NewOrderbook returns a new orderbook. An error is returned when orders are
// not sorted as expected or when they are no orders in at least one side of the book.
func NewOrderbook(Exchanger string, bids, asks []*Order) (*OrderBook, error) {
	// verify orderbook isn't empty
	if len(bids) == 0 {
		return nil, fmt.Errorf("Orderbook: no ask orders (%s).", Exchanger)
	} else if len(asks) == 0 {
		return nil, fmt.Errorf("Orderbook: no bid orders (%s).", Exchanger)
	}

	// verify bid orders are sorted
	maxBid := bids[0].Price
	for _, o := range bids[1:] {
		if o.Price > maxBid {
			return nil, fmt.Errorf("Orderbook: %s bid orders are not sorted.", Exchanger)
		}
		maxBid = o.Price
	}

	// verify ask orders are sorted
	minAsk := asks[0].Price
	for _, o := range asks[1:] {
		if o.Price < minAsk {
			return nil, fmt.Errorf("Orderbook: %s ask orders are not sorted", Exchanger)
		}
		minAsk = o.Price
	}

	return &OrderBook{Exchanger, bids, asks}, nil
}

// TODO: inline those function calls
func FetchOrderBook(url string, v interface{}) error {
	return httpreq.Get(url, nil, v)
}

func ReverseOrders(orders []*Order) []*Order {
	n := len(orders)
	output := make([]*Order, n)
	for i := 0; i < n; i++ {
		output[i] = orders[n-1-i]
	}
	return output
}
