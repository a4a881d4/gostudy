package ecc

import (
	"math/rand"
	"math/big"
)

var (
	  rnd = rand.New(rand.NewSource(99))
)

func zero() *big.Int {
	return big.NewInt(0)
}

func(curve *ECC) Sign(key, d *big.Int, hash []byte) (*big.Int,*big.Int, int64) {
	
	D := zero().Set(d)
	
	if D.Cmp(zero()) == 0 {
		D.Rand(rnd, &curve.N)
	}
 	secp256k1halfN := new(big.Int).Div(&curve.N, big.NewInt(2)) 
	dG := curve.PointScale(&curve.G,D)

	r := zero().Set(&dG.X)

	r.Mod(r,&curve.N)

	iD := zero().ModInverse(D, &curve.N)

	e := zero().SetBytes(hash)

	s:= zero().Mul(key,r)

	s.Add(s,e)

	s.Mul(s,iD)

	s.Mod(s,&curve.N)

	v := int64(dG.Y.Bit(0))

	if s.Cmp(secp256k1halfN) > 0 {
		return curve.Sign(key,zero(),hash)
	}

	return r,s,v
}

func(curve *ECC) Verify(Q *ECPoint, r, s *big.Int, hash []byte) bool {
	
	e  := zero().SetBytes(hash)
	
	iS := zero().ModInverse(s, &curve.N)

	u1 := zero().Mul(e,iS)
	u1.Mod(u1,&curve.N)

	u2 := zero().Mul(r,iS)
	u2.Mod(u2,&curve.N)

	uG := curve.PointScale(&curve.G,u1)

	uQ := curve.PointScale(Q,u2)

	X  := curve.PointAdd(uG,uQ)

	vr := zero().Mod(&X.X,&curve.N)

	return (vr.Cmp(r) == 0)
}

func(curve *ECC) Hack(key, r, s *big.Int, hash []byte) *big.Int {
	// d = (e+kr)/s
	e  := zero().SetBytes(hash)

	iS := zero().ModInverse(s, &curve.N)

	d  := zero().Mul(key,r)

	d.Add(d,e)

	d.Mul(d,iS) 

	d.Mod(d,&curve.N)

	return d
}
// v = 0 even, v = 1 odd
func(curve *ECC) Recover(r,s *big.Int, v int64, hash []byte) *ECPoint {
	e  := zero().SetBytes(hash)

	iR := zero().ModInverse(r, &curve.N)
	
	Q := curve.NewPoint(r)
	if v != int64(Q.Y.Bit(0)) {
		Q.Y = *(big.NewInt(0).Sub(&curve.P,&Q.Y))
	}

	e   = e.Mul(e,iR)
	is := zero().Mul(s,iR)

	eG := curve.PointScale(&curve.G,e)
	sQ := curve.PointScale(Q,is)

	R := curve.PointSub(sQ,eG)

	return R
}