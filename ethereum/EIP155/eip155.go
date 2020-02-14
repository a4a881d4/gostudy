package eip155

import (
	"math/big"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"golang.org/x/crypto/sha3"
)

type EIP155 struct {
	chainId, chainIdMul *big.Int
}

type Transaction struct {
	AccountNonce uint64
	Price        *big.Int
	GasLimit     uint64
	Recipient    *common.Address
	Amount       *big.Int
	Payload      []byte

	// Signature values
	// V *big.Int
	// R *big.Int
	// S *big.Int
	Signature
}

func NewEIP155(cID int64) EIP155 {
	return EIP155{
		chainId:    big.NewInt(cID),
		chainIdMul: big.NewInt(cID*2),
	}
}

func rlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewLegacyKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}

func NewTransaction(nonce uint64, to *common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *Transaction {
	if len(data) > 0 {
		data = common.CopyBytes(data)
	}
	d := Transaction{
		AccountNonce: nonce,
		Recipient:    to,
		Payload:      data,
		Amount:       new(big.Int),
		GasLimit:     gasLimit,
		Price:        new(big.Int),
	}
	d.R,d.S,d.V = new(big.Int), new(big.Int), new(big.Int)
	if amount != nil {
		d.Amount.Set(amount)
	}
	if gasPrice != nil {
		d.Price.Set(gasPrice)
	}

	return &d
}

func(s *EIP155) HashExt(tx *Transaction) common.Hash {
	return rlpHash([]interface{}{
		tx.AccountNonce,
		tx.Price,
		tx.GasLimit,
		tx.Recipient,
		tx.Amount,
		tx.Payload,
		s.chainId, uint(0), uint(0),
	})
}

func(s *EIP155) Hash(tx *types.Transaction) common.Hash {
	return rlpHash([]interface{}{
		tx.Nonce(),
		tx.GasPrice(),
		tx.Gas(),
		tx.To(),
		tx.Value(),
		tx.Data(),
		s.chainId, uint(0), uint(0),
	})
}

func(s *EIP155) V(iv int64) *big.Int {
	iv += 35
	v := big.NewInt(iv)
	v.Add(v,s.chainIdMul)
	return v
}

func(s *EIP155) IV(v *big.Int) int64 {
	t := new(big.Int).Set(v)
	t.Sub(t,s.chainIdMul)
	return t.Int64() - 35
}

