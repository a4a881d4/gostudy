package web3ext

import (
	"bytes"

	"encoding/hex"

	"github.com/regcostajr/go-web3/providers"
	"github.com/regcostajr/go-web3/dto"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/core/types"
)

type Web3Ext struct {
	provider providers.ProviderInterface
}

func NewWeb3Ext(provider providers.ProviderInterface) *Web3Ext {
	ext := new(Web3Ext)
	ext.provider = provider
	return ext
}

func(ext *Web3Ext) GetRawTransactionByHash(hash string) (string, error) {
	params := make([]string, 1)
	params[0] = hash
	pointer := &dto.RequestResult{}
	if err := ext.provider.SendRequest(pointer, "eth_getRawTransactionByHash", params); err == nil {
		return pointer.ToString()
	} else {
		return "", err
	}
}

func(ext *Web3Ext) GetTransactionByHash(hash string) (*types.Transaction, error) {
	if raw, err := ext.GetRawTransactionByHash(hash); err == nil {
		input, _ := hex.DecodeString(raw[2:])
		s := types.Transaction{}
		rlp.Decode(bytes.NewReader(input), &s)
		return &s, nil 
	} else {
		return nil, err
	}
}
