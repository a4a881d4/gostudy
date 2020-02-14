package web3ext

import (
	_ "fmt"
	"bytes"

	_ "github.com/regcostajr/go-web3/dto"
	"github.com/ethereum/go-ethereum/accounts/abi"
)

type ContractExt struct {
	Abi  *abi.ABI
	Web3 *Web3Ext 
}

func(ext *Web3Ext) NewContract(abiJson []byte) (*ContractExt, error) {
	r := bytes.NewReader(abiJson)
	a, err := abi.JSON(r)
	if err != nil {
		return nil,err
	} else {
		return &ContractExt{
			Abi: &a,
			Web3: ext,
		}, nil
	}
}

/*
func (contract *ContractExt) Call(tx *eip155.Transaction, functionName string, args ...interface{}) (*dto.RequestResult, error) {

	data,err := contract.Abi.Pack(functionName,args)
	if err != nil {
		return nil,err
	}

	tx.Payload = data
	return contract.Web3.Call(transaction)
}
*/