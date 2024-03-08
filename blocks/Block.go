package blocks

import (
	"fmt"
	"github.com/quad/quad-node/common"
	memDatabase "github.com/quad/quad-node/database"
)

type Block struct {
	BaseBlock        BaseBlock   `json:"base_block"`
	Chain            uint8       `json:"chain"`
	TransactionsHash common.Hash `json:"transactions_hashes"`
	BlockHash        common.Hash `json:"block_hash"`
}

func (tb Block) GetBaseBlock() BaseBlock {
	return tb.BaseBlock
}
func (tb Block) GetBlockHeaderHash() common.Hash {
	return tb.BaseBlock.BlockHeaderHash
}
func (tb Block) GetBlockTimeStamp() int64 {
	return tb.BaseBlock.BlockTimeStamp
}
func (tb Block) GetRewardPercentage() int16 {
	return tb.BaseBlock.RewardPercentage
}
func (tb Block) GetChain() uint8 {
	return tb.Chain
}
func (tb Block) GetTransactionsHash() common.Hash {
	return tb.TransactionsHash
}
func (tb Block) GetBlockHash() common.Hash {
	return tb.BlockHash
}
func (tb Block) GetBytes() []byte {
	b := tb.BaseBlock.GetBytes()
	b = append(b, tb.Chain)
	b = append(b, tb.TransactionsHash.GetBytes()...)
	b = append(b, tb.BlockHash.GetBytes()...)
	return b
}

func (tb Block) GetFromBytes(b []byte) (Block, error) {
	b, err := tb.BaseBlock.GetFromBytes(b)
	if err != nil {
		return Block{}, err
	}
	tb.Chain = b[0]
	tb.TransactionsHash = common.GetHashFromBytes(b[1:33])
	tb.BlockHash = common.GetHashFromBytes(b[33:65])
	if len(b) > 65 {
		return Block{}, fmt.Errorf("wrongly decompile block")
	}
	return tb, nil
}

func (tb Block) CalcBlockHash() (common.Hash, error) {
	toByte, err := common.CalcHashToByte(tb.GetBytes())
	if err != nil {
		return common.Hash{}, err
	}
	hash := common.GetHashFromBytes(toByte)
	return hash, nil
}

func (tb Block) CheckProofOfSynergy() bool {
	return CheckProofOfSynergy(tb.BaseBlock)
}

func (b Block) GetTransactionsHashes(tempMerkleTrie *transactionsPool.MerkleTree, height int64) ([]common.Hash, error) {
	txsHashes, err := tempMerkleTrie.LoadTransactionsHashes(height)
	if err != nil {
		return nil, err
	}
	hs := []common.Hash{}
	for _, hb := range txsHashes {
		eh := common.GetHashFromBytes(hb)
		hs = append(hs, eh)
	}
	return hs, nil
}

func (bl Block) StoreBlock() error {
	err := memDatabase.MainDB.Put(append(common.BlocksDBPrefix[:], bl.GetBlockHash().GetBytes()...), bl.GetBytes())
	if err != nil {
		return err
	}
	bh := common.GetByteInt64(bl.GetBaseBlock().BaseHeader.Height)
	err = memDatabase.MainDB.Put(append(common.BlockByHeightDBPrefix[:], bh...), bl.GetBlockHash().GetBytes())
	if err != nil {
		return err
	}

	return nil
}
func LoadHashOfBlock(height int64) ([]byte, error) {
	bh := common.GetByteInt64(height)
	hashb, err := memDatabase.MainDB.Get(append(common.BlockByHeightDBPrefix[:], bh...))
	if err != nil {
		return nil, err
	}
	return hashb, nil
}

func LoadBlock(height int64) (Block, error) {
	bh := common.GetByteInt64(height)
	hb, err := memDatabase.MainDB.Get(append(common.BlockByHeightDBPrefix[:], bh...))
	if err != nil {
		return Block{}, err
	}
	abl, err := memDatabase.MainDB.Get(append(common.BlocksDBPrefix[:], hb...))
	if err != nil {
		return Block{}, err
	}
	block := Block{}
	b, err := block.GetFromBytes(abl)
	if err != nil {
		return Block{}, err
	}
	return b, nil
}
