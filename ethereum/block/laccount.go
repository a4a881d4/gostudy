package main

import (
  "os"
  "io"
  "fmt"
  "bytes"
  "strings"
  "math/big"
  "encoding/binary"
  "encoding/hex"
  "github.com/ethereum/go-ethereum/common"
  "github.com/ethereum/go-ethereum/crypto"
  
  "github.com/ethereum/go-ethereum/core/types"

  "github.com/ethereum/go-ethereum/rlp"

  "github.com/syndtr/goleveldb/leveldb"
  "github.com/syndtr/goleveldb/leveldb/opt"
)

var (
  // databaseVerisionKey tracks the current database version.
  databaseVerisionKey = []byte("DatabaseVersion")

  // Data item prefixes (use single byte to avoid mixing data types, avoid `i`, used for indexes).
  headerPrefix       = []byte("h") // headerPrefix + num (uint64 big endian) + hash -> header
  headerHashSuffix   = []byte("n") // headerPrefix + num (uint64 big endian) + headerHashSuffix -> hash
  headerNumberPrefix = []byte("H") // headerNumberPrefix + hash -> num (uint64 big endian)
  emptyRoot = common.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")
)

func encodeBlockNumber(number uint64) []byte {
  enc := make([]byte, 8)
  binary.BigEndian.PutUint64(enc, number)
  return enc
}

// headerKey = headerPrefix + num (uint64 big endian) + hash
func headerKey(number uint64, hash common.Hash) []byte {
  return append(append(headerPrefix, encodeBlockNumber(number)...), hash.Bytes()...)
}

// headerHashKey = headerPrefix + num (uint64 big endian) + headerHashSuffix
func headerHashKey(number uint64) []byte {
  return append(append(headerPrefix, encodeBlockNumber(number)...), headerHashSuffix...)
}

func keybytesToHex(str []byte) []byte {
  l := len(str)*2 + 1
  var nibbles = make([]byte, l)
  for i, b := range str {
    nibbles[i*2] = b / 16
    nibbles[i*2+1] = b % 16
  }
  nibbles[l-1] = 16
  return nibbles
}

func compactToHex(compact []byte) []byte {
  if len(compact) == 0 {
    return compact
  }
  base := keybytesToHex(compact)
  // delete terminator flag
  if base[0] < 2 {
    base = base[:len(base)-1]
  }
  // apply odd flag
  chop := 2 - base[0]&1
  return base[chop:]
}
// hasTerm returns whether a hex key has the terminator flag.
func hasTerm(s []byte) bool {
  return len(s) > 0 && s[len(s)-1] == 16
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
type Storage map[common.Hash]Account

var indices = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f", "[17]"}

type node interface {
  fstring(string) string
  cache() (hashNode, bool)
  canUnload(cachegen, cachelimit uint16) bool
  String() string
}

type (
  fullNode struct {
    Children [17]node // Actual trie node data to encode/decode (needs custom encoder)
    flags    nodeFlag
  }
  shortNode struct {
    Key   []byte
    Val   node
    flags nodeFlag
  }
  hashNode  []byte
  valueNode []byte
)

// EncodeRLP encodes a full node into the consensus RLP format.
func (n *fullNode) EncodeRLP(w io.Writer) error {
  return rlp.Encode(w, n.Children)
}

func (n *fullNode) copy() *fullNode   { copy := *n; return &copy }
func (n *shortNode) copy() *shortNode { copy := *n; return &copy }

// nodeFlag contains caching-related metadata about a node.
type nodeFlag struct {
  hash  hashNode // cached hash of the node (may be nil)
  gen   uint16   // cache generation counter
  dirty bool     // whether the node has changes that must be written to the database
}

// canUnload tells whether a node can be unloaded.
func (n *nodeFlag) canUnload(cachegen, cachelimit uint16) bool {
  return !n.dirty && cachegen-n.gen >= cachelimit
}

func (n *fullNode) canUnload(gen, limit uint16) bool  { return n.flags.canUnload(gen, limit) }
func (n *shortNode) canUnload(gen, limit uint16) bool { return n.flags.canUnload(gen, limit) }
func (n hashNode) canUnload(uint16, uint16) bool      { return false }
func (n valueNode) canUnload(uint16, uint16) bool     { return false }

func (n *fullNode) cache() (hashNode, bool)  { return n.flags.hash, n.flags.dirty }
func (n *shortNode) cache() (hashNode, bool) { return n.flags.hash, n.flags.dirty }
func (n hashNode) cache() (hashNode, bool)   { return nil, true }
func (n valueNode) cache() (hashNode, bool)  { return nil, true }

// Pretty printing.
func (n *fullNode) String() string  { return n.fstring("") }
func (n *shortNode) String() string { return n.fstring("") }
func (n hashNode) String() string   { return n.fstring("") }
func (n valueNode) String() string  { return n.fstring("") }

func (n *fullNode) fstring(ind string) string {
  resp := fmt.Sprintf("[\n%s  ", ind)
  for i, node := range n.Children {
    if node == nil {
      resp += fmt.Sprintf("%s: <nil> ", indices[i])
    } else {
      resp += fmt.Sprintf("%s: %v", indices[i], node.fstring(ind+"  "))
    }
  }
  return resp + fmt.Sprintf("\n%s] ", ind)
}

func (n *shortNode) fstring(ind string) string {
  return fmt.Sprintf("{%x: %v} ", n.Key, n.Val.fstring(ind+"  "))
}

func (n hashNode) fstring(ind string) string {
  return fmt.Sprintf("<%x> ", []byte(n))
}

func (n valueNode) fstring(ind string) string {
  return fmt.Sprintf("%x ", []byte(n))
}

// decodeNode parses the RLP encoding of a trie node.
func decodeNode(hash, buf []byte, cachegen uint16) (node, error) {
  if len(buf) == 0 {
    return nil, io.ErrUnexpectedEOF
  }
  elems, _, err := rlp.SplitList(buf)
  if err != nil {
    return nil, fmt.Errorf("decode error: %v", err)
  }
  switch c, _ := rlp.CountValues(elems); c {
    case 2:
      n, err := decodeShort(hash, buf, elems, cachegen)
      return n, wrapError(err, "short")
    case 17:
      n, err := decodeFull(hash, buf, elems, cachegen)
      return n, wrapError(err, "full")
    default:
      return nil, fmt.Errorf("invalid number of list elements: %v", c)
  }
}

func decodeShort(hash, buf, elems []byte, cachegen uint16) (node, error) {
  kbuf, rest, err := rlp.SplitString(elems)
  if err != nil {
    return nil, err
  }
  flag := nodeFlag{hash: hash, gen: cachegen}
  key := compactToHex(kbuf)
  if hasTerm(key) {
    // value node
    val, _, err := rlp.SplitString(rest)
    if err != nil {
      return nil, fmt.Errorf("invalid value node: %v", err)
    }
    return &shortNode{key, append(valueNode{}, val...), flag}, nil
  }
  r, _, err := decodeRef(rest, cachegen)
  if err != nil {
    return nil, wrapError(err, "val")
  }
  return &shortNode{key, r, flag}, nil
}

func decodeFull(hash, buf, elems []byte, cachegen uint16) (*fullNode, error) {
  n := &fullNode{flags: nodeFlag{hash: hash, gen: cachegen}}
  for i := 0; i < 16; i++ {
    cld, rest, err := decodeRef(elems, cachegen)
    if err != nil {
      return n, wrapError(err, fmt.Sprintf("[%d]", i))
    }
    n.Children[i], elems = cld, rest
  }
  val, _, err := rlp.SplitString(elems)
  if err != nil {
    return n, err
  }
  if len(val) > 0 {
    n.Children[16] = append(valueNode{}, val...)
  }
  return n, nil
}

const hashLen = len(common.Hash{})

func decodeRef(buf []byte, cachegen uint16) (node, []byte, error) {
  kind, val, rest, err := rlp.Split(buf)
  if err != nil {
    return nil, buf, err
  }
  switch {
  case kind == rlp.List:
    if size := len(buf) - len(rest); size > hashLen {
      err := fmt.Errorf("oversized embedded node (size is %d bytes, want size < %d)", size, hashLen)
      return nil, buf, err
    }
    n, err := decodeNode(nil, buf, cachegen)
    return n, rest, err
  case kind == rlp.String && len(val) == 0:
    // empty node
    return nil, rest, nil
  case kind == rlp.String && len(val) == 32:
    return append(hashNode{}, val...), rest, nil
  default:
    return nil, nil, fmt.Errorf("invalid RLP string size %d (want 0 or 32)", len(val))
  }
}

type decodeError struct {
  what  error
  stack []string
}

func wrapError(err error, ctx string) error {
  if err == nil {
    return nil
  }
  if decErr, ok := err.(*decodeError); ok {
    decErr.stack = append(decErr.stack, ctx)
    return decErr
  }
  return &decodeError{err, []string{ctx}}
}

func (err *decodeError) Error() string {
  return fmt.Sprintf("%v (decode path: %s)", err.what, strings.Join(err.stack, "<-"))
}

func decodeHash(db *leveldb.DB, hash []byte, depth int) (node, error) {
  val,err := db.Get(hash,nil)
  if err==nil {
    return decodeNode(hash,val,0) 
  } else {
    return nil,err
  }
}

func toString(s []byte) string {
  r := ""
  for _,n := range s {
    r = r+fmt.Sprintf("-%x",n)
  }
  return r
}

func dump(db *leveldb.DB,n node,depth int,s []byte,accounts Storage) {
  switch fn := n.(type) {
    case *fullNode:
      for i,h := range fn.Children {
        if h!=nil {
          fmt.Println(ws(depth)+"child ",toString(s))
          dump(db,h,depth+1,append(s,byte(i)),accounts)
        }
      }
    case *shortNode:
      fmt.Println(ws(depth)+"ShortNode",fn.String())
      k := hexToKeybytes(append(s,fn.Key...))
      fmt.Println(ws(depth)+"ShortNode Key",fmt.Sprintf("%x ",k))
      dump(db,fn.Val,depth+1,s,accounts)
      
      var account Account
      if val,ok := fn.Val.(valueNode); ok {
        if err := rlp.DecodeBytes(val, &account); err == nil {
          accounts[common.BytesToHash(k)] = account
        } else {
          buf := bytes.NewBuffer(val)
          s := rlp.NewStream(buf, 0)
          for {
            if err := rlpdump(db, s, depth+2); err != nil {
              if err != io.EOF {
              }
              break
            }
            fmt.Println()
          }
        }
      } 
    case hashNode:
      fmt.Println(ws(depth)+toString(s)+":hash Node",fn.String())
      dumpKey(db,fn,depth+1,s,accounts)
    case valueNode:
      fmt.Println(ws(depth)+toString(s)+":value Node",fn.String())
  }
}
func rlpdump(db *leveldb.DB, s *rlp.Stream, depth int) error {
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
        if err := rlpdump(db, s, depth+1); err == rlp.EOL {
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

func hexToKeybytes(hex []byte) []byte {
  if hasTerm(hex) {
    hex = hex[:len(hex)-1]
  }
  key := make([]byte, (len(hex)+1)/2)
  if len(hex)&1 != 0 {
    // fmt.Println("can't convert hex key of odd length",len(hex))
    nhex := make([]byte, len(hex)+1)
    copy(nhex[1:],hex)
    decodeNibbles(nhex, key)
  } else {
    decodeNibbles(hex, key)
  }
  return key
}

func decodeNibbles(nibbles []byte, bytes []byte) {
  for bi, ni := 0, 0; ni < len(nibbles); bi, ni = bi+1, ni+2 {
    bytes[bi] = nibbles[ni]<<4 | nibbles[ni+1]
  }
}

func dumpKey(db *leveldb.DB, hash []byte, depth int, s []byte,accounts Storage) error {
  n,err := decodeHash(db,hash,0)
  if n!=nil {
    fmt.Println("dumpKey",depth,n.String())
  }
  
  if err==nil {
    dump(db,n,depth,s,accounts)
  } 
  return err 
}

func ws(n int) string {
  return strings.Repeat("  ", n)
}

// go run ethereum/block/laccount.go ../ethereum/chain/geth/chaindata/ 951ac03f86ad5d42719963beb01498a07b40d81f

func main() {
  opts := &opt.Options{OpenFilesCacheCapacity: 5}
  path := os.Args[1]
  db, err := leveldb.OpenFile(path, opts)
  if err != nil {
    fmt.Println("err", err)
  }
  blob, _ := db.Get(databaseVerisionKey, nil)
  fmt.Println("Version", blob)

  var number uint64
  for number = 0;number<14041+1;number++ { //0x12d8c2 number=1304924 0x1272c2
    
    if blob,err := db.Get(headerHashKey(number),nil); err == nil {
      
      data, err := db.Get(headerKey(number, common.BytesToHash(blob)),nil)
      if err != nil {
        fmt.Println(number,"hash not in db",err)
      } 

      var h types.Header
      if err := rlp.DecodeBytes(data, &h); err == nil {
        // hjson,_ := h.MarshalJSON()
        // fmt.Println("===========Head of Block",number,"==================")
        // fmt.Println(string(hjson))
        accounts := make(Storage)
        if err := dumpKey(db,h.Root.Bytes(),0,[]byte{},accounts); err == nil {
          count := 0
          for k,v := range(accounts) {
            fmt.Printf("%3d. account(%x) =\n%s",count,k,v.String())
            count ++
          }
          address,_ := hex.DecodeString(os.Args[2])
          addHash := crypto.Keccak256Hash(address[:])
          if v,ok := accounts[addHash]; ok {
            fmt.Println(os.Args[2]," has ",v.Balance)
          } else {
            fmt.Printf("hash is %x ",addHash)
            fmt.Println("Cannot find account ",os.Args[2])
          }
          for k,v := range(accounts) {
            if v.Root != emptyRoot {
              if code,err := db.Get(v.CodeHash,nil); err==nil {
                fmt.Println("Contract code: ",k[:8],code[:8])
              }
              nop := make(Storage)
              dumpKey(db,v.Root.Bytes(),0,[]byte{},nop)
            }
          }
        }
      } 
    } 
  }
  db.Close()
}
