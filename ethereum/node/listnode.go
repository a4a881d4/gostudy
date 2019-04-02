package main

import (
	"fmt"
	_ "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/syndtr/goleveldb/leveldb"
	_ "github.com/syndtr/goleveldb/leveldb/errors"
	_ "github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	_ "github.com/syndtr/goleveldb/leveldb/storage"
	"github.com/syndtr/goleveldb/leveldb/util"
	_ "net"
	_ "net/url"
	"os"
	_ "strconv"

	"github.com/ethereum/go-ethereum/p2p/enr"
)

const NodeIDBits = 512
/*
type NodeID [NodeIDBits / 8]byte
type Node struct {
	IP       net.IP // len 4 for IPv4 or 16 for IPv6
	UDP, TCP uint16 // port numbers
	ID       NodeID // the node's public key

	// This is a cached copy of sha3(ID) which is used for node
	// distance calculations. This is part of Node in order to make it
	// possible to write tests that need a node at a certain distance.
	// In those tests, the content of sha will not actually correspond
	// with ID.
	sha common.Hash

	// whether this node is currently being pinged in order to replace
	// it in a bucket
	contested bool
}
*/
// Node represents a host on the network.
type ID [32]byte
type Node struct {
	r  enr.Record
	id ID
}

// func (n *Node) String() string {
// 	u := url.URL{Scheme: "enode"}
// 	if n.Incomplete() {
// 		u.Host = fmt.Sprintf("%x", n.ID[:])
// 	} else {
// 		addr := net.TCPAddr{IP: n.IP, Port: int(n.TCP)}
// 		u.User = url.User(fmt.Sprintf("%x", n.ID[:]))
// 		u.Host = addr.String()
// 		if n.UDP != n.TCP {
// 			u.RawQuery = "discport=" + strconv.Itoa(int(n.UDP))
// 		}
// 	}
// 	return u.String()
// }
// func (n *Node) Incomplete() bool {
// 	return n.IP == nil
// }
func main() {
	path := os.Args[1]
	var (
		nodeDBVersionKey = []byte("version")
		nodeDBItemPrefix = []byte("n:")
	)
	opts := &opt.Options{OpenFilesCacheCapacity: 5}
	db, err := leveldb.OpenFile(path, opts)
	if err != nil {
		fmt.Println("err", err)
	}
	blob, err := db.Get(nodeDBVersionKey, nil)
	fmt.Println("Version", blob)

	it := db.NewIterator(util.BytesPrefix(nodeDBItemPrefix), nil)
	for it.Next() {
		fmt.Println("key",it.Key())
		if blob, err := db.Get(it.Key(), nil); err != nil {
			fmt.Println(err)
			break
		} else {
			var n Node
			if err := rlp.DecodeBytes(blob, &n); err == nil {
				// fmt.Println(n.String())
			} else {
				fmt.Println(blob)
				fmt.Println(err)
			}
		}
		fmt.Println(".")
	}
	db.Close()
}
