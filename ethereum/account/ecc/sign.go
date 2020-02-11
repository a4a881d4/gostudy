package ecc

import (
	"fmt"
	"math/rand"
	"math/big"
)

var (
	  rnd = rand.New(rand.NewSource(99))
)

func zero() *big.Int {
	return big.NewInt(0)
}

func(curve *ECC) Sign(key, d *big.Int, hash []byte) (*big.Int,*big.Int) {
	
	D := zero().Set(d)
	
	if D.Cmp(zero()) == 0 {
		D.Rand(rnd, &curve.N)
	}
  
	dG := curve.PointScale(&curve.G,D)

	r := zero().Rem(&dG.X,&curve.N)

	iD := zero().ModInverse(D, &curve.N)

	e := zero().SetBytes(hash)

	s := zero().Mul(iD,zero().Add(e,zero().Mul(r,key)))

	s = zero().Rem(s, &curve.N)

	return r,s
}

func(curve *ECC) Verify(Q *ECPoint, r, s *big.Int, hash []byte) (bool,*big.Int) {
	
	e  := zero().SetBytes(hash)
	
	iS := zero().ModInverse(s, &curve.N)

	u1 := zero().Rem(zero().Mul(e,iS),&curve.N)

	u2 := zero().Rem(zero().Mul(r,iS),&curve.N)

	uG := curve.PointScale(&curve.G,u1)

	uQ := curve.PointScale(Q,u2)

	X  := curve.PointAdd(uG,uQ)

	vr := zero().Rem(&X.X,&curve.N)

	return (vr.Cmp(r) == 0),vr
}

func(curve *ECC) Verify2(Q *ECPoint, r, s *big.Int, hash []byte) (bool,*big.Int) {
	
	e  := zero().SetBytes(hash)
	
	iS := zero().ModInverse(s, &curve.N)

	// u1 := zero().Rem(zero().Mul(e,iS),&curve.N)

	// u2 := zero().Rem(zero().Mul(r,iS),&curve.N)

	uG := curve.PointScale(&curve.G,e)
	fmt.Println("e",e.Text(16))

	uQ := curve.PointScale(Q,r)

	S  := curve.PointAdd(uG,uQ)

	X  := curve.PointScale(S,iS)

	vr := zero().Rem(&X.X,&curve.N)

	return (vr.Cmp(r) == 0),vr
}
func(curve *ECC) Hack(key, r, s *big.Int, hash []byte) *big.Int {
	// d = (e+kr)/s
	e  := zero().SetBytes(hash)

	iS := zero().ModInverse(s, &curve.N)

	d  := zero().Add(e,zero().Mul(key,r))

	d  = zero().Mul(d,iS) 

	d  = zero().Rem(d,&curve.N)

	return d
}