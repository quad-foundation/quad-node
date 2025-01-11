package pubkeys

import (
	"bytes"
	"fmt"
	"github.com/quad-foundation/quad-node/common"
	memDatabase "github.com/quad-foundation/quad-node/database"
	"log"
)

type MerkleTree struct {
	Root    []MerkleNode
	TxPK    []common.PubKey
	Address common.Address
	DB      *memDatabase.AnyBlockchainDB
}
type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Data  []byte
}

var GlobalMerkleTree *MerkleTree

func InitPermanentTrie() {
	merkleNodes, err := NewMerkleTree([]common.PubKey{})
	if err != nil {
		log.Fatalf(err.Error())
	}
	GlobalMerkleTree = new(MerkleTree)
	GlobalMerkleTree.Root = merkleNodes
	db := memDatabase.NewInMemoryDB()
	GlobalMerkleTree.DB = &db
}

func NewMerkleTree(data []common.PubKey) ([]MerkleNode, error) {
	var nodes []MerkleNode
	for _, pk := range data {
		node, err := NewMerkleNode(nil, nil, pk.GetBytes())
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, *node)
	}
	for len(nodes) > 1 {
		if len(nodes)%2 != 0 {
			nodes = append(nodes, nodes[len(nodes)-1])
		}
		var level []MerkleNode
		for i := 0; i < len(nodes); i += 2 {
			node, err := NewMerkleNode(&nodes[i], &nodes[i+1], nil)
			if err != nil {
				return nil, err
			}
			level = append(level, *node)
		}
		nodes = level
	}
	return nodes, nil
}
func NewMerkleNode(left, right *MerkleNode, data []byte) (*MerkleNode, error) {
	node := MerkleNode{}
	if left == nil && right == nil {
		hash, err := common.CalcHashToByte(data)
		if err != nil {
			log.Println("hash calculation fails")
			return nil, err
		}
		node.Data = hash[:]
	} else {
		prevHashes := append(left.Data, right.Data...)
		hash, err := common.CalcHashToByte(prevHashes)
		if err != nil {
			log.Println("hash calculation fails")
		}
		node.Data = hash[:]
	}
	node.Left = left
	node.Right = right
	return &node, nil
}

func (t *MerkleTree) IsPubKeyInTree(pk common.PubKey) bool {
	left, _ := t.Root[0].containsPK(0, pk.GetBytes())
	right, _ := t.Root[1].containsPK(0, pk.GetBytes())
	return left || right
}

func (n *MerkleNode) containsPK(index int64, hash []byte) (bool, int64) {
	index++
	if n == nil {
		return false, index - 1
	}
	if bytes.Equal(n.Data, hash) {
		return true, index - 1
	}

	left, _ := n.Left.containsPK(index, hash)
	right, indexRight := n.Right.containsPK(index, hash)
	contrains := left || right
	return contrains, indexRight
}

func (t *MerkleTree) GetRootHash() []byte {
	if len(t.Root) > 0 {
		return t.Root[0].Data
	}
	return common.EmptyHash().GetBytes()
}

func BuildMerkleTree(address common.Address, pubKeys []common.PubKey, db *memDatabase.AnyBlockchainDB) (*MerkleTree, error) {

	merkleNodes, _ := NewMerkleTree(pubKeys)
	tree := new(MerkleTree)
	tree.Root = merkleNodes
	tree.TxPK = pubKeys
	tree.DB = db
	tree.Address = address
	return tree, nil
}

func (tree *MerkleTree) StoreTree(address common.Address) error {

	merkleNodes := tree.Root
	prefix := common.PubKeyMerkleTrieDBPrefix[:]
	key := append(prefix, address.GetBytes()...)

	treeb, err := common.Marshal(merkleNodes, common.MerkleNodeDBPrefix)
	if err != nil {
		return err
	}
	err = (*tree.DB).Put(key, treeb)
	if err != nil {
		return err
	}
	prefix = common.PubKeyRootHashMerkleTreeDBPrefix[:]
	key = append(prefix, address.GetBytes()...)
	err = (*tree.DB).Put(key, tree.GetRootHash())
	if err != nil {
		return err
	}
	ret := []byte{}
	for _, pk := range tree.TxPK {
		ret = append(ret, common.BytesToLenAndBytes(pk.GetBytes())...)
	}
	prefix = common.PubKeyBytesMerkleTrieDBPrefix[:]
	key = append(prefix, address.GetBytes()...)
	err = (*tree.DB).Put(key, ret)
	if err != nil {
		return err
	}
	return nil
}

