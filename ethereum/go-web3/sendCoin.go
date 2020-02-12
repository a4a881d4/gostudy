package main
import (
	"fmt"
	"math/big"
	"os"
	"strconv"

	"github.com/a4a881d4/gostudy/ethereum/go-web3/web3ext"
	"github.com/a4a881d4/gostudy/ethereum/account/ecc"
	"github.com/a4a881d4/gostudy/constant"

	"github.com/regcostajr/go-web3"
	"github.com/regcostajr/go-web3/providers"
)
var (
	defaultkeyString = "5a9d617f0db5a9a7d1ec4b97f8e5b12801d0c3a6386802fce907e7cd9fdead81"
)
func main() {
	var connection = web3.NewWeb3(providers.NewHTTPProvider("127.0.0.1:8501", 10, false))
	
	curve := ecc.NewSecp256K1()
  	prK := new(big.Int)
    prK.SetString(constant.PrivateKeyStart,16)

	ext     := web3ext.NewWeb3Ext(connection.Provider)
	Config := web3ext.NewConfig(big.NewInt(0x3b9aca00),0x61a80,931)

	var key string
	var value int64
	if len(os.Args) == 1 {
    	key   = defaultkeyString
    	value = 10
  	} else {
  		value,_   = strconv.ParseInt(os.Args[1], 10, 64)
    	key = os.Args[2]
  	}
	
	txC,err := ext.SendCoin(0,"0x0caebc448230a6f9a7c998aa8b452ec0ab02aef6", 
		big.NewInt(0).Mul(big.NewInt(1), big.NewInt(1E18)), 
		key, 
		[]byte("p2p transaction"), Config)
		if err != nil {
			fmt.Println(err)
		}
	for i:=0;i<32;i++ {
		to := "0x"+curve.PrivateKey2Address(prK)
		_,err = ext.SendCoin(txC,to, 
			big.NewInt(0).Mul(big.NewInt(value), big.NewInt(1E18)), 
			key, 
			[]byte("p2p transaction"), Config)
		if err != nil {
			fmt.Println(err)
		}
		txC++
		prK.Add(prK,big.NewInt(1))
	}
	
}