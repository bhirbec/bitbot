// docker build -t bitbot-img . && docker run --rm bitbot-img
package main

import (
	"exchanger/bitfinex"
	"exchanger/bter"
	"exchanger/hitbtc"
	"exchanger/kraken"
	"fmt"
)

func main() {
	krakenBook, err := kraken.OrderBook("XXBTXLTC")
	if err != nil {
		fmt.Printf("%s\n", err)
	} else {
		fmt.Printf("%s\n", krakenBook.Asks[0].Price)
	}

	hitbtcBook, err := hitbtc.OrderBook("LTCBTC")
	if err != nil {
		fmt.Printf("%s\n", err)
	} else {
		fmt.Printf("%s\n", hitbtcBook.Asks[0].Price)
	}

	bitfinexBook, err := bitfinex.OrderBook("LTCBTC")
	if err != nil {
		fmt.Printf("%s\n", err)
	} else {
		fmt.Printf("%s\n", bitfinexBook.Asks[0].Price)
	}

	bterOrderBook, err := bter.OrderBook("LTC_BTC")
	if err != nil {
		fmt.Printf("%s\n", err)
	} else {
		fmt.Printf("%s\n", bterOrderBook.Asks[0].Price)
	}

}
