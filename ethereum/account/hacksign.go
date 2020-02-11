package main

import (
	"fmt"
	"math/big"

	"encoding/hex"
	
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/common"
	_ "github.com/ethereum/go-ethereum/crypto"

	"github.com/a4a881d4/gostudy/ethereum/account/ecc"
)
// go run ethereum/account/hacksign.go
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

	d := hackD()
	sign(d)
}

func hackD() *big.Int {
	curve := ecc.NewSecp256K1()
	// hack(key, r, s *big.Int, hash []byte) *big.Int
	key := big.NewInt(0)
	key.SetString("5a9d617f0db5a9a7d1ec4b97f8e5b12801d0c3a6386802fce907e7cd9fdead81",16)	
	r   := big.NewInt(0)
	r.SetString("3f31307871e1fd9c95136b5e3f2d193baac31d8d1efaa76e31c87bd6d7c5f547",16)	
	s   := big.NewInt(0)
	s.SetString("10ed9f0942feccfb626a91859e5c097ad52c14b291a39e17deb8c01a24fe37de",16)
	v   := 0x76a
	hash, _:= hex.DecodeString("5768ceae61357f4022ff731c2263b70113a3f23215db52742c7892208ac337b8")
	d   := curve.Hack(key,r,s,hash)
	fmt.Println(v,d.Text(16))
	return d
}

// func(curve *ECC) Sign(key, d *big.Int, hash []byte) (*big.Int,*big.Int) {
func sign(d *big.Int) (r,s *big.Int) {
	curve := ecc.NewSecp256K1()
	// hack(key, r, s *big.Int, hash []byte) *big.Int
	key := big.NewInt(0)
	key.SetString("5a9d617f0db5a9a7d1ec4b97f8e5b12801d0c3a6386802fce907e7cd9fdead81",16)	
	hash, _:= hex.DecodeString("5768ceae61357f4022ff731c2263b70113a3f23215db52742c7892208ac337b8")
	r,s   = curve.Sign(key,d,hash)
	fmt.Println(r.Text(16),s.Text(16))
	return r,s
}