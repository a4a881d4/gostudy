package main

import (
	"fmt"

	"github.com/regcostajr/go-web3"
	"github.com/regcostajr/go-web3/providers"

	"github.com/a4a881d4/gostudy/ethereum/go-web3/web3ext"
)

var (
	source = `
pragma solilidy ^0.5.0;

contract Hello {
  function mul(uint a) public returns (uint b) {
  b = a*7 + 10;
  }
}
	`
)
func main() {
	var connection = web3.NewWeb3(providers.NewHTTPProvider("127.0.0.1:8501", 10, false))
	ext := web3ext.NewWeb3Ext(connection.Provider)

	if code, err := ext.CompileSolidity(source); err == nil {
		// json,_ := code.MarshalJSON()
		fmt.Println(string(code))
	} else {
		fmt.Println(err)
	}
}
