package web3ext

import (
	_ "fmt"
	"bytes"

	"github.com/regcostajr/go-web3/dto"
	"github.com/regcostajr/go-web3/complex/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
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

func (contract *ContractExt) Call(v interface{}, tx *dto.TransactionParameters, functionName string, args ...interface{}) (interface{}, error) {

	data,err := contract.Abi.Pack(functionName,args...)
	if err != nil {
		return nil,err
	} 
	tx.Data = types.ComplexString("0x" + common.Bytes2Hex(data))

	ret,err := contract.Web3.connection.Eth.Call(tx)
	if err != nil {
		return nil,err
	}

	bRet,_ := ret.ToComplexString()
	byteRet := common.Hex2Bytes(bRet.ToHex()[2:])
	err = contract.Abi.Unpack(v,functionName,byteRet) 

	return v,err
}

func (contract *ContractExt) Deploy(tx *dto.TransactionParameters, bytecode string, args ...interface{}) (string, error) {

	data,err := contract.Abi.Pack("",args...)
	if err != nil {
		return "",err
	} 
	tx.Data = types.ComplexString(bytecode + common.Bytes2Hex(data))

	return contract.Web3.connection.Eth.SendTransaction(tx)
}

func (contract *ContractExt) Do(tx *dto.TransactionParameters, functionName string, args ...interface{}) (string, error) {

	data,err := contract.Abi.Pack(functionName,args...)
	if err != nil {
		return "",err
	} 
	tx.Data = types.ComplexString(data)

	return contract.Web3.connection.Eth.SendTransaction(tx)
}