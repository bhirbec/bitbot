// docker build -t bitbot-img . && docker run --rm bitbot-img
package main

import (
	"exchanger/kraken"
	"fmt"
)

func main() {
	orderBook, err := kraken.FetchOrderBook("XXBTXLTC")

	if err != nil {
		fmt.Printf("%s\n", err)
	} else {
		fmt.Printf("%s\n", orderBook.Asks[0].Price)
	}
}
