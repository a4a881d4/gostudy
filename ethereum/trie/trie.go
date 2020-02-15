package trie
import (
	"fmt"
	"io"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
)

var indices = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f", "[17]"}

type Node interface {
	fstring(string) string
	String() string
	Travel(fn func(Node))
}

type (
	FullNode struct {
		Children [17]Node // Actual trie node data to encode/decode (needs custom encoder)
		Self     HashNode
	}
	ShortNode struct {
		Key   []byte
		Val   Node
		Self  HashNode
	}
	HashNode  []byte
	ValueNode []byte
)

// nilValueNode is used when collapsing internal trie nodes for hashing, since
// unset children need to serialize correctly.
var NilValueNode = ValueNode(nil)

func (n *FullNode) String() string  { return n.fstring("") }
func (n *ShortNode) String() string { return n.fstring("") }
func (n HashNode) String() string   { return n.fstring("") }
func (n ValueNode) String() string  { return n.fstring("") }

func (n *FullNode) fstring(ind string) string {
	resp := fmt.Sprintf("[\n%s  ", ind)
	for i, node := range &n.Children {
		if node == nil {
			resp += fmt.Sprintf("%s: <nil> ", indices[i])
		} else {
			resp += fmt.Sprintf("%s: %v", indices[i], node.fstring(ind+"  "))
		}
	}
	return resp + fmt.Sprintf("\n%s] ", ind)
}

func (n *ShortNode) fstring(ind string) string {
	return fmt.Sprintf("{%x: %v} ", n.Key, n.Val.fstring(ind+"  "))
}

func (n HashNode) fstring(ind string) string {
	return fmt.Sprintf("<%x> ", []byte(n))
}

func (n ValueNode) fstring(ind string) string {
	return fmt.Sprintf("%x ", []byte(n))
}

func(n *FullNode) Travel(fn func(Node)) {
	fn(n)
	for _,v := range(n.Children) {
		if v != nil {
			v.Travel(fn)
		}
	}
}

func(n *ShortNode) Travel(fn func(Node)) {
	fn(n)
	n.Val.Travel(fn)
}

func(n HashNode) Travel(fn func(Node)) {
	fn(n)
}

func(n ValueNode) Travel(fn func(Node)) {
	fn(n)
}

func(n *ShortNode) HashKey() []byte {
	var s []byte
	return hexToKeybytes(append(s,n.Key...))
}

func DecodeNode(hash,buf []byte) (Node,error) {
	return decodeNode(hash,buf)
}

func decodeNode(hash, buf []byte) (Node, error) {
	if len(buf) == 0 {
		return nil, io.ErrUnexpectedEOF
	}
	elems, _, err := rlp.SplitList(buf)
	if err != nil {
		return nil, fmt.Errorf("decode error: %v", err)
	}
	switch c, _ := rlp.CountValues(elems); c {
	case 2:
		n, err := decodeShort(hash, elems)
		return n, wrapError(err, "short")
	case 17:
		n, err := decodeFull(hash, elems)
		return n, wrapError(err, "full")
	default:
		return nil, fmt.Errorf("invalid number of list elements: %v", c)
	}
}

func decodeShort(hash, elems []byte) (Node, error) {
	kbuf, rest, err := rlp.SplitString(elems)
	if err != nil {
		return nil, err
	}
	key := compactToHex(kbuf)
	if hasTerm(key) {
		// value node
		val, _, err := rlp.SplitString(rest)
		if err != nil {
			return nil, fmt.Errorf("invalid value node: %v", err)
		}
		return &ShortNode{key, append(ValueNode{}, val...), hash}, nil
	}
	r, _, err := decodeRef(rest)
	if err != nil {
		return nil, wrapError(err, "val")
	}
	return &ShortNode{key, r, hash}, nil
}

func decodeFull(hash, elems []byte) (*FullNode, error) {
	n := &FullNode{Self:hash}
	for i := 0; i < 16; i++ {
		cld, rest, err := decodeRef(elems)
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
		n.Children[16] = append(ValueNode{}, val...)
	}
	return n, nil
}

const hashLen = len(common.Hash{})

func decodeRef(buf []byte) (Node, []byte, error) {
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
		n, err := decodeNode(nil, buf)
		return n, rest, err
	case kind == rlp.String && len(val) == 0:
		// empty node
		return nil, rest, nil
	case kind == rlp.String && len(val) == 32:
		return append(HashNode{}, val...), rest, nil
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
