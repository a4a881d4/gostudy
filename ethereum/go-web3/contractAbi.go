package main

import (
	_ "encoding/json"
	"fmt"
	"os"
	"io/ioutil"
	_ "math/big"
	
	"github.com/a4a881d4/gostudy/ethereum/go-web3/web3ext"

	"github.com/regcostajr/go-web3"
	_ "github.com/regcostajr/go-web3/dto"
	"github.com/regcostajr/go-web3/providers"
)


// go run ethereum/go-web3/contractAbi.go ./contract/o5g.abi
func main() {

	var connection = web3.NewWeb3(providers.NewHTTPProvider("127.0.0.1:8501", 10, false))
	ext     := web3ext.NewWeb3Ext(connection.Provider)

	abi, err      := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Println(err)
		panic("File io")
	}
	contract, err := ext.NewContract(abi)

	for k,v := range(contract.Abi.Methods) {
		fmt.Printf("%x %20s %s\n",v.ID(),k,v.Sig())
	}

}
