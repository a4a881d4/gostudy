package main

import (
	"fmt"

	"github.com/regcostajr/go-web3"
	"github.com/regcostajr/go-web3/complex/types"
	"github.com/regcostajr/go-web3/dto"
	"github.com/regcostajr/go-web3/providers"
	"math/big"

	"github.com/a4a881d4/gostudy/ethereum/account/ecc"
	"github.com/a4a881d4/gostudy/constant"

)

func main() {

	var connection = web3.NewWeb3(providers.NewHTTPProvider("127.0.0.1:8501", 10, false))

	coinbase, err := connection.Eth.GetCoinbase()

	if err != nil {
		fmt.Println(err)
		panic("fail")
	}
	
	// result, err := connection.Personal.UnlockAccount(coinbase, "", 300)

	// if err != nil {
	// 	fmt.Println(err)
	// 	panic("unlock fail")
	// }

	// if !result {
	// 	fmt.Println("Can't unlock account")
	// 	panic("unlock fail")
	// }

	curve := ecc.NewSecp256K1()
	prK := new(big.Int)
	prK.SetString(constant.PrivateKeyStart,16)
	for i:=0;i<32;i++ {

	    addr := "0x" + curve.PrivateKey2Address(prK)
		transaction := new(dto.TransactionParameters)
		transaction.From = coinbase
		transaction.To = addr
		transaction.Value = big.NewInt(0).Mul(big.NewInt(1), big.NewInt(1E18))
		transaction.Gas = big.NewInt(400000)
		transaction.Data = types.ComplexString("p2p transaction")

		txID, err := connection.Eth.SendTransaction(transaction)

		if err != nil {
			fmt.Println(err)
			panic("send fail")
		}

		fmt.Println(txID)
		prK.Add(prK,big.NewInt(1))		
	}


}