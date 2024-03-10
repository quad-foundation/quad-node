package blocks

import (
	"bytes"
	"fmt"
	"github.com/quad/quad-node/common"
	"github.com/quad/quad-node/transactionsPool"
)

func CheckBlock(newBlock Block, lastBlock Block) (*transactionsPool.MerkleTree, error) {
	chain := newBlock.GetChain()
	blockHeight := newBlock.GetHeader().Height

	if common.CheckHeight(chain, blockHeight) == false {
		return nil, fmt.Errorf("improper height value in block")
	}

	if bytes.Compare(lastBlock.BlockHash.GetBytes(), newBlock.GetHeader().PreviousHash.GetBytes()) != 0 {
		return nil, fmt.Errorf("last block hash not much to store in new block")
	}
	// needs to check block and process
	if newBlock.CheckProofOfSynergy() == false {
		return nil, fmt.Errorf("proof of synergy fails of block")
	}
	hash, err := newBlock.CalcBlockHash()
	if err != nil {
		return nil, err
	}
	if bytes.Compare(hash.GetBytes(), newBlock.BlockHash.GetBytes()) != 0 {
		return nil, fmt.Errorf("wrong hash of block")
	}
	rootMarkleTrie := newBlock.GetHeader().RootMerkleTree
	txs := newBlock.TransactionsHashes
	txsBytes := make([][]byte, len(txs))
	for _, tx := range txs {
		hash := tx.GetBytes()
		txsBytes = append(txsBytes, hash)
	}
	merkleTrie, err := transactionsPool.BuildMerkleTree(blockHeight, txsBytes)
	if err != nil {
		return nil, err
	}
	if bytes.Compare(merkleTrie.GetRootHash(), rootMarkleTrie.GetBytes()) != 0 {
		return nil, fmt.Errorf("root merkleTrie hash check fails")
	}
	return merkleTrie, nil
}
