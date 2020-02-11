package web3ext

import (
	"github.com/regcostajr/go-web3/providers"
	"github.com/regcostajr/go-web3/dto"
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
