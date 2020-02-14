package main

import (
	"encoding/json"
	"fmt"
	"os"
	"io/ioutil"
	"math/big"
	"github.com/a4a881d4/gostudy/ethereum/go-web3/web3ext"
	
	"github.com/regcostajr/go-web3"
	"github.com/regcostajr/go-web3/dto"
	"github.com/regcostajr/go-web3/providers"
)


// go run ethereum/go-web3/contractDeployExt.go ./contract/o5g.abi ./contract/o5g.bin
func main() {

	content, err := ioutil.ReadFile(os.Args[2])

	type Contract struct {
		Bytecode string `json:"object"`
	}

	var jsonResponse Contract

	json.Unmarshal(content, &jsonResponse)

	fmt.Println(jsonResponse)
	bytecode := jsonResponse.Bytecode


	var connection = web3.NewWeb3(providers.NewHTTPProvider("127.0.0.1:8501", 10, false))
	ext     := web3ext.NewWeb3Ext(connection.Provider)

	abi, err      := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Println(err)
		panic("File io")
	}
	contract, err := ext.NewContract(abi)

	transaction := new(dto.TransactionParameters)
	coinbase, err := connection.Eth.GetCoinbase()
	transaction.From = coinbase
	transaction.Gas = big.NewInt(3000000)

	hash, err := contract.Deploy(transaction, "0x"+bytecode, big.NewInt(1000_000_000_000),"MCOIN",uint8(2),"O5G")

	if err != nil {
		fmt.Println(err)
		panic("Deploy failure")
	}

	var receipt *dto.TransactionReceipt

	for receipt == nil {
		receipt, err = connection.Eth.GetTransactionReceipt(hash)
	}

	if err != nil {
		fmt.Println(err)
		panic("Receipt Failure")
	}

	fmt.Println("Contract Address: ", receipt.ContractAddress)
}
