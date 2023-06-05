package blocks

import (
	"github.com/chainpqc/chainpqc-node/common"
)

type ContractsBlock struct {
	BaseBlock        BaseBlock   `json:"base_block"`
	Chain            uint8       `json:"chain"`
	TransactionsHash common.Hash `json:"transactions_hash"`
	BlockHash        common.Hash `json:"block_hash"`
}

func (tb ContractsBlock) GetBaseBlock() BaseBlock {
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
func (tb ContractsBlock) GetFromBytes(b []byte) (AnyBlock, error) {
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
	return CheckProofOfSynergy(tb.BaseBlock)
}
