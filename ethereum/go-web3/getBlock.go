package main

import (
	"fmt"

	"math/big"

	"github.com/regcostajr/go-web3"
	"github.com/regcostajr/go-web3/eth/block"
	"github.com/regcostajr/go-web3/providers"
)

func main() {

	var connection = web3.NewWeb3(providers.NewHTTPProvider("127.0.0.1:8501", 10, false))

	blockNumber, _ := connection.Eth.GetBlockNumber()

	var bn int64
	var accounts = make(map[string]*big.Int) 
	for bn = 0; bn < blockNumber.Int64(); bn += 1 {
		blockByNumber, err := connection.Eth.GetBlockByNumber(big.NewInt(bn), false)

		if err != nil {
			fmt.Println(err)
			panic(err)
		}
		if _,ok := accounts[blockByNumber.Miner]; !ok {
			accounts[blockByNumber.Miner] = big.NewInt(0)
		}
		count, err := connection.Eth.GetBlockTransactionCountByHash(blockByNumber.Hash)

		if err != nil {
			fmt.Println(err)
			panic(err)
		}

		// block, err := connection.Eth.GetBlockByHash(blockByHash, false)
		// GetBlockTransactionCountByHash
		if count.Cmp(big.NewInt(0)) > 0 {
			var tc int64
			for tc = 0; tc < count.Int64(); tc+= 1 {
				tx,_ := connection.Eth.GetTransactionByBlockHashAndIndex(blockByNumber.Hash,big.NewInt(tc))
				if _,ok := accounts[tx.From]; !ok {
					accounts[tx.From] = big.NewInt(0)
				}
				if _,ok := accounts[tx.To]; !ok {
					accounts[tx.To] = big.NewInt(0)
				} 
			}
			fmt.Printf("%5d block has %d tx, miner is %s\n",bn,count.Int64(),blockByNumber.Miner)
		}
	}

	for k,_ := range(accounts) {
		bal, err := connection.Eth.GetBalance(k, block.LATEST)
		if err == nil {
			accounts[k] = accounts[k].Set(bal)
		}
	}
	ac := 0
	space := "                                                        "
	for k,v := range(accounts) {
		if v.Cmp(big.NewInt(0)) > 0 {
			bal := v.String()
			fmt.Printf("%4d. [%s] has %s Wei\n",ac,k,space[:(30-len(bal))]+bal)
			ac += 1
		}
	}
}
