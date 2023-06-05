package blocks

import (
	"fmt"
	"github.com/chainpqc/chainpqc-node/common"
)

type PubKeysBlock struct {
	BaseBlock        BaseBlock   `json:"base_block"`
	Chain            uint8       `json:"chain"`
	TransactionsHash common.Hash `json:"pub_keys_hash"`
	BlockHash        common.Hash `json:"block_hash"`
}

func (tb PubKeysBlock) GetBaseBlock() BaseBlock {
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
func (tb PubKeysBlock) GetFromBytes(b []byte) (AnyBlock, error) {
	if len(b) < 33 {
		return nil, fmt.Errorf("not enough bytes for Pubkey Block unmarshal")
	}
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
	return CheckProofOfSynergy(tb.BaseBlock)
}
