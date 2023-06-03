package block

import (
	"github.com/chainpqc/chainpqc-node/blocks"
	"github.com/chainpqc/chainpqc-node/common"
)

type ContractsBlock struct {
	BaseBlock     blocks.BaseBlock `json:"base_block"`
	Chain         uint8            `json:"chain"`
	ContractsHash common.Hash      `json:"contracts_hash"`
	BlockHash     common.Hash      `json:"block_hash"`
}
