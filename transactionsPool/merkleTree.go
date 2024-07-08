package transactionsPool

import (
	"bytes"
	"fmt"
	"github.com/quad-foundation/quad-node/common"
	memDatabase "github.com/quad-foundation/quad-node/database"
	"log"
)

type MerkleTree struct {
	Root     []MerkleNode
	TxHashes [][]byte
	DB       *memDatabase.AnyBlockchainDB
}
type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Data  []byte
}

var GlobalMerkleTree *MerkleTree

func InitPermanentTrie() {
	merkleNodes, err := NewMerkleTree([][]byte{})
	if err != nil {
		log.Fatalf(err.Error())
	}
	GlobalMerkleTree = new(MerkleTree)
	GlobalMerkleTree.Root = merkleNodes
	db := memDatabase.NewInMemoryDB()
	GlobalMerkleTree.DB = &db
}

func NewMerkleTree(data [][]byte) ([]MerkleNode, error) {
	var nodes []MerkleNode
	for _, datum := range data {
		node, err := NewMerkleNode(nil, nil, datum)
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
func (t *MerkleTree) IsHashInTree(hash []byte) bool {
	left, _ := t.Root[0].containsHash(0, hash)
	right, _ := t.Root[1].containsHash(0, hash)
	return left || right
}
func (n *MerkleNode) containsHash(index int64, hash []byte) (bool, int64) {
	index++
	if n == nil {
		return false, index - 1
	}
	if bytes.Equal(n.Data, hash) {
		return true, index - 1
	}

	left, _ := n.Left.containsHash(index, hash)
	right, indexRight := n.Right.containsHash(index, hash)
	contrains := left || right
	return contrains, indexRight
}

func (t *MerkleTree) GetRootHash() []byte {
	if len(t.Root) > 0 {
		return t.Root[0].Data
	}
	return common.EmptyHash().GetBytes()
}

func BuildMerkleTree(height int64, blockTransactionsHashes [][]byte, db *memDatabase.AnyBlockchainDB) (*MerkleTree, error) {

	merkleNodes, _ := NewMerkleTree(blockTransactionsHashes)
	tree := new(MerkleTree)
	tree.Root = merkleNodes
	tree.TxHashes = blockTransactionsHashes
	tree.DB = db
	err := tree.StoreTree(height)
	if err != nil {
		return nil, err
	}
	return tree, nil
}

func (tree *MerkleTree) StoreTree(height int64) error {

	merkleNodes := tree.Root
	prefix := common.MerkleNodeDBPrefix[:]
	key := append(prefix, common.GetByteInt64(height)...)

	treeb, err := common.Marshal(merkleNodes, common.MerkleNodeDBPrefix)
	if err != nil {
		return err
	}
	err = (*tree.DB).Put(key, treeb)
	if err != nil {
		return err
	}
	prefix = common.RootHashMerkleTreeDBPrefix[:]
	key = append(prefix, common.GetByteInt64(height)...)
	err = (*tree.DB).Put(key, tree.GetRootHash())
	if err != nil {
		return err
	}
	ret := []byte{}
	for _, hash := range tree.TxHashes {
		ret = append(ret, hash...)
	}
	prefix = common.TransactionsHashesByHeightDBPrefix[:]
	key = append(prefix, common.GetByteInt64(height)...)
	err = (*tree.DB).Put(key, ret)
	if err != nil {
		return err
	}
	return nil
}

func LoadTransactionsHashes(height int64) ([][]byte, error) {
	prefix := common.TransactionsHashesByHeightDBPrefix[:]
	key := append(prefix, common.GetByteInt64(height)...)
	hashes, err := (*GlobalMerkleTree.DB).Get(key)
	if err != nil {
		return nil, err
	}
	ret := [][]byte{}
	for i := 0; i < len(hashes)/32; i++ {
		hash := hashes[i*32 : (i+1)*32]
		ret = append(ret, hash)
	}
	return ret, nil
}

func LoadHashMerkleTreeByHeight(height int64) ([]byte, error) {
	prefix := common.RootHashMerkleTreeDBPrefix[:]
	key := append(prefix, common.GetByteInt64(height)...)
	hash, err := (*GlobalMerkleTree.DB).Get(key)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

func (t *MerkleTree) Destroy() {
	if t != nil {
		t.Root = nil
		t.TxHashes = nil
		t.DB = nil
	}
}

func LoadTreeWithoutHashes(height int64) (*MerkleTree, error) {

	tree := new(MerkleTree)
	prefix := common.MerkleNodeDBPrefix[:]
	key := append(prefix, common.GetByteInt64(height)...)
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

	prefix = common.RootHashMerkleTreeDBPrefix[:]
	key = append(prefix, common.GetByteInt64(height)...)
	rootHash, err := (*GlobalMerkleTree.DB).Get(key)
	if err != nil {
		return &MerkleTree{}, err
	}
	tree.Root[0].Data = rootHash

	return tree, nil
}

func FindTransactionInBlocks(targetHash []byte, height int64) (int64, error) {

	tree, err := LoadTreeWithoutHashes(height)
	if err != nil {
		return -1, err
	}
	if len(tree.Root) == 0 {
		return -1, fmt.Errorf("no merkle tree root hash")
	}

	exists, h := tree.Root[0].containsHash(height, targetHash)
	if exists {
		return h, nil
	}
	return -1, fmt.Errorf("hash of transaction not found")
}

func LastHeightStoredInMerleTrie() (int64, error) {
	i := int64(0)
	for {
		ib := common.GetByteInt64(i)
		prefix := append(common.RootHashMerkleTreeDBPrefix[:], ib...)
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

func RemoveMerkleTrieFromDB(height int64) error {
	hb := common.GetByteInt64(height)
	prefix := append(common.RootHashMerkleTreeDBPrefix[:], hb...)
	err := (*GlobalMerkleTree.DB).Delete(prefix)
	if err != nil {
		log.Println("cannot remove root merkle trie hash", err)
		return err
	}
	prefix = append(common.MerkleNodeDBPrefix[:], hb...)
	err = (*GlobalMerkleTree.DB).Delete(prefix)
	if err != nil {
		log.Println("cannot remove merkle trie node", err)
		return err
	}
	prefix = append(common.TransactionsHashesByHeightDBPrefix[:], hb...)
	err = (*GlobalMerkleTree.DB).Delete(prefix)
	if err != nil {
		log.Println("cannot remove merkle trie transaction hashes", err)
		return err
	}
	return nil
}
