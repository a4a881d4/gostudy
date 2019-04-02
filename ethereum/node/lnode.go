package main

import (
  "github.com/ethereum/go-ethereum/p2p/enr"
  _ "github.com/ethereum/go-ethereum/p2p/enode"
  "os"
  "io"
  "fmt"
  "bytes"
  "crypto/ecdsa"
  "encoding/binary"

  _ "github.com/ethereum/go-ethereum/common"
  "github.com/ethereum/go-ethereum/rlp"

  "github.com/syndtr/goleveldb/leveldb"
  _ "github.com/syndtr/goleveldb/leveldb/errors"
  _ "github.com/syndtr/goleveldb/leveldb/iterator"
  "github.com/syndtr/goleveldb/leveldb/opt"
  _ "github.com/syndtr/goleveldb/leveldb/storage"
  "github.com/syndtr/goleveldb/leveldb/util"
  "net"
  "net/url"
  "strconv"
  "github.com/ethereum/go-ethereum/crypto"
)
// Keys in the node database.
const (
  dbVersionKey   = "version" // Version of the database to flush if changes
  dbNodePrefix   = "n:"      // Identifier to prefix node entries with
  dbLocalPrefix  = "local:"
  dbDiscoverRoot = "v4"

  // These fields are stored per ID and IP, the full key is "n:<ID>:v4:<IP>:findfail".
  // Use nodeItemKey to create those keys.
  dbNodeFindFails = "findfail"
  dbNodePing      = "lastping"
  dbNodePong      = "lastpong"
  dbNodeSeq       = "seq"

  // Local information is keyed by ID only, the full key is "local:<ID>:seq".
  // Use localItemKey to create those keys.
  dbLocalSeq = "seq"
)
// splitNodeKey returns the node ID of a key created by nodeKey.
func splitNodeKey(key []byte) (id ID, rest []byte) {
  if !bytes.HasPrefix(key, []byte(dbNodePrefix)) {
    return ID{}, nil
  }
  item := key[len(dbNodePrefix):]
  copy(id[:], item[:len(id)])
  return id, item[len(id)+1:]
}
// splitNodeItemKey returns the components of a key created by nodeItemKey.
func splitNodeItemKey(key []byte) (id ID, ip net.IP, field string) {
  id, key = splitNodeKey(key)
  // Skip discover root.
  if string(key) == dbDiscoverRoot {
    return id, nil, ""
  }
  key = key[len(dbDiscoverRoot)+1:]
  // Split out the IP.
  ip = net.IP(key[:16])
  if ip4 := ip.To4(); ip4 != nil {
    ip = ip4
  }
  key = key[16+1:]
  // Field is the remainder of key.
  field = string(key)
  return id, ip, field
}
type Secp256k1 ecdsa.PublicKey
type ID [32]byte
type Node struct {
  r  enr.Record
  id ID
}
func (n ID) String() string {
  return fmt.Sprintf("%x", n[:])
}
func nodeKey(id ID) []byte {
  key := append([]byte(dbNodePrefix), id[:]...)
  key = append(key, ':')
  key = append(key, dbDiscoverRoot...)
  return key
}
// Incomplete returns true for nodes with no IP address.
func (n *Node) Incomplete() bool {
  return n.IP() == nil
}

// Load retrieves an entry from the underlying record.
func (n *Node) Load(k enr.Entry) error {
  return n.r.Load(k)
}

// IP returns the IP address of the node.
func (n *Node) IP() net.IP {
  var ip net.IP
  n.Load((*enr.IP)(&ip))
  return ip
}

// UDP returns the UDP port of the node.
func (n *Node) UDP() int {
  var port enr.UDP
  n.Load(&port)
  return int(port)
}

// UDP returns the TCP port of the node.
func (n *Node) TCP() int {
  var port enr.TCP
  n.Load(&port)
  return int(port)
}

// Pubkey returns the secp256k1 public key of the node, if present.
func (n *Node) Pubkey() *ecdsa.PublicKey {
  var key ecdsa.PublicKey
  if n.Load((*Secp256k1)(&key)) != nil {
    return nil
  }
  return &key
}
func (v Secp256k1) ENRKey() string { return "secp256k1" }

// EncodeRLP implements rlp.Encoder.
func (v Secp256k1) EncodeRLP(w io.Writer) error {
  return rlp.Encode(w, crypto.CompressPubkey((*ecdsa.PublicKey)(&v)))
}

// DecodeRLP implements rlp.Decoder.
func (v *Secp256k1) DecodeRLP(s *rlp.Stream) error {
  buf, err := s.Bytes()
  if err != nil {
    return err
  }
  pk, err := crypto.DecompressPubkey(buf)
  if err != nil {
    return err
  }
  *v = (Secp256k1)(*pk)
  return nil
}

func (n *Node) v4URL() string {
  var (
    scheme enr.ID
    nodeid string
    key    ecdsa.PublicKey
  )
  n.Load(&scheme)
  n.Load((*Secp256k1)(&key))
  switch {
  case scheme == "v4" || key != ecdsa.PublicKey{}:
    nodeid = fmt.Sprintf("%x", crypto.FromECDSAPub(&key)[1:])
  default:
    nodeid = fmt.Sprintf("%s.%x", scheme, n.id[:])
  }
  u := url.URL{Scheme: "enode"}
  if n.Incomplete() {
    u.Host = nodeid
  } else {
    addr := net.TCPAddr{IP: n.IP(), Port: n.TCP()}
    u.User = url.User(nodeid)
    u.Host = addr.String()
    if n.UDP() != n.TCP() {
      u.RawQuery = "discport=" + strconv.Itoa(n.UDP())
    }
  }
  return u.String()
}
func main() {
  var (
    nodeDBVersionKey = []byte("version")
    nodeDBItemPrefix = []byte("n:")
  )
  opts := &opt.Options{OpenFilesCacheCapacity: 5}
  path := os.Args[1]
  db, err := leveldb.OpenFile(path, opts)
  if err != nil {
    fmt.Println("err", err)
  }
  blob, _ := db.Get(nodeDBVersionKey, nil)
  fmt.Println("Version", blob)
  
  it := db.NewIterator(util.BytesPrefix(nodeDBItemPrefix), nil)
  for it.Next() {
    id, ip, rest := splitNodeItemKey(it.Key())
    time, _ := db.Get(it.Key(),nil)
    if ip != nil {
      val, _ := binary.Varint(time)
      fmt.Println("id",id.String(),ip,rest,val)
    } else {
      blob, err := db.Get(nodeKey(id),nil)
      fmt.Println("find_node",id.String(),ip,rest)
      var n Node
      if err == nil {
        if err := rlp.DecodeBytes(blob, &n.r); err == nil {
          fmt.Println(n.v4URL())
        } else {
          fmt.Println(blob)
          fmt.Println(err)
        }  
      }
    }
  }
  db.Close()
}