func LoadPubKeys(address common.Address) ([]common.PubKey, error) {
	prefix := common.PubKeyBytesMerkleTrieDBPrefix[:]
	key := append(prefix, address.GetBytes()...)
	pkbytes, err := (*GlobalMerkleTree.DB).Get(key)
	if err != nil {
		return nil, err
	}
	ret := []common.PubKey{}
	b := pkbytes[:]
	var pkb []byte
	for len(b) > 0 {
		pkb, b, err = common.BytesWithLenToBytes(b[:])
		if err != nil {
			return nil, err
		}
		pk := common.PubKey{}
		err = pk.Init(pkb[:])
		if err != nil {
			return nil, err
		}
		ret = append(ret, pk)
		if len(b) <= 0 {
			break
		}
	}
	return ret, nil
}

func LoadHashMerkleTreeByAddress(address common.Address) ([]byte, error) {
	prefix := common.PubKeyRootHashMerkleTreeDBPrefix[:]
	key := append(prefix, address.GetBytes()...)
	hash, err := (*GlobalMerkleTree.DB).Get(key)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

func (t *MerkleTree) Destroy() {
	if t != nil {
		t.Root = nil
		t.TxPK = nil
		t.DB = nil
	}
}

func LoadTreeWithoutPK(address common.Address) (*MerkleTree, error) {

	tree := new(MerkleTree)
	prefix := common.PubKeyMerkleTrieDBPrefix[:]
	key := append(prefix, address.GetBytes()...)
	treeb, err := (*GlobalMerkleTree.DB).Get(key)
	if err != nil {
		return &MerkleTree{}, err
	}
	var merkleNodes []MerkleNode
	err = common.Unmarshal(treeb, common.MerkleNodeDBPrefix, &merkleNodes)
	if err != nil {
		return &MerkleTree{}, err
	}
	tree.Root = merkleNodes

	prefix = common.PubKeyRootHashMerkleTreeDBPrefix[:]
	key = append(prefix, address.GetBytes()...)
	rootHash, err := (*GlobalMerkleTree.DB).Get(key)
	if err != nil {
		return &MerkleTree{}, err
	}
	tree.Root[0].Data = rootHash

	return tree, nil
}

func FindPubKeyForAddress(pk common.PubKey, address common.Address) (int64, error) {

	tree, err := LoadTreeWithoutPK(address)
	if err != nil {
		return -1, err
	}
	if len(tree.Root) == 0 {
		return -1, fmt.Errorf("no merkle tree root hash")
	}
	left, hl := tree.Root[0].containsPK(0, pk.GetBytes())
	right, hr := tree.Root[1].containsPK(0, pk.GetBytes())
	if left {
		return hl, nil
	}
	if right {
		return hr, nil
	}
	return -1, fmt.Errorf("pub key not found")
}

func LastIndexStoredInMerleTrie() (int64, error) {
	i := int64(0)
	for {
		ib := common.GetByteInt64(i)
		prefix := append(common.PubKeyRootHashMerkleTreeDBPrefix[:], ib...)
		isKey, err := (*GlobalMerkleTree.DB).IsKey(prefix)
		if err != nil {
			return i - 1, err
		}
		if isKey == false {
			break
		}
		i++
	}
	return i - 1, nil
}

func RemoveMerkleTrieFromDB(address common.Address) error {
	hb := address.GetBytes()
	prefix := append(common.PubKeyRootHashMerkleTreeDBPrefix[:], hb...)
	err := (*GlobalMerkleTree.DB).Delete(prefix)
	if err != nil {
		log.Println("cannot remove root merkle trie hash", err)
		return err
	}
	prefix = append(common.PubKeyMerkleTrieDBPrefix[:], hb...)
	err = (*GlobalMerkleTree.DB).Delete(prefix)
	if err != nil {
		log.Println("cannot remove merkle trie node", err)
		return err
	}
	prefix = append(common.PubKeyBytesMerkleTrieDBPrefix[:], hb...)
	err = (*GlobalMerkleTree.DB).Delete(prefix)
	if err != nil {
		log.Println("cannot remove merkle trie transaction hashes", err)
		return err
	}
	return nil
}
