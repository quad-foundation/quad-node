package block

import (
	"github.com/chainpqc/chainpqc-node/blocks"
	"github.com/chainpqc/chainpqc-node/common"
)

type PubKeysBlock struct {
	BaseBlock   blocks.BaseBlock `json:"base_block"`
	Chain       uint8            `json:"chain"`
	PubKeysHash common.Hash      `json:"pub_keys_hash"`
	BlockHash   common.Hash      `json:"block_hash"`
}
