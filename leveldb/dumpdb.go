package main

import (
  "fmt"
  "bytes"
  "flag"
  "encoding/binary"
  "math/big"
  "time"
  "os"
  "strings"

  "github.com/a4a881d4/gostudy/ethereum/trie"

  "github.com/syndtr/goleveldb/leveldb"
  "github.com/syndtr/goleveldb/leveldb/opt"

  "github.com/ethereum/go-ethereum/common"
  "github.com/ethereum/go-ethereum/core/types"
  "github.com/ethereum/go-ethereum/rlp"
)

var (
  dbDir          = flag.String("db", `E:\work-ref\copydb\chaindata`, "level db dir")
  blockBody      = flag.Bool("b", false, "dump Block body")
  chainConfig    = flag.Bool("c", false, "chain config")
  rootHash       = flag.Bool("r", false, "dump root hash")
  nodeInfo       = flag.Bool("n", false, "dump node info")
  accountInfo    = flag.Bool("a", false, "dump account info")
  internalInfo   = flag.Bool("i", false, "dump internal account info")
  headerBody     = flag.Bool("h", false, "dump block header")
  headerTime     = flag.Bool("t", false, "dump block time")
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
  Root     trie.HashNode
  CodeHash trie.HashNode
}

func(a *Account) isExt() bool {
  return common.BytesToHash(a.Root) == emptyRoot
}

func (a *Account) String() string {
  r := ""
  r += "[\n"
  r += fmt.Sprintf("\tNonce: %d\n",a.Nonce)
  r += fmt.Sprintf("\tBalance: %d\n",a.Balance)
  if !a.isExt() {
    r += fmt.Sprintf("\tRoot: %s\n",a.Root.String())
    r += fmt.Sprintf("\tCode: %s\n",a.CodeHash.String())
  }
  r += "]\n"
  return r
}

type RecordItem struct {
  node   trie.Node
  parent trie.Node
  index  int
}
type Storage map[[32]byte]*RecordItem 

func(i *RecordItem) Set(n trie.Node, idx int) {
  i.parent = n
  i.index = idx
}
// go run .\leveldb\dumpdb.go -c -r -db E:\work-ref\copydb\chaindata
func main() {
  flag.Parse()

  opts := &opt.Options{OpenFilesCacheCapacity: 5}
  path := *dbDir
  db, err := leveldb.OpenFile(path, opts)
  if err != nil {
    fmt.Println("err", err)
  }
  iter := db.NewIterator(nil, nil)

  var storage = make(Storage)

  for iter.Next() {
    key := iter.Key()
    value := iter.Value()
    p,h := perfix(key)
    switch {
    case p == "HD":
      if len(h) == 40 {
        if *headerBody {
          dumpHead(h,value)
        }
        if *headerTime {
          dumpHeadNumberWithTime(h,value)
        }
      }
    case p == "CF" && *chainConfig:
      fmt.Println(p,common.Bytes2Hex(h),string(value))
    case p == "BB" && *blockBody:
      dumpBody(h,value)
    case p == "ND":
      recordNode(h,value,storage)
      switch {
      case *nodeInfo:
        dumpNode(h,value)
      case *accountInfo:
        dumpAccount(h,value)
      case *internalInfo:
        intAccount(h,value)
      }      
    }
  }
  if *rootHash {
    roots,count := findRoot(storage)

    for k,v := range(roots) {
      fmt.Printf("%5d. ",k)
      fmt.Println(common.Bytes2Hex(v[:]))
    }
    fmt.Println("Total:",count)  
  }
}

