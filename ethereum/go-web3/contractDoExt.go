package main

import (
	_ "encoding/json"
	"fmt"
	"os"
	"io/ioutil"
	"math/big"
	"github.com/a4a881d4/gostudy/ethereum/go-web3/web3ext"
	"github.com/a4a881d4/gostudy/ethereum/account/ecc"
	"github.com/a4a881d4/gostudy/constant"

	"github.com/regcostajr/go-web3"
	"github.com/regcostajr/go-web3/dto"
	"github.com/regcostajr/go-web3/providers"

	"github.com/ethereum/go-ethereum/common"
)


// go run ethereum/go-web3/contractDoExt.go ./contract/o5g.abi 0xf83da04712c5304919fc4985f4f75d3522418870
func main() {

	curve := ecc.NewSecp256K1()
  	prK := new(big.Int)
    prK.SetString(constant.PrivateKeyStart,16)

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
	transaction.To = os.Args[2]

	for i:=0;i<4096;i++ {
		to := curve.PrivateKey2Address(prK)
		_, err := contract.Do(transaction, "transfer", common.HexToAddress(to), big.NewInt(1000_000_000))

		if err != nil {
			fmt.Println(err)
			panic("Do failure")
		}

		// var receipt *dto.TransactionReceipt

		// for receipt == nil {
		// 	receipt, err = connection.Eth.GetTransactionReceipt(hash)
		// }

		// if err != nil {
		// 	fmt.Println(err)
		// 	panic("Receipt Failure")
		// }

		// fmt.Println("receipt: ", receipt)
		prK.Add(prK,big.NewInt(1))
	}
}
