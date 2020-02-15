package main

import (
  "os"
  "fmt"
  "math/big"
  "encoding/hex"

  "github.com/syndtr/goleveldb/leveldb"
  "github.com/syndtr/goleveldb/leveldb/opt"

  "github.com/ethereum/go-ethereum/common"
  "github.com/ethereum/go-ethereum/crypto"
  "github.com/ethereum/go-ethereum/rlp"
 
)

var (
  emptyRoot = common.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")
)

type Account struct {
  Nonce    uint64
  Balance  *big.Int
  Root     common.Hash // merkle root of the storage trie
  CodeHash []byte
}

func (a *Account) String() string {
  r := ""
  r += "[\n"
  r += fmt.Sprintf("\tNonce: %d\n",a.Nonce)
  r += fmt.Sprintf("\tBalance: %d\n",a.Balance)
  if a.Root != emptyRoot {
    r += fmt.Sprintf("\tRoot: %x\n",a.Root.Bytes())
    r += fmt.Sprintf("\tCode: %x\n",a.CodeHash)
  }
  r += "]\n"
  return r
}

// go run .\leveldb\dumpLastState.go E:\work-ref\copydb\chaindata 951ac03f86ad5d42719963beb01498a07b40d81f
func main() {

  opts := &opt.Options{OpenFilesCacheCapacity: 5}
  path := os.Args[1]
  db, err := leveldb.OpenFile(path, opts)
  if err != nil {
    fmt.Println(err)
  }

  address,_ := hex.DecodeString(os.Args[2])
  addHash := crypto.Keccak256Hash(address[:])
  if v,err := db.Get(addHash.Bytes(),nil); err != nil {
    fmt.Println(err)
  } else {
    var a Account
    err = rlp.DecodeBytes(v,&a)
    fmt.Println(a.String())
  }

} 
          