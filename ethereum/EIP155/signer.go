package eip155

import (
	"math/big"
	"github.com/a4a881d4/gostudy/ethereum/account/ecc"

)

type Signature struct{
	R,S,V *big.Int
}

type Signer interface {
	Sign(hash []byte) *Signature
	Verify(si *Signature, hash []byte) (bool,string)
}

type EIP155Signer struct{
	c    *ecc.ECC
	e    *EIP155
	prK  *big.Int
	puK  *ecc.ECPoint
}

func(e155 *EIP155) NewEIP155Signer(ec *ecc.ECC, key string) *EIP155Signer {
	return &EIP155Signer{
		c:   ecc,
		prK: big.NewInt(0).SetString(keyString,16),
		e: e155,	
	}
}

func(s *EIP155Signer) BuildKey() {
	s.puK = s.c.PointScale(s.c.G,s.prK)
}

func(s *EIP155Signer) Sign(hash []byte) *Signature {
	r,s,v := s.c.Sign(s.prK,big.NewInt(0))
	return &Signature{
		R: r,
		S: s,
		V: 
	}
}

func(s *EIP155Signer) Verify(si *Signature, hash []byte) (bool,string) {
	s.prK = big.NewInt(0)
	s.puK = s.c.Recover(r,s,v,hash)
	ok := s.c.Verify(s.puK, r, s)
	return ok,"0x"+s.c.PublicKey2Address(s.puK)
}