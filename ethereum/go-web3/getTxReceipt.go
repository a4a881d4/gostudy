package main

import (
	"fmt"
	"os"
	web3 "github.com/regcostajr/go-web3"
	"github.com/regcostajr/go-web3/providers"
)

// go run ethereum/go-web3/getTxReceipt.go 0xa4417316e331412fe1875597bd8ce12b079c909d18219361571b757c7ba9ad03
func main() {
	var connection = web3.NewWeb3(providers.NewHTTPProvider("127.0.0.1:8501", 10, false))

	if txr, err := connection.Eth.GetTransactionReceipt(os.Args[1]); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(txr)
	}
}
