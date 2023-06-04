package blocks

import (
	"github.com/chainpqc/chainpqc-node/blocks"
	"github.com/chainpqc/chainpqc-node/common"
)

type PubKeysBlock struct {
	BaseBlock        blocks.BaseBlock `json:"base_block"`
	Chain            uint8            `json:"chain"`
	TransactionsHash common.Hash      `json:"pub_keys_hash"`
	BlockHash        common.Hash      `json:"block_hash"`
}

func (tb PubKeysBlock) GetBaseBlock() blocks.BaseBlock {
	return tb.BaseBlock
}
func (tb PubKeysBlock) GetBlockHeaderHash() common.Hash {
	return tb.BaseBlock.BlockHeaderHash
}
func (tb PubKeysBlock) GetBlockTimeStamp() int64 {
	return tb.BaseBlock.BlockTimeStamp
}
func (tb PubKeysBlock) GetRewardPercentage() int16 {
	return tb.BaseBlock.RewardPercentage
}
func (tb PubKeysBlock) GetChain() uint8 {
	return tb.Chain
}
func (tb PubKeysBlock) GetTransactionsHash() common.Hash {
	return tb.TransactionsHash
}
func (tb PubKeysBlock) GetBlockHash() common.Hash {
	return tb.BlockHash
}
func (tb PubKeysBlock) GetBytes() []byte {
	b := tb.BaseBlock.GetBytes()
	b = append(b, tb.Chain)
	b = append(b, tb.TransactionsHash.GetBytes()...)
	return b
}
func (tb PubKeysBlock) CalcBlockHash() (common.Hash, error) {
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
func (tb PubKeysBlock) CheckProofOfSynergy() bool {
	return blocks.CheckProofOfSynergy(tb.BaseBlock)
}
