package main

import (
	"fmt"
	web3 "github.com/regcostajr/go-web3"
	"github.com/regcostajr/go-web3/eth/block"
	"github.com/regcostajr/go-web3/providers"
)

func main() {
	var connection = web3.NewWeb3(providers.NewHTTPProvider("127.0.0.1:8501", 10, false))

	coinbase, _ := connection.Eth.GetCoinbase()

	if bal, err := connection.Eth.GetBalance(coinbase, block.LATEST); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(coinbase, " has ", bal)
	}
}
