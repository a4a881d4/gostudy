package main

import(
	"fmt"
	ecc "./ecc"
	"math/big"
)
// go run ethereum/account/key2addr.go 5a9d617f0db5a9a7d1ec4b97f8e5b12801d0c3a6386802fce907e7cd9fdead81

func main() {
  curve := ecc.NewSecp256K1()
  prK := new(big.Int)
  prK.SetString(os.Args[1],16)
  fmt.Println(curve.PrivateKey2Address(prK))
}
