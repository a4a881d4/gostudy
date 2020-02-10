package main

import(
	"fmt"
	"github.com/a4a881d4/gostudy/ethereum/account/ecc"
	"github.com/a4a881d4/gostudy/constant"
	"math/big"
	"os"
)
// go run ethereum/account/key2addr.go 4000000000000000000000000000000000000000000000000000000000000000

func main() {
  curve := ecc.NewSecp256K1()
  prK := new(big.Int)
  if os.Args.len() < 2 {
    prK.SetString(constant.PrivateKey2Address,16)
  } else {
    prK.SetString(os.Args[1],16)
  }

  for i:=0;i<16;i++ {
  	fmt.Println(curve.PrivateKey2Address(prK))
  	prK.Add(prK,big.NewInt(1))
  }
}
