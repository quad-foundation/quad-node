package transactionsPool

import (
	"bytes"
	"github.com/quad-foundation/quad-node/common"
	memDatabase "github.com/quad-foundation/quad-node/database"
	"log"
)

type MerkleTree struct {
	Root []MerkleNode
	DB   *memDatabase.AnyBlockchainDB
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

func BuildMerkleTree(height int64, blockTransactionsHashes [][]byte) (*MerkleTree, error) {

	merkleNodes, err := NewMerkleTree(blockTransactionsHashes)
	prefix := common.MerkleNodeDBPrefix[:]
	key := append(prefix, common.GetByteInt64(height)...)

	tree := new(MerkleTree)
	tree.Root = merkleNodes
	db := memDatabase.NewInMemoryDB()
	tree.DB = &db
	treeb, err := common.Marshal(merkleNodes, common.MerkleNodeDBPrefix)
	if err != nil {
		return nil, err
	}
	err = (*tree.DB).Put(key, treeb)
	if err != nil {
		return nil, err
	}
	prefix = common.RootHashMerkleTreeDBPrefix[:]
	key = append(prefix, common.GetByteInt64(height)...)
	err = (*tree.DB).Put(key, tree.GetRootHash())
	if err != nil {
		return nil, err
	}
	ret := []byte{}
	for _, hash := range blockTransactionsHashes {
		ret = append(ret, hash...)
	}
	prefix = common.TransactionsHashesByHeightDBPrefix[:]
	key = append(prefix, common.GetByteInt64(height)...)
	err = (*tree.DB).Put(key, ret)
	if err != nil {
		return nil, err
	}
	return tree, nil
}

func (tree *MerkleTree) StoreTree(height int64, blockTransactionsHashes [][]byte) error {

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
	for _, hash := range blockTransactionsHashes {
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

func (tree *MerkleTree) LoadTransactionsHashes(height int64) ([][]byte, error) {
	prefix := common.TransactionsHashesByHeightDBPrefix[:]
	key := append(prefix, common.GetByteInt64(height)...)
	hashes, err := (*tree.DB).Get(key)
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

func (tree *MerkleTree) LoadMerkleNode(rootNodehash []byte, chain uint8) (MerkleNode, error) {

	prefix := common.MerkleNodeDBPrefix[:]
	key := append(prefix, rootNodehash...)
	mnb, err := (*tree.DB).Get(key)
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

func (tree *MerkleTree) LoadHashMerkleTreeByHeight(height int64) ([]byte, error) {
	prefix := common.RootHashMerkleTreeDBPrefix[:]
	key := append(prefix, common.GetByteInt64(height)...)
	hash, err := (*tree.DB).Get(key)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

func (t *MerkleTree) Destroy() {
	if t != nil {
		t.Root = nil
		t.DB = nil
	}
}

//func (tree *MerkleTree) LoadWholeMerkleTreeByHeight(height int64) ([][]byte, error) {
//	hash, err := tree.LoadHashMerkleTreeByHeight(height)
//	if err != nil {
//		return nil, err
//	}
//	prefix := common.MerkleTreeDBPrefix[:]
//	key := append(prefix, common.GetByteInt64(height)...)
//	mt, err := (*tree.db).Get(key)
//	if err != nil {
//		return nil, err
//	}
//	ret := [][]byte{}
//	for i := 0; i < len(tree.D)/32; i++ {
//		hash = tree[i*32 : (i+1)*32]
//		ret = append(ret, hash)
//	}
//	return ret, nil
//}

//func FindTransactionInBlocks(targetHash []byte, height int64) (int64, error) {
//	chain := common.GetChainForHeight(height)
//	tree, err := LoadWholeMerkleTreeByHeight(height)
//	if err != nil {
//		return -1, err
//	}
//	if len(tree) == 0 {
//		return -1, fmt.Errorf("no merkle tree root hash")
//	}
//	node, err := LoadMerkleNode(tree[0], chain)
//
//	if err != nil {
//		return -1, err
//	}
//	exists, h := node.containsHash(height, targetHash)
//	if exists {
//		return h, nil
//	}
//	return -1, fmt.Errorf("hash of transaction not found")
//}
