package main

import (
  "os"
  "fmt"
  "errors"
  "encoding/binary"
  "encoding/hex"
  "math/big"

  "github.com/a4a881d4/gostudy/ethereum/trie"

  "github.com/syndtr/goleveldb/leveldb"
  "github.com/syndtr/goleveldb/leveldb/opt"

  "github.com/ethereum/go-ethereum/common"
  _ "github.com/ethereum/go-ethereum/core/types"
  "github.com/ethereum/go-ethereum/rlp"
  "github.com/ethereum/go-ethereum/crypto"
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
//go run .\leveldb\dumpState.go E:\work-ref\copydb\chaindata f1edcee46a4b5c651581e3abd4e080bac14d7cb0c3f80fde463a2975ff5861b3
func main() {

	opts := &opt.Options{OpenFilesCacheCapacity: 5}
	path := os.Args[1]
	db, err := leveldb.OpenFile(path, opts)
	
	defer func(){
		db.Close()
	}()

	if err != nil {
		fmt.Println("err", err)
		fmt.Println(err)
		panic("open db failure")
	}
	
	Root := common.HexToHash(os.Args[2]) 
	if _,err = db.Get(Root.Bytes(),nil); err == nil {
		root,err := dumpKey(db,Root.Bytes())
		if err == nil {
			root.Travel(printNode())
			root.Travel(dumpAccount())
		} else {
			fmt.Println(err)
		}
	}	
}

func dumpAccount() func(trie.Node) {
	return func(node trie.Node) {
		var a Account
		if vn,ok := node.(trie.ValueNode); ok {
			if err := rlp.DecodeBytes(vn,&a); err == nil {
				fmt.Println(a.String())
			}
		}	
	}
}

func findHash(root *trie.FullNode, addr string, a *Account) {
	address,_ := hex.DecodeString(addr)
	addHash := crypto.Keccak256Hash(address[:])
	if n,err := root.Find(addHash.Bytes()); err!=nil {
		fmt.Println(err)
	} else {
		sn := n.(*trie.ShortNode)
		rlp.DecodeBytes(sn.Val.(trie.ValueNode),a)
	}
}

func printNode() func(trie.Node) {
	return func(node trie.Node){
		fmt.Println(node.String())
	}
}

func Accounts(addr string, a *Account) func(trie.Node,[]byte) {
	address,_ := hex.DecodeString(addr)
	addHash := crypto.Keccak256Hash(address[:])
	addHashStr := common.Bytes2Hex(addHash.Bytes())
	return func(node trie.Node,args []byte) {
		if vn,ok := node.(trie.ValueNode); ok {
			hashStr := common.Bytes2Hex(trie.HexToKey(args))
			if hashStr == addHashStr {
				rlp.DecodeBytes(vn,a)
			}
		}
	}
}

func findAccount(addr string, a *Account) func(trie.Node) {
	address,_ := hex.DecodeString(addr)
	addHash := crypto.Keccak256Hash(address[:])
	addHashStr := common.Bytes2Hex(addHash.Bytes())
	cut := addHashStr[len(addHashStr)-20:]
	r := func(n trie.Node) {
		if sn,ok := n.(*trie.ShortNode); ok {
			keyStr := common.Bytes2Hex(sn.HashKey())
			if len(keyStr)>20 {
				keyCut := keyStr[len(keyStr)-20:]
				if  cut == keyCut {
					rlp.DecodeBytes(sn.Val.(trie.ValueNode),a)
				}
			}
		}
	}
	return r
}

func hash2Node(db *leveldb.DB, hash []byte) (trie.Node,error) {
	v,err := db.Get(hash,nil)
	if err != nil {
		return trie.NilValueNode, err
	}
	return trie.DecodeNode(hash,v)
}

func dumpKey(db *leveldb.DB, hash []byte) (trie.Node, error) {
	root,err := hash2Node(db,hash)
	if err != nil {
		return root,err
	}
	var Errors []error
  	dump := func(n trie.Node) {
  		switch n.(type) {
		case *trie.FullNode:
		fn,_ := n.(*trie.FullNode)
		for k,v := range(fn.Children) {
			if hn,ok := v.(trie.HashNode); ok {
				fn.Children[k],err = hash2Node(db,[]byte(hn))
				if err != nil {
					fmt.Println(err)
					Errors = append(Errors,err)
				} 
			}
		}
		case *trie.ShortNode:
		sn,_ := n.(*trie.ShortNode)
		if hn,ok := sn.Val.(trie.HashNode); ok {
			sn.Val,err = hash2Node(db,[]byte(hn))
			if err != nil {
				fmt.Println(err)
				Errors = append(Errors,err)
			}
		}
		case trie.HashNode:
			fmt.Println("Can not dump into HashNode")
			Errors = append(Errors,errors.New("dump into hash"))
		default:
			return
  		}
  	}
  	root.Travel(dump)
  	Es := ""
  	for _,v := range(Errors){
  		Es += fmt.Sprintln(v)
  	}
  	if len(Errors)==0 {
  		return root,nil
  	} else {
  		return root,errors.New(Es)
  	}
}
