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
var (
	keyString = "5a9d617f0db5a9a7d1ec4b97f8e5b12801d0c3a6386802fce907e7cd9fdead81"
	testHash  = "5768ceae61357f4022ff731c2263b70113a3f23215db52742c7892208ac337b8"
)
// go run ethereum/account/hacksign.go
func tx() {
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
func main() {
	curve := ecc.NewSecp256K1()
	fmt.Println("G.X",curve.G.X.Text(16))
	NG := curve.PointScale(&curve.G,big.NewInt(0).Add(big.NewInt(1),&curve.N))
	fmt.Println("NG.X",NG.X.Text(16))
	d   := hackD()
	r,s := sign(d)

	hack(r,s)
	verify(r,s)
	check(r,s)
}

func hack(r,s *big.Int) *big.Int {
	curve := ecc.NewSecp256K1()
	// hack(key, r, s *big.Int, hash []byte) *big.Int
	key := big.NewInt(0)
	key.SetString(keyString,16)	
	hash, _:= hex.DecodeString(testHash)
	d   := curve.Hack(key,r,s,hash)
	fmt.Println("hack",d.Text(16))
	return d	
}
func zero() *big.Int {
	return big.NewInt(0)
}
func check(r,s *big.Int) {
	curve := ecc.NewSecp256K1()
	hash, _:= hex.DecodeString(testHash)
	e := zero().SetBytes(hash)
	key := big.NewInt(0)
	key.SetString(keyString,16)

	iS := zero().ModInverse(s, &curve.N)

	u1 := zero().Rem(zero().Mul(e,iS),&curve.N)

	u2 := zero().Rem(zero().Mul(r,iS),&curve.N)

	d := u1.Add(u1,u2.Mul(u2,key))
	d.Rem(d,&curve.N)
	fmt.Println("check",d.Text(16))
	dG := curve.PointScale(&curve.G,d)
	nr := zero().Rem(&dG.X,&curve.N)
	fmt.Println("nr",nr.Text(16))
	
}
func verify(r,s *big.Int) {
//	func(curve *ECC) Verify(Q *ECPoint, r, s *big.Int, hash []byte) bool {
	curve := ecc.NewSecp256K1()
	key := big.NewInt(0)
	key.SetString(keyString,16)
	hash, _:= hex.DecodeString(testHash)
	
	Q := curve.PointScale(&curve.G,key)

	if ok,vr := curve.Verify2(Q,r,s,hash); ok {
		fmt.Println("OK")
	} else {
		fmt.Println("F",vr.Text(16))
	}
}
func hackD() *big.Int {
	curve := ecc.NewSecp256K1()
	// hack(key, r, s *big.Int, hash []byte) *big.Int
	key := big.NewInt(0)
	key.SetString(keyString,16)	
	r   := big.NewInt(0)
	r.SetString("3f31307871e1fd9c95136b5e3f2d193baac31d8d1efaa76e31c87bd6d7c5f547",16)	
	s   := big.NewInt(0)
	s.SetString("10ed9f0942feccfb626a91859e5c097ad52c14b291a39e17deb8c01a24fe37de",16)
	v   := 0x76a
	hash, _:= hex.DecodeString(testHash)
	d   := curve.Hack(key,r,s,hash)
	fmt.Println("hackD",v,d.Text(16))
	Q := curve.PointScale(&curve.G,key)

	if ok,vr := curve.Verify(Q,r,s,hash); ok {
		fmt.Println("-ok")
	} else {
		fmt.Println("-f",vr.Text(16))
	}
	return d
}

// func(curve *ECC) Sign(key, d *big.Int, hash []byte) (*big.Int,*big.Int) {
func sign(d *big.Int) (r,s *big.Int) {
	curve := ecc.NewSecp256K1()
	// hack(key, r, s *big.Int, hash []byte) *big.Int
	key := big.NewInt(0)
	key.SetString(keyString,16)	
	hash, _:= hex.DecodeString(testHash)
	r,s   = curve.Sign(key,d,hash)
	fmt.Println("r=",r.Text(16))
	fmt.Println("s=",s.Text(16))
	return r,s
}