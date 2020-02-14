package main

import (
	"fmt"
	"os"
	"io/ioutil"
	"math/big"
	
	"github.com/regcostajr/go-web3"
	"github.com/regcostajr/go-web3/dto"
	"github.com/regcostajr/go-web3/providers"
)

// go run ethereum/go-web3/contractCall.go ./contract/o5g.abi 0x0349579e4f126dccb0408e17ba9e4f0207b60df3
func main() {

	var connection = web3.NewWeb3(providers.NewHTTPProvider("127.0.0.1:8501", 10, false))
	abi, err := ioutil.ReadFile(os.Args[1])
	contract, err := connection.Eth.NewContract(string(abi))

	transaction := new(dto.TransactionParameters)
	coinbase, err := connection.Eth.GetCoinbase()
	transaction.From = coinbase
	transaction.Gas = big.NewInt(4000000)

	transaction.To = os.Args[2]
	
	result, err := contract.Call(transaction, "name")

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(result.ToComplexString())
}
