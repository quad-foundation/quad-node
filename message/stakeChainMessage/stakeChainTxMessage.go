package message

import (
	"github.com/chainpqc/chainpqc-node/transactionType"
)

var chain uint8 = 2

type StakeChainTransactionsMessage struct {
	Transaction []transactionType.StakeChainTransaction `json:"transaction"`
}

func (c StakeChainTransactionsMessage) GetHeight() int64 {
	return c.Transaction[0].Height
}
