package transactionType

import (
	"bytes"
	"fmt"
	"github.com/chainpqc/chainpqc-node/common"
	memDatabase "github.com/chainpqc/chainpqc-node/database"
	"log"
)

type MerkleTree struct {
	Root *MerkleNode
}
type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Data  []byte
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
func (t *MerkleTree) IsHashInTree(hash []byte) (bool, int64) {
	return t.Root.containsHash(0, hash)
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

func BuildMerkleTree(blockMerkleTreeHash []byte, height int64, blockTransactionsHashes [][]byte) (MerkleNode, [][]byte, error) {
	blockTransactionsHashes = append([][]byte{blockMerkleTreeHash}, blockTransactionsHashes...)
	tree, err := NewMerkleTree(blockTransactionsHashes)
	if err != nil {
		return MerkleNode{}, nil, err
	}

	treeData := [][]byte{}
	for _, data := range tree {
		treeData = append(treeData, data.Data)
	}
	err = StoreMerkleTree(tree[0], height, treeData)
	if err != nil {
		return MerkleNode{}, nil, err
	}
	return tree[0], treeData, nil
}

func StoreMerkleTree(rootMarkleTree MerkleNode, height int64, treeHashes [][]byte) error {
	chain := common.GetChainForHeight(height)
	prefix := []byte{common.RootHashMerkleTreeDBPrefix[0], chain}

	key := append(prefix, common.GetByteInt64(height/5)...)
	err := memDatabase.Store(key, rootMarkleTree.Data)
	if err != nil {
		return err
	}
	prefix = []byte{common.MerkleTreeDBPrefix[0], chain}
	key = append(prefix, rootMarkleTree.Data...)
	treeData := []byte{}
	for _, data := range treeHashes {
		treeData = append(treeData, data...)
	}
	err = memDatabase.Store(key, treeData)
	if err != nil {
		return err
	}
	prefix = []byte{common.MerkleNodeDBPrefix[0], chain}
	key = append(prefix, rootMarkleTree.Data...)
	common.MerkleNodeDBPrefix[1] = chain
	mnb, err := common.Marshal(rootMarkleTree, common.MerkleNodeDBPrefix)
	if err != nil {
		return err
	}
	err = memDatabase.Store(key, mnb)
	if err != nil {
		return err
	}
	return nil
}

func LoadMerkleNode(rootNodehash []byte, chain uint8) (MerkleNode, error) {

	prefix := []byte{common.MerkleNodeDBPrefix[0], chain}
	key := append(prefix, rootNodehash...)
	mnb, err := memDatabase.Load(key)
	if err != nil {
		return MerkleNode{}, err
	}
	mn := MerkleNode{}
	common.MerkleNodeDBPrefix[1] = chain
	err = common.Unmarshal(mnb, common.MerkleNodeDBPrefix, &mn)
	if err != nil {
		return MerkleNode{}, err
	}
	return mn, nil
}

func LoadHashMerkleTreeByHeight(height int64) ([]byte, error) {
	chain := common.GetChainForHeight(height)
	prefix := []byte{common.RootHashMerkleTreeDBPrefix[0], chain}
	key := append(prefix, common.GetByteInt64(height/5)...)
	hash, err := memDatabase.Load(key)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

func LoadWholeMerkleTreeByHeight(height int64) ([][]byte, error) {
	hash, err := LoadHashMerkleTreeByHeight(height)
	if err != nil {
		return nil, err
	}
	chain := common.GetChainForHeight(height)
	prefix := []byte{common.MerkleTreeDBPrefix[0], chain}
	key := append(prefix, hash...)
	tree, err := memDatabase.Load(key)
	if err != nil {
		return nil, err
	}
	ret := [][]byte{}
	for i := 0; i < len(tree)/32; i++ {
		hash = tree[i*32 : (i+1)*32]
		ret = append(ret, hash)
	}
	return ret, nil
}

func FindTransactionInBlocks(targetHash []byte, height int64) (int64, error) {
	chain := common.GetChainForHeight(height)
	tree, err := LoadWholeMerkleTreeByHeight(height)
	if err != nil {
		return -1, err
	}
	node, err := LoadMerkleNode(tree[0], chain)
	if err != nil {
		return -1, err
	}
	exists, h := node.containsHash(height, targetHash)
	if exists {
		return h, nil
	}
	return -1, fmt.Errorf("hash of transaction not found")
}
