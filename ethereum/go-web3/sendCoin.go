package main
import (
	"fmt"
	"math/big"

	"github.com/a4a881d4/gostudy/ethereum/go-web3/web3ext"

	"github.com/regcostajr/go-web3"
	"github.com/regcostajr/go-web3/providers"
)

func main() {
	var connection = web3.NewWeb3(providers.NewHTTPProvider("127.0.0.1:8501", 10, false))

	ext := web3ext.NewWeb3Ext(connection.Provider)

	if err := ext.SendCoin("0x0caebc448230a6f9a7c998aa8b452ec0ab02aef6", 
		big.NewInt(0).Mul(big.NewInt(1), big.NewInt(1E18)), 
		"5a9d617f0db5a9a7d1ec4b97f8e5b12801d0c3a6386802fce907e7cd9fdead81", 
		[]byte("p2p transaction"), nil); err != nil {
		fmt.Println(err)
	}
}