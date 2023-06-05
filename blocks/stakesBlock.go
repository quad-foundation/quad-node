package blocks

import (
	"github.com/chainpqc/chainpqc-node/common"
)

type StakesBlock struct {
	BaseBlock        BaseBlock   `json:"base_block"`
	Chain            uint8       `json:"chain"`
	TransactionsHash common.Hash `json:"stakes_hash"`
	BlockHash        common.Hash `json:"block_hash"`
}

func (tb StakesBlock) GetBaseBlock() BaseBlock {
	return tb.BaseBlock
}
func (tb StakesBlock) GetBlockHeaderHash() common.Hash {
	return tb.BaseBlock.BlockHeaderHash
}
func (tb StakesBlock) GetBlockTimeStamp() int64 {
	return tb.BaseBlock.BlockTimeStamp
}
func (tb StakesBlock) GetRewardPercentage() int16 {
	return tb.BaseBlock.RewardPercentage
}
func (tb StakesBlock) GetChain() uint8 {
	return tb.Chain
}
func (tb StakesBlock) GetTransactionsHash() common.Hash {
	return tb.TransactionsHash
}
func (tb StakesBlock) GetBlockHash() common.Hash {
	return tb.BlockHash
}
func (tb StakesBlock) GetBytes() []byte {
	b := tb.BaseBlock.GetBytes()
	b = append(b, tb.Chain)
	b = append(b, tb.TransactionsHash.GetBytes()...)
	return b
}
func (tb StakesBlock) GetFromBytes(b []byte) (AnyBlock, error) {
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
func (tb StakesBlock) CalcBlockHash() (common.Hash, error) {
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

func (tb StakesBlock) CheckProofOfSynergy() bool {
	return CheckProofOfSynergy(tb.BaseBlock)
}