func dumpBody(h,v []byte) {
  if len(v) < 10 {
    return
  }
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

func readHead(h,v []byte) (*types.Header, uint64, []byte, error) {
  if len(v) < 10 {
    return nil,0,[]byte{},fmt.Errorf("too small Header size")
  }
  bn   := binary.BigEndian.Uint64(h[:8])
  hash := h[8:]

  var header types.Header
  err := rlp.DecodeBytes(v,&header)
  return &header,bn,hash,err
}

func dumpHead(h,v []byte) {
  if header,bn,hash,err := readHead(h,v); err == nil {
    if json,err := header.MarshalJSON(); err != nil {
      fmt.Println(err)
    } else {
      fmt.Println("BB",bn,common.Bytes2Hex(hash))
      fmt.Println(string(json))
    }
  }
}

func dumpHeadNumberWithTime(h,v []byte) {
  if header,bn,_,err := readHead(h,v); err == nil {
    fmt.Println(bn,time.Unix(int64(header.Time),0))
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

func dumpAccount(k,v []byte) {
  if n,err := trie.DecodeNode(k,v); err != nil {
    fmt.Println(err)
  } else {
    switch fn := n.(type) {
    case *trie.ShortNode:
      if val,ok := fn.Val.(trie.ValueNode); ok {
        var a Account
        rlp.DecodeBytes(val,&a)
        if a.isExt() {
          fmt.Println(common.Bytes2Hex(k))
          fmt.Println(a.String())
        }
      } 
    }
  }
}

func intAccount(k,v []byte) {
  n,err := trie.DecodeNode(k,v); 
  if err != nil {
    fmt.Println(err)
    return
  } 
  switch fn := n.(type) {
  case *trie.ShortNode:
    if val,ok := fn.Val.(trie.ValueNode); ok {
      var a Account
      rlp.DecodeBytes(val,&a)
      if !a.isExt() && len(a.CodeHash)!=0 {
        fmt.Println(common.Bytes2Hex(k))
        fmt.Println(a.String())
      }
    } 
  }
}
func HashKey(k []byte) [32]byte {
  var r [32]byte
  length := 32
  if len(k) < 32 {
    length = len(k)
  }
  copy(r[:], k[:length])
  return r
}
func recordNode(k,v []byte, rec Storage) {
  if n,err := trie.DecodeNode(k,v); err != nil {
    fmt.Println(err)
  } else {
    rec[HashKey(k)] = &RecordItem{n,nil,-1}
  }
}

func findRoot(rec Storage) ([][32]byte,int) {
  var count int
  count = 0
  for _,v := range(rec) {
    count ++
    if fn,ok := v.node.(*trie.FullNode); ok {
      for idx,child := range(fn.Children) {
        if child != nil {
          key := HashKey(child.(trie.HashNode))
          if _,ok := rec[key]; ok {
            rec[key].Set(fn,idx)
            fn.Children[idx] = rec[key].node
          }          
        }
      }
    }
  }
  var ret [][32]byte
  for k,v := range(rec) {
    if v.index == -1 {
      ret = append(ret,k)
    }
  }
  return ret,count
}
func dumpBytes(val []byte) {
  r := bytes.NewReader(val)
  s := rlp.NewStream(r, 0)
  for {
    if err := dump(s, 0); err != nil {
      break
    }
    fmt.Println()
  }
}
func dump(s *rlp.Stream, depth int) error {
  kind, size, err := s.Kind()
  if err != nil {
    return err
  }
  switch kind {
  case rlp.Byte, rlp.String:
    str, err := s.Bytes()
    if err != nil {
      return err
    }
    if len(str) == 0 || isASCII(str) {
      fmt.Printf("%s%q", ws(depth), str)
    } else {
      fmt.Printf("%s%x", ws(depth), str)
    }
  case rlp.List:
    s.List()
    defer s.ListEnd()
    if size == 0 {
      fmt.Print(ws(depth) + "[]")
    } else {
      fmt.Println(ws(depth) + "[")
      for i := 0; ; i++ {
        if i > 0 {
          fmt.Print(",\n")
        }
        if err := dump(s, depth+1); err == rlp.EOL {
          break
        } else if err != nil {
          return err
        }
      }
      fmt.Print(ws(depth) + "]")
    }
  }
  return nil
}

func isASCII(b []byte) bool {
  for _, c := range b {
    if c < 32 || c > 126 {
      return false
    }
  }
  return true
}

func ws(n int) string {
  return strings.Repeat("  ", n)
}

func die(args ...interface{}) {
  fmt.Fprintln(os.Stderr, args...)
  os.Exit(1)
}