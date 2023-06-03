package block

import (
	"github.com/chainpqc/chainpqc-node/blocks"
	"github.com/chainpqc/chainpqc-node/common"
)

type TransactionsBlock struct {
	BaseBlock        blocks.BaseBlock `json:"base_block"`
	Chain            uint8            `json:"chain"`
	TransactionsHash common.Hash      `json:"transaction_hashes"`
	BlockHash        common.Hash      `json:"block_hash"`
}
