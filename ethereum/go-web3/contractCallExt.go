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


// go run ethereum/go-web3/contractCallExt.go ./contract/o5g.abi 0xf83da04712c5304919fc4985f4f75d3522418870
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

	for k,v := range(contract.Abi.Methods) {
		fmt.Printf("%x %20s %s\n",v.ID(),k,v.Sig())
	}

	transaction := new(dto.TransactionParameters)
	coinbase, err := connection.Eth.GetCoinbase()
	transaction.From = coinbase
	transaction.Gas = big.NewInt(4000000)

	transaction.To = os.Args[2]
	v := ""
	result, err := contract.Call(&v,transaction, "name")

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("name  ",*result.(*string))
	}

	result, err = contract.Call(&v,transaction, "symbol")

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("symbol",*result.(*string))
	}

	for i:=0;i<32;i++ {
		b := big.NewInt(0)
		addr := common.HexToAddress(curve.PrivateKey2Address(prK))
		result, err = contract.Call(&b,transaction, "balances", addr)

		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("%5d. ",i)
			fmt.Println("balance of",addr.Hex(),*result.(**big.Int))
		}
		prK.Add(prK,big.NewInt(1))
	}

}
