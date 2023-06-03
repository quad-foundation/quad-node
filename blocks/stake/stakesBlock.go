package block

import (
	"github.com/chainpqc/chainpqc-node/blocks"
	"github.com/chainpqc/chainpqc-node/common"
)

type StakesBlock struct {
	BaseBlock  blocks.BaseBlock `json:"base_block"`
	Chain      uint8            `json:"chain"`
	StakesHash common.Hash      `json:"stakes_hash"`
	BlockHash  common.Hash      `json:"block_hash"`
}
