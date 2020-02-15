package main

import (
  "os"
  "fmt"
  "bytes"
  "encoding/binary"
  "math/big"

  "github.com/a4a881d4/gostudy/ethereum/trie"

  "github.com/syndtr/goleveldb/leveldb"
  "github.com/syndtr/goleveldb/leveldb/opt"

  "github.com/ethereum/go-ethereum/common"
  "github.com/ethereum/go-ethereum/core/types"
  "github.com/ethereum/go-ethereum/rlp"
)

var (

  emptyRoot = common.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")
 // Data item prefixes (use single byte to avoid mixing data types, avoid `i`, used for indexes).
  headerPrefix       = []byte("h") // headerPrefix + num (uint64 big endian) + hash -> header
  headerTDSuffix     = []byte("t") // headerPrefix + num (uint64 big endian) + hash + headerTDSuffix -> td
  headerHashSuffix   = []byte("n") // headerPrefix + num (uint64 big endian) + headerHashSuffix -> hash
  headerNumberPrefix = []byte("H") // headerNumberPrefix + hash -> num (uint64 big endian)

  blockBodyPrefix     = []byte("b") // blockBodyPrefix + num (uint64 big endian) + hash -> block body
  blockReceiptsPrefix = []byte("r") // blockReceiptsPrefix + num (uint64 big endian) + hash -> block receipts

  txLookupPrefix  = []byte("l") // txLookupPrefix + hash -> transaction/receipt lookup metadata
  bloomBitsPrefix = []byte("B") // bloomBitsPrefix + bit (uint16 big endian) + section (uint64 big endian) + hash -> bloom bits

  preimagePrefix = []byte("secure-key-")      // preimagePrefix + hash -> preimage
  configPrefix   = []byte("ethereum-config-") // config prefix for the db

  // Chain index prefixes (use `i` + single byte to avoid mixing data types).
  BloomBitsIndexPrefix = []byte("iB") // BloomBitsIndexPrefix is the data table of a chain indexer to track its progress

  Prefixes = map[string][]byte{
    "HD": headerPrefix,
    "TD": headerTDSuffix,
    "NB": headerHashSuffix,
    "HN": headerNumberPrefix,
    "BB": blockBodyPrefix,
    "BR": blockReceiptsPrefix,
    "LT": txLookupPrefix,
    "BM": bloomBitsPrefix,
    "CF": configPrefix,
    "IB": BloomBitsIndexPrefix,
  }
)

func perfix(key []byte) (string,[]byte) {

  for k,v := range(Prefixes) {
    if bytes.Compare(v,key[:len(v)]) == 0 {
      return k,key[len(v):]
    }
  }
  if len(key)==32 {
    return "ND",key
  }
  return "HX",key
}

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

func main() {

  opts := &opt.Options{OpenFilesCacheCapacity: 5}
  path := os.Args[1]
  db, err := leveldb.OpenFile(path, opts)
  if err != nil {
    fmt.Println("err", err)
  }
  iter := db.NewIterator(nil, nil)
  for iter.Next() {
    key := iter.Key()
    value := iter.Value()
    p,h := perfix(key)
    switch {
    case p == "CF":
      fmt.Println(p,common.Bytes2Hex(h),string(value))
    case p == "BB":
      dumpBody(h,value)
    case p == "ND":
      dumpNode(h,value)
    }
  }
}

func dumpBody(h,v []byte) {
  if len(v) < 10 {
    return
  }
  binary.BigEndian.Uint64(h[:8])
  fmt.Println("BB",binary.BigEndian.Uint64(h[:8]),common.Bytes2Hex(h[8:]))

  var body types.Body
  rlp.DecodeBytes(v,&body)
  for _,tx := range(body.Transactions) {
    if jtx,err := tx.MarshalJSON(); err != nil {
      fmt.Println(err)
    } else {
      fmt.Println(string(jtx))
    }
  }
}

func dumpNode(k,v []byte) {
  if n,err := trie.DecodeNode(k,v); err != nil {
    fmt.Println(err)
  } else {
    switch fn := n.(type) {
      case *trie.ShortNode:
        if val,ok := fn.Val.(trie.ValueNode); ok {
          var a Account
          rlp.DecodeBytes(val,&a)
          fmt.Println(common.Bytes2Hex(k))
          fmt.Println(a.String())
        } else {
          fmt.Println(n.String())
        }
      default: 
        fmt.Println(n.String())
    }
  }
}

