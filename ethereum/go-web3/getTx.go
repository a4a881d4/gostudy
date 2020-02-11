package main

import (
	"fmt"
	"os"
	"encoding/json"

	"github.com/regcostajr/go-web3"
	"github.com/regcostajr/go-web3/providers"
	"github.com/regcostajr/go-web3/dto"

	"github.com/a4a881d4/gostudy/go-web3/web3ext"
)

// go run ethereum/go-web3/getTx.go 0x5768ceae61357f4022ff731c2263b70113a3f23215db52742c7892208ac337b8


func main() {
	var connection = web3.NewWeb3(providers.NewHTTPProvider("127.0.0.1:8501", 10, false))

	tx, _ := connection.Eth.GetTransactionByHash(os.Args[1])

	if jsonStu, err := json.Marshal(tx); err != nil {
		fmt.Println("Marshal fail")
	} else {
		fmt.Println(string(jsonStu))
	}
	params := make([]string, 1)
	params[0] = os.Args[1]
	pointer := &dto.RequestResult{}
	if err := connection.Provider.SendRequest(pointer, "eth_getRawTransactionByHash", params); err == nil {
		raw,_ := pointer.ToString()
		fmt.Println("raw",raw)
	}
	ext := web3ext.NewWeb3Ext(connection.Provider)
	if raw, err := ext.GetRawTransactionByHash(os.Args[1]); err != nil {
		fmt.Println(raw)
	} else {
		fmt.Println(err)
	}
}