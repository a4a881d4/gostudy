package web3ext

import (
	"bytes"
	"errors"
	"encoding/hex"
	"math/big"
	"fmt"

	"github.com/regcostajr/go-web3"
	"github.com/regcostajr/go-web3/providers"
	"github.com/regcostajr/go-web3/dto"
	"github.com/regcostajr/go-web3/eth/block"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/a4a881d4/gostudy/ethereum/account/ecc"
	"github.com/a4a881d4/gostudy/ethereum/EIP155"
)

var (
	noEnoughCoin         = errors.New("No enough coin")
	UNPARSEABLEINTERFACE = errors.New("Unparseable Interface")
	WEBSOCKETNOTDENIFIED = errors.New("Websocket connection dont exist")
	defaultConfig        = NewConfig(big.NewInt(0x3b9aca00),0x61a80,931)
)

type Web3Ext struct {
	provider providers.ProviderInterface
	connection *web3.Web3
}
type Web3Config struct {
	Price        *big.Int
	GasLimit     uint64
	ChainID      int64
}
func NewConfig(price *big.Int, gasLimit uint64, cID int64) *Web3Config {
	R         := new(Web3Config)
	R.Price    = new(big.Int).Set(price)
	R.GasLimit = gasLimit
	R.ChainID  = cID
	return R
}
func NewWeb3Ext(provider providers.ProviderInterface) *Web3Ext {
	ext := new(Web3Ext)
	ext.provider   = provider
	ext.connection = web3.NewWeb3(provider)
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

func(ext *Web3Ext) SendRawTransaction(raw string) (string, error) {
	params := make([]string, 1)
	params[0] = raw
	pointer := &dto.RequestResult{}
	if err := ext.provider.SendRequest(pointer, "eth_sendRawTransaction", params); err == nil {
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

func(ext *Web3Ext) SendCoin(txC uint64, toString string, value *big.Int, keyString string, data []byte, config *Web3Config) (uint64,error) {
	if config == nil {
		config = defaultConfig
	}
	
	curve := ecc.NewSecp256K1()
	prK := big.NewInt(0)
	prK.SetString(keyString,16)	
	from := "0x"+curve.PrivateKey2Address(prK)

	
	bal,err := ext.connection.Eth.GetBalance(from, block.LATEST)
	if err != nil {
		return 0,err
	}
	
	if value.Cmp(bal) > 0 {
		return 0,noEnoughCoin
	}

	txCount,err := ext.connection.Eth.GetTransactionCount(from, block.LATEST)
	if err != nil {
		return 0,err
	}

	e155 := eip155.NewEIP155(config.ChainID)
	to := common.HexToAddress(toString)
	var tx *eip155.Transaction
	if txC != 0 {
		tx = eip155.NewTransaction(txC,&to,value,config.GasLimit,config.Price,data)
	} else {
		tx = eip155.NewTransaction(txCount.Uint64()+1,&to,value,config.GasLimit,config.Price,data)
	}
	
	// to := common.HexToAddress("0x0caebc448230a6f9a7c998aa8b452ec0ab02aef6")
	// tx := eip155.NewTransaction(0x12, 
	// 	&to,
	// 	big.NewInt(0).Mul(big.NewInt(1), big.NewInt(1E18)), 
	// 	config.GasLimit,config.Price, []byte("p2p transaction"))
	// d,_ := new(big.Int).SetString("2acbf03c5e393cde014666cb29a1079ab01f5429fb8e8cae36c873a48640a9c9",16)
	d := big.NewInt(0)
	hash := e155.HashExt(tx)
	
	r,s,v := curve.Sign(prK, d, hash.Bytes())
	tx.R,tx.S,tx.V = r,s,e155.V(v)
	// Q := curve.Recover(r,s,e155.IV(tx.V),hash.Bytes())
	// fmt.Println(curve.PublicKey2Address(Q))
	raw,_ := rlp.EncodeToBytes(tx)
	var etx types.Transaction
	rlp.DecodeBytes(raw,&etx)
	rawString := "0x"+common.Bytes2Hex(raw)
	rR,err := ext.SendRawTransaction(rawString)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(rR)
	}
	if json,err := etx.MarshalJSON(); err == nil {
		fmt.Println(string(json))
	} else {
		return txCount.Uint64(),err
	}
	return txCount.Uint64(),nil
}