package main

import (
  "os"
  "io"
  "fmt"
  "bytes"
  "strings"
  _ "encoding/hex"
  _ "golang.org/x/crypto/sha3"

  "encoding/binary"
  "github.com/ethereum/go-ethereum/common"
  _ "github.com/ethereum/go-ethereum/crypto"
  
  _ "github.com/ethereum/go-ethereum/common/hexutil"
  
  "github.com/ethereum/go-ethereum/core/types"

  "github.com/ethereum/go-ethereum/rlp"
  "github.com/ethereum/go-ethereum/trie"

  "github.com/syndtr/goleveldb/leveldb"
  "github.com/syndtr/goleveldb/leveldb/opt"
  _ "github.com/syndtr/goleveldb/leveldb/util"
)

var (
  // databaseVerisionKey tracks the current database version.
  databaseVerisionKey = []byte("DatabaseVersion")

  // headHeaderKey tracks the latest know header's hash.
  headHeaderKey = []byte("LastHeader")

  // headBlockKey tracks the latest know full block's hash.
  headBlockKey = []byte("LastBlock")

  // headFastBlockKey tracks the latest known incomplete block's hash during fast sync.
  headFastBlockKey = []byte("LastFast")

  // fastTrieProgressKey tracks the number of trie entries imported during fast sync.
  fastTrieProgressKey = []byte("TrieSync")

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

// headerTDKey = headerPrefix + num (uint64 big endian) + hash + headerTDSuffix
func headerTDKey(number uint64, hash common.Hash) []byte {
  return append(headerKey(number, hash), headerTDSuffix...)
}

// headerHashKey = headerPrefix + num (uint64 big endian) + headerHashSuffix
func headerHashKey(number uint64) []byte {
  return append(append(headerPrefix, encodeBlockNumber(number)...), headerHashSuffix...)
}

// headerNumberKey = headerNumberPrefix + hash
func headerNumberKey(hash common.Hash) []byte {
  return append(headerNumberPrefix, hash.Bytes()...)
}

// blockBodyKey = blockBodyPrefix + num (uint64 big endian) + hash
func blockBodyKey(number uint64, hash common.Hash) []byte {
  return append(append(blockBodyPrefix, encodeBlockNumber(number)...), hash.Bytes()...)
}

// blockReceiptsKey = blockReceiptsPrefix + num (uint64 big endian) + hash
func blockReceiptsKey(number uint64, hash common.Hash) []byte {
  return append(append(blockReceiptsPrefix, encodeBlockNumber(number)...), hash.Bytes()...)
}

// txLookupKey = txLookupPrefix + hash
func txLookupKey(hash common.Hash) []byte {
  return append(txLookupPrefix, hash.Bytes()...)
}

// bloomBitsKey = bloomBitsPrefix + bit (uint16 big endian) + section (uint64 big endian) + hash
func bloomBitsKey(bit uint, section uint64, hash common.Hash) []byte {
  key := append(append(bloomBitsPrefix, make([]byte, 10)...), hash.Bytes()...)

  binary.BigEndian.PutUint16(key[1:], uint16(bit))
  binary.BigEndian.PutUint64(key[3:], section)

  return key
}

// preimageKey = preimagePrefix + hash
func preimageKey(hash common.Hash) []byte {
  return append(preimagePrefix, hash.Bytes()...)
}

func preimageKeyByte(hash []byte) []byte {
  return append(preimagePrefix, hash...)
}

// configKey = configPrefix + hash
func configKey(hash common.Hash) []byte {
  return append(configPrefix, hash.Bytes()...)
}
func ReadBodyRLP(db *leveldb.DB, hash common.Hash, number uint64) rlp.RawValue {
  data, _ := db.Get(blockBodyKey(number, hash),nil)
  return data
}
// ReadBody retrieves the block body corresponding to the hash.
func ReadBody(db *leveldb.DB, hash common.Hash, number uint64) *types.Body {
  data := ReadBodyRLP(db, hash, number)
  if len(data) == 0 {
    return nil
  }
  body := new(types.Body)
  if err := rlp.Decode(bytes.NewReader(data), body); err != nil {
    return nil
  }
  return body
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
var indices = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f", "[17]"}

type node interface {
  fstring(string) string
  cache() (hashNode, bool)
  canUnload(cachegen, cachelimit uint16) bool
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

func mustDecodeNode(hash, buf []byte, cachegen uint16) node {
  n, err := decodeNode(hash, buf, cachegen)
  if err != nil {
    panic(fmt.Sprintf("node %x: %v", hash, err))
  }
  return n
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
    // 'embedded' node reference. The encoding must be smaller
    // than a hash in order to be valid.
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

// wraps a decoding error with information about the path to the
// invalid child node (for debugging encoding issues).
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
// secureKeyPrefix is the database key prefix used to store trie node preimages.
var secureKeyPrefix = []byte("secure-key-")

// secureKeyLength is the length of the above prefix + 32byte hash.
const secureKeyLength = 11 + 32
// headerNumberKey = headerNumberPrefix + hash
func SecureKey(hash common.Hash) []byte {
  return append(secureKeyPrefix, hash.Bytes()...)
}

func decodeHash(db *leveldb.DB, hash []byte, depth int) (node, error) {
  val,err := db.Get(hash,nil)
  if err==nil {
    return decodeNode(hash,val,0) 
  } else {
    return nil,err
  }
}

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
  for number = 0x1279e7;number<0x1279e7+1;number++ { //0x12d8c2 number=1304924 0x1272c2
    if blob,err := db.Get(headerHashKey(number),nil); err == nil {
      hash := common.BytesToHash(blob)
      data, _ := db.Get(headerKey(number, hash),nil)
      var h types.Header
      err := rlp.DecodeBytes(data, &h)
      if err == nil {
        body := ReadBody(db,hash,number)
        if(len(body.Transactions)>0){
          var txs types.Transactions
          txs = body.Transactions
          tr := Derive(txs)
          hash := types.DeriveSha(txs)
          dumpTrie(tr,hash.Bytes())
          // it := trie.NewIterator(tr.NodeIterator(nil))
          // for it.Next() {
          //   fmt.Printf("%x\n",it.Value)
          // }
          // fmt.Printf("%x-%x\n",hashT.Bytes(),h.TxHash.Bytes())
          str,_ := h.MarshalJSON()
          fmt.Println(string(str))
          for _,t := range(body.Transactions) {
            str,_  = t.MarshalJSON()
            fmt.Println(string(str))
          }
        }
      }
    }
  }
  db.Close()
}

func dumpBlob(db *leveldb.DB, blob []byte) {
  buf := bytes.NewBuffer(blob)
  s := rlp.NewStream(buf, 0)
  for {
    if err := rlpdump(db, s, 0); err != nil {
      if err != io.EOF {
      }
      break
    }
    fmt.Println()
  }
}

func toString(s []byte) string {
  r := ""
  for _,n := range s {
    r = r+fmt.Sprintf("-%x",n)
  }
  return r
}
func dump(db *leveldb.DB,n node,depth int,s []byte) {
  switch fn := n.(type) {
  case *fullNode:
    for i,h := range fn.Children {
      if h!=nil {
        fmt.Println(ws(depth)+"child ",toString(s))
        dump(db,h,depth+1,append(s,byte(i)))
      }
    }
  case *shortNode:
    fmt.Println(ws(depth)+"ShortNode",fn.String())
    k := hexToKeybytes(append(s,fn.Key...))
    fmt.Println(ws(depth)+"ShortNode Key",fmt.Sprintf("%x ",k))
    dump(db,fn.Val,depth+1,s)
  case hashNode:
    fmt.Println(ws(depth)+toString(s)+":hash Node",fn.String())
    dumpKey(db,fn,depth+1,s)
  case valueNode:
    fmt.Println(ws(depth)+toString(s)+":value Node",fn.String())
    buf := bytes.NewBuffer(fn)
    s := rlp.NewStream(buf, 0)
    for {
      if err := rlpdump(db, s, 0); err != nil {
        if err != io.EOF {
        }
        break
      }
      fmt.Println()
    }
  }
}

func hexToKeybytes(hex []byte) []byte {
  if hasTerm(hex) {
    hex = hex[:len(hex)-1]
  }
  if len(hex)&1 != 0 {
    fmt.Println("can't convert hex key of odd length",len(hex))
    return []byte{0}
  }
  key := make([]byte, (len(hex)+1)/2)
  decodeNibbles(hex, key)
  return key
}

func decodeNibbles(nibbles []byte, bytes []byte) {
  for bi, ni := 0, 0; ni < len(nibbles); bi, ni = bi+1, ni+2 {
    bytes[bi] = nibbles[ni]<<4 | nibbles[ni+1]
  }
}
func dumpKey(db *leveldb.DB, hash []byte, depth int, s []byte) error {
  n,err := decodeHash(db,hash,0)
  if err==nil {
    dump(db,n,depth,s)
  } 
  return err 
}

// func rlpDumpKey()
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
      // dumpKey(db,str,depth)
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

func ws(n int) string {
  return strings.Repeat("  ", n)
}

func Derive(list types.DerivableList) *trie.Trie {
  keybuf := new(bytes.Buffer)
  trie := new(trie.Trie)
  for i := 0; i < list.Len(); i++ {
    keybuf.Reset()
    rlp.Encode(keybuf, uint(i))
    trie.Update(keybuf.Bytes(), list.GetRlp(i))
  }
  return trie
}

func dumpTrie(t *trie.Trie, root []byte) {
  if blob,err := t.TryGet([]byte{0x80}); err == nil {
    if node,err := decodeNode(root,blob,0); err == nil {
      fmt.Println(node.fstring(""))
    } else {
      fmt.Println("node ",err,blob)
    }
  } else {
    fmt.Println("blob ",err)
  }
}