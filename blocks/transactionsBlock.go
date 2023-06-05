package blocks

import (
	"github.com/chainpqc/chainpqc-node/common"
)

type TransactionsBlock struct {
	BaseBlock        BaseBlock   `json:"base_block"`
	Chain            uint8       `json:"chain"`
	TransactionsHash common.Hash `json:"transaction_hashes"`
	BlockHash        common.Hash `json:"block_hash"`
}

func (tb TransactionsBlock) GetBaseBlock() BaseBlock {
	return tb.BaseBlock
}
func (tb TransactionsBlock) GetBlockHeaderHash() common.Hash {
	return tb.BaseBlock.BlockHeaderHash
}
func (tb TransactionsBlock) GetBlockTimeStamp() int64 {
	return tb.BaseBlock.BlockTimeStamp
}
func (tb TransactionsBlock) GetRewardPercentage() int16 {
	return tb.BaseBlock.RewardPercentage
}
func (tb TransactionsBlock) GetChain() uint8 {
	return tb.Chain
}
func (tb TransactionsBlock) GetTransactionsHash() common.Hash {
	return tb.TransactionsHash
}
func (tb TransactionsBlock) GetBlockHash() common.Hash {
	return tb.BlockHash
}
func (tb TransactionsBlock) GetBytes() []byte {
	b := tb.BaseBlock.GetBytes()
	b = append(b, tb.Chain)
	b = append(b, tb.TransactionsHash.GetBytes()...)
	return b
}

func (tb TransactionsBlock) GetFromBytes(b []byte) (AnyBlock, error) {
	b, err := tb.BaseBlock.GetFromBytes(b)
	if err != nil {
		return nil, err
	}
	tb.Chain = b[0]
	hash, err := common.GetHashFromBytes(b[1:33])
	if err != nil {
		return nil, err
	}
	tb.TransactionsHash = hash
	return AnyBlock(tb), nil
}

func (tb TransactionsBlock) CalcBlockHash() (common.Hash, error) {
	toByte, err := common.CalcHashToByte(tb.GetBytes())
	if err != nil {
		return common.Hash{}, err
	}
	hash := common.Hash{}
	hash, err = hash.Init(toByte)
	if err != nil {
		return common.Hash{}, err
	}
	return hash, nil
}

func (tb TransactionsBlock) CheckProofOfSynergy() bool {
	return CheckProofOfSynergy(tb.BaseBlock)
}
