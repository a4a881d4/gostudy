package main

import (
	"fmt"
	"math/big"

	"encoding/hex"
	
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/common"
	_ "github.com/ethereum/go-ethereum/crypto"

	"github.com/a4a881d4/gostudy/ethereum/account/ecc"
	"github.com/a4a881d4/gostudy/ethereum/EIP155"
)
var (
	keyString = "5a9d617f0db5a9a7d1ec4b97f8e5b12801d0c3a6386802fce907e7cd9fdead81"
	
	rString   = "312da615d716aa4615e3dea5d211ba67c9277f275e7cda3ffbb97414f46885cb"
	sString   = "2e0e1e69327a884079f6fcabc15d18e2c20c754214618eb65c4fd523759beefd"
	testHash  = "a88b69cacbf06eddb07e3176f55fb290ac5e90507404aaac67826b04d053bd38"
	E155Hash  = "467cfbc43dca36fa2cf057d4b2e7a1a45d6c66628047579516b3feee9ed54f7b"

)
// go run ethereum/account/hacksign.go

func main() {
	// curve := ecc.NewSecp256K1()
	// fmt.Println("G.X",curve.G.X.Text(16))
	// NG := curve.PointScale(&curve.G,big.NewInt(0).Add(big.NewInt(1),&curve.N))
	// fmt.Println("NG.X",NG.X.Text(16))
	E155Hash = tx().Hex()[2:]

	d   := hackD()
	r,s,v := sign(d)

	verify(r,s)

	fmt.Println(v)

	// hack(r,s)
	// check(r,s)
	// checkSwap(d)
}

func verify(r,s *big.Int) {
	curve := ecc.NewSecp256K1()
	key := big.NewInt(0)
	key.SetString(keyString,16)
	hash, _:= hex.DecodeString(E155Hash)
	
	Q := curve.PointScale(&curve.G,key)

	if curve.Verify(Q,r,s,hash) {
		fmt.Println("OK")
	} else {
		fmt.Println("F")
	}
}

func hackD() *big.Int {
	curve := ecc.NewSecp256K1()
	key := big.NewInt(0)
	key.SetString(keyString,16)	
	r   := big.NewInt(0)
	r.SetString(rString,16)	
	s   := big.NewInt(0)
	s.SetString(sString,16)
	v   := 0x76a
	hash, _:= hex.DecodeString(E155Hash)
	d   := curve.Hack(key,r,s,hash)
	fmt.Println("hackD",v,d.Text(16))
	Q := curve.PointScale(&curve.G,key)

	if curve.Verify(Q,r,s,hash) {
		fmt.Println("-ok")
	} else {
		fmt.Println("-f")
	}
	return d
}

func sign(d *big.Int) (r,s *big.Int, v int64) {
	curve := ecc.NewSecp256K1()
	key := big.NewInt(0)
	key.SetString(keyString,16)	
	hash, _:= hex.DecodeString(E155Hash)
	r,s,v   = curve.Sign(key,d,hash)
	fmt.Println("r =",r.Text(16))
	fmt.Println("s =",s.Text(16))
	fmt.Println("v =",v)
	return r,s,v
}

func tx() common.Hash {
	/*
	"nonce":"0x12","gasPrice":"0x3b9aca00","gas":"0x61a80","to":"0x0caebc448230a6f9a7c998aa8b452ec0ab02aef6","value":"0xde0b6b3a76400
00","input":"0x703270207472616e73616374696f6e"
	*/
	var to = common.HexToAddress("0caebc448230a6f9a7c998aa8b452ec0ab02aef6")
	amount := big.NewInt(0).Mul(big.NewInt(1), big.NewInt(1E18))
	gasPrice := big.NewInt(1000000000)
	
	var str string = "p2p transaction"
	var data []byte = []byte(str)

	tx := types.NewTransaction(0x12,to,amount,400000,gasPrice,data)

	if json,err := tx.MarshalJSON(); err == nil {
		fmt.Println(string(json))
	} else {
		fmt.Println("error")
	}
	e155 := eip155.NewEIP155(931)
	hash := e155.Hash(tx)
	fmt.Println("hash",hash.Hex())
	return hash
}

func toString(p *ecc.ECPoint) string {
	r :=fmt.Sprintf(" x = %x \n y = %x \n",&p.X,&p.Y)
	return r
}

func checkSwap(d *big.Int) {
	curve := ecc.NewSecp256K1()
	key := big.NewInt(0)
	key.SetString(keyString,16)

	fmt.Println(toString(&curve.G),"-----G------")
	
	fmt.Println("d-0 is ",d.Text(16))
	dG := curve.PointScale(&curve.G,d)
	fmt.Println(toString(dG),"-----dG------")
	
	Q := curve.PointScale(&curve.G,key)
	fmt.Println(toString(Q),"-----Q------")
	
	fmt.Println("d-1 is ",d.Text(16))
	dQ := curve.PointScale(Q,d)
	fmt.Println(toString(dQ),"-----dQ------")
	
	dkG := curve.PointScale(&curve.G,zero().Mul(d,key))
	fmt.Println(toString(dkG),"-----dkG------")
}

func zero() *big.Int {
	return big.NewInt(0)
}

func check(r,s *big.Int) {
	curve := ecc.NewSecp256K1()
	hash, _:= hex.DecodeString(E155Hash)
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

func hack(r,s *big.Int) *big.Int {
	curve := ecc.NewSecp256K1()
	key := big.NewInt(0)
	key.SetString(keyString,16)	
	hash, _:= hex.DecodeString(E155Hash)
	d   := curve.Hack(key,r,s,hash)
	fmt.Println("hack",d.Text(16))
	return d	
}

