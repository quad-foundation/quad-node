package blocks

import (
	"github.com/chainpqc/chainpqc-node/blocks"
	"github.com/chainpqc/chainpqc-node/common"
)

type ContractsBlock struct {
	BaseBlock        blocks.BaseBlock `json:"base_block"`
	Chain            uint8            `json:"chain"`
	TransactionsHash common.Hash      `json:"contracts_hash"`
	BlockHash        common.Hash      `json:"block_hash"`
}

func (tb ContractsBlock) GetBaseBlock() blocks.BaseBlock {
	return tb.BaseBlock
}
func (tb ContractsBlock) GetBlockHeaderHash() common.Hash {
	return tb.BaseBlock.BlockHeaderHash
}
func (tb ContractsBlock) GetBlockTimeStamp() int64 {
	return tb.BaseBlock.BlockTimeStamp
}
func (tb ContractsBlock) GetRewardPercentage() int16 {
	return tb.BaseBlock.RewardPercentage
}
func (tb ContractsBlock) GetChain() uint8 {
	return tb.Chain
}
func (tb ContractsBlock) GetTransactionsHash() common.Hash {
	return tb.TransactionsHash
}
func (tb ContractsBlock) GetBlockHash() common.Hash {
	return tb.BlockHash
}
func (tb ContractsBlock) GetBytes() []byte {
	b := tb.BaseBlock.GetBytes()
	b = append(b, tb.Chain)
	b = append(b, tb.TransactionsHash.GetBytes()...)
	return b
}
func (tb ContractsBlock) CalcBlockHash() (common.Hash, error) {
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
func (tb ContractsBlock) CheckProofOfSynergy() bool {
	return blocks.CheckProofOfSynergy(tb.BaseBlock)
}
