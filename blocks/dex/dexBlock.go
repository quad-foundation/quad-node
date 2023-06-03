package block

import (
	"github.com/chainpqc/chainpqc-node/blocks"
	"github.com/chainpqc/chainpqc-node/common"
)

type DexBlock struct {
	BaseBlock blocks.BaseBlock `json:"base_block"`
	Chain     uint8            `json:"chain"`
	DexHash   common.Hash      `json:"dex_hash"`
	BlockHash common.Hash      `json:"block_hash"`
}
