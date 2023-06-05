package blocks

import (
	"github.com/chainpqc/chainpqc-node/common"
)

type DexBlock struct {
	BaseBlock        BaseBlock   `json:"base_block"`
	Chain            uint8       `json:"chain"`
	TransactionsHash common.Hash `json:"transactions_hash"`
	BlockHash        common.Hash `json:"block_hash"`
}

func (tb DexBlock) GetBaseBlock() BaseBlock {
	return tb.BaseBlock
}
func (tb DexBlock) GetBlockHeaderHash() common.Hash {
	return tb.BaseBlock.BlockHeaderHash
}
func (tb DexBlock) GetBlockTimeStamp() int64 {
	return tb.BaseBlock.BlockTimeStamp
}
func (tb DexBlock) GetRewardPercentage() int16 {
	return tb.BaseBlock.RewardPercentage
}
func (tb DexBlock) GetChain() uint8 {
	return tb.Chain
}
func (tb DexBlock) GetTransactionsHash() common.Hash {
	return tb.TransactionsHash
}
func (tb DexBlock) GetBlockHash() common.Hash {
	return tb.BlockHash
}
func (tb DexBlock) GetFromBytes(b []byte) (AnyBlock, error) {
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
func (tb DexBlock) GetBytes() []byte {
	b := tb.BaseBlock.GetBytes()
	b = append(b, tb.Chain)
	b = append(b, tb.TransactionsHash.GetBytes()...)
	return b
}
func (tb DexBlock) CalcBlockHash() (common.Hash, error) {
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
func (tb DexBlock) CheckProofOfSynergy() bool {
	return CheckProofOfSynergy(tb.BaseBlock)
}
