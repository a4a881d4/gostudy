package main

import (
	"fmt"
	"math/big"
	
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/common"
	_ "github.com/ethereum/go-ethereum/crypto"
)

func main() {
	// func newTransaction(nonce uint64, to *common.Address, amount *big.Int, gasLimit uint64
	//                   , gasPrice *big.Int, data []byte) *Transaction
	var to = common.HexToAddress("0caebc448230a6f9a7c998aa8b452ec0ab02aef6")
	amount := big.NewInt(0).Mul(big.NewInt(1), big.NewInt(1E18))
	gasPrice := big.NewInt(40000000)
	
	var str string = "p2p"
	var data []byte = []byte(str)

	tx := types.NewTransaction(1,to,amount,40000,gasPrice,data)

	if json,err := tx.MarshalJSON(); err == nil {
		fmt.Println(string(json))
	} else {
		fmt.Println("error")
	}
}