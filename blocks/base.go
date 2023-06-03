package blocks

import (
	"github.com/chainpqc/chainpqc-node/common"
)

type BaseHeader struct {
	PreviousHash     common.Hash      `json:"previous_hash"`
	Difficulty       int32            `json:"difficulty"`
	Height           int64            `json:"height"`
	DelegatedAccount common.Address   `json:"delegated_account"`
	OperatorAccount  common.Address   `json:"operator_account"`
	SignatureMsg     common.Signature `json:"signature_msg"`
	NonceMessage     []byte           `json:"nonce_message"`
}

type BaseBlock struct {
	BaseHeader       BaseHeader  `json:"header"`
	BlockHeaderHash  common.Hash `json:"block_header_hash"`
	BlockTimeStamp   int64       `json:"block_time_stamp"`
	RewardPercentage int16       `json:"reward_percentage"`
}
