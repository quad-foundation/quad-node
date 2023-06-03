package message

import (
	transactionType "github.com/chainpqc/chainpqc-node/transactionType/stakeChainTransaction"
)

var chain uint8 = 2

type StakeChainTransactionsMessage struct {
	Transaction transactionType.StakeChainTransaction `json:"transaction"`
}

func (c StakeChainTransactionsMessage) GetHeight() int64 {
	return c.Transaction.Height
}
