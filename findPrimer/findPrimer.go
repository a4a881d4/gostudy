package main

import (
  "math/big"
  "fmt"
  "math/rand"
)

type Checker interface{
  check(d,q chan *big.Int)
}

type Fworker struct {
  x *big.Int
}

type Dworker struct {
  x *big.Int
}

func (c Dworker) check(d,q chan *big.Int) {
  for {
    toCheck := <-d
    y := big.NewInt(0).Mod(toCheck,c.x)
    if(y.Cmp(big.NewInt(0)) != 0) {
      q <- toCheck
    } 
  }
}

func (c Fworker) check(d,q chan *big.Int) {
  for {
    toCheck := <-d
    y := big.NewInt(0).Exp(c.x,big.NewInt(0).Sub(toCheck,big.NewInt(1)),toCheck)
    if(y.Cmp(big.NewInt(1)) == 0) {
      q <- toCheck
    }
  }  
}

func NewDWorker(x int64) *Dworker {
  return &Dworker{big.NewInt(x)}
}

func NewFWorker(x int64) *Fworker {
  return &Fworker{big.NewInt(x)}
}

func RandomBigInt(r *rand.Rand) *big.Int {
  return big.NewInt(0).Rand(r,big.NewInt(0).Lsh(big.NewInt(1),1024))
}

func main() {
  result := make(map[string]string)
  primers := []int64{2,3,5,7,11,13,17,19,23,29,31,37,41,43,47,53,59,61,67,71,73,79,83,89,97}
  size := len(primers)
  workers := make([]Checker,size*2)
  for i:=0;i<size;i++ {
    workers[i] = *NewDWorker(primers[i])
    workers[i+size] = *NewFWorker(primers[i])
  }
  ch := make(chan *big.Int,5)
  dch := ch
  r := rand.New(rand.NewSource(99))
  for i:=0;i<len(workers);i++ {
    och := make(chan *big.Int,5)
    go workers[i].check(dch,och)
    dch = och
  }

  rd := big.NewInt(2)
  for{
    select {
      case ch<-rd:
        rd = RandomBigInt(r)
        if(rd.ProbablyPrime(64)) {
          fmt.Print(".")
          // fmt.Println(rd.Text(16)," is a primer")
        }
      case r:=<-dch:
        fmt.Print("-")
        // fmt.Println(r.Text(16)," maybe a primer")
        result[r.String()] = ""
        ch<-big.NewInt(0).Add(r,big.NewInt(2))
        twin := big.NewInt(0).Sub(r,big.NewInt(2))
        if _, ok := result[twin.String()]; ok {
          result[twin.String()] = r.String()
          fmt.Printf("\nfind twins\n")
          fmt.Println(twin)
          fmt.Println(r)
        } 
    }
  }
}