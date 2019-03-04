package main

import (
	"fmt"
  "math/big"
  "golang.org/x/crypto/sha3"
  "encoding/hex"
  "os"
)

func main() {
  curve := NewSecp256K1()
  prK := new(big.Int)
  prK.SetString(os.Args[1],16)
  fmt.Println(curve.privateKey2Address(prK))
}

type ECPoint struct {
  X big.Int
  Y big.Int
}

type ECC struct {
  P big.Int
  A big.Int
  G ECPoint
}

func NewSecp256K1() (*ECC) {
  var ret = new(ECC)
  ret.P.SetString("0FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F",16)
  ret.A.SetString("7",16)
  x := new(big.Int)
  x.SetString("79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798",16)
  ret.G = *(ret.NewPoint(x))
  return ret 
}

func(p *ECPoint) Bytes() []byte {
  var buf = make([]byte,64,64)
  for i:=0;i<64;i++ {
    buf[i] = 0
  }
  copy(buf,(&p.X).Bytes())
  copy(buf[32:],(&p.Y).Bytes())
  return buf
}
func(curve *ECC) FieldNormal(a *big.Int) (*big.Int) {
  r := new(big.Int)
  r.Add(a,&curve.P)
  r.Rem(r,&curve.P)
  return r
}

func(curve *ECC) FieldAdd(a, b *big.Int) (*big.Int) {
  c := new(big.Int)
  c.Add(a,b)
  return curve.FieldNormal(c)
}

func(curve *ECC) FieldSub(a, b *big.Int) (*big.Int) {
  c := new(big.Int)
  c.Sub(a,b)
  return curve.FieldNormal(c)
}

func(curve *ECC) FieldMul(a, b *big.Int) (*big.Int) {
  c := new(big.Int)
  c.Mul(a,b)
  return curve.FieldNormal(c)
}

func(curve *ECC) FieldDiv(a, b *big.Int) (*big.Int) {
  z := curve.Inverse(b)
  c := new(big.Int).Mul(a,z)
  return curve.FieldNormal(c)
}

func(curve *ECC) Sqrt(x *big.Int) (*big.Int) {
  s := big.NewInt(1)
  s.Add(s,&curve.P)
  s.Rsh(s,2)
  r := new(big.Int)
  r.Set(x)
  r.Exp(r,s,&curve.P)
  return r
}

func(curve *ECC) Inverse(x *big.Int) (*big.Int) {
  s := new(big.Int).Set(&curve.P)
  s.Sub(s,big.NewInt(2))
  r := new(big.Int).Set(x)
  r.Exp(r,s,&curve.P)
  return r
}

func(curve *ECC) NewPoint(x *big.Int) (*ECPoint) {
  p := new(ECPoint)
  p.X = *x
  c := curve.FieldMul(x,x)
  c = curve.FieldMul(c,x)
  c = curve.FieldAdd(c,&curve.A)
  p.Y = *(curve.Sqrt(c))
  return p   
}

func(curve *ECC) PointAdd(Q,P *ECPoint) (*ECPoint) {
  if(P == nil) {
    return Q
  }
  if(Q == nil) {
    return P
  }
  if(Q.X.Cmp(&P.X) == 0 && Q.Y.Cmp(&P.Y) != 0) {
    return nil
  }
  if(Q.X.Cmp(&P.X) == 0 && Q.Y.Cmp(&P.Y) == 0) {
    return curve.PointTwice(Q)
  }
  lamda := curve.FieldDiv(curve.FieldSub(&Q.Y,&P.Y),
    curve.FieldSub(&Q.X,&P.X))
  R := new(ECPoint)
  R.X = *(curve.FieldSub(
      curve.FieldMul(lamda,lamda),
      curve.FieldAdd(&Q.X,&P.X)))
  R.Y = *(curve.FieldSub(
      curve.FieldMul(lamda,curve.FieldSub(&P.X,&R.X)),
      &P.Y))
  return R
}

func(curve *ECC) PointTwice(Q *ECPoint) (*ECPoint) {
  if(Q == nil) {
    return nil
  }
  lamda := curve.FieldDiv(
      curve.FieldMul(curve.FieldMul(&Q.X,&Q.X),big.NewInt(3)),
      curve.FieldMul(&Q.Y,big.NewInt(2)))
  R := new(ECPoint)
  R.X = *(curve.FieldSub(
      curve.FieldMul(lamda,lamda),
      curve.FieldAdd(&Q.X,&Q.X)))
  R.Y = *(curve.FieldSub(
      curve.FieldMul(lamda,curve.FieldSub(&Q.X,&R.X)),
      &Q.Y))
  return R
}

func(curve *ECC) PointScale(Q *ECPoint, N *big.Int) (*ECPoint) {
  var R *ECPoint
  n := new(big.Int)
  n.Set(N)
  g := &curve.G
  for {
    if(big.NewInt(0).Cmp(n) == 0) {
      break
    }
    nbit := n.Bits()
    if((nbit[0]&1) == 1) {
      R = curve.PointAdd(R,g)
    }
    g = curve.PointTwice(g)
    n.Rsh(n,1)
  }
  return R
}

func(curve *ECC) privateKey2Address(key *big.Int) (string) {
  puK := curve.PointScale(&curve.G,key)
  var hasher = sha3.NewLegacyKeccak256()
  hasher.Write(puK.Bytes())
  addr := hasher.Sum(nil)[12:]
  return hex.EncodeToString(addr)
}