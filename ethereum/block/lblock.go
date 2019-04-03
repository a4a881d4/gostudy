package main

import (
  "os"
  _ "io"
  "fmt"
  "bytes"

  "encoding/binary"
  "github.com/ethereum/go-ethereum/common"
  "github.com/ethereum/go-ethereum/core/types"

  "github.com/ethereum/go-ethereum/rlp"

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
  for number = 0x12d8c2;number<0x12d8c2+1;number++ {
    if blob,err := db.Get(headerHashKey(number),nil); err == nil {
      hash := common.BytesToHash(blob)
//      fmt.Println(number,hash.Hex())
      data, _ := db.Get(headerKey(number, hash),nil)
      var h types.Header
      err := rlp.DecodeBytes(data, &h)

      if err == nil {
//        str,_ := h.MarshalJSON()
        // fmt.Println(string(str))
      }
      body := ReadBody(db,hash,number)
      if(len(body.Transactions)>0){
	str,_ := h.MarshalJSON()
        fmt.Println(string(str))
        str,_  = body.Transactions[0].MarshalJSON()
	fmt.Println(string(str))
      }
    }
  }
  db.Close()
}
