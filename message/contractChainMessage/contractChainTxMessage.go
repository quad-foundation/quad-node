package message

import (
	"github.com/chainpqc/chainpqc-node/message"
)

var chain uint8 = 4

type ContractsChainTransactionsMessage message.AnyTransactionsMessage

func (c ContractsChainTransactionsMessage) GetProperChain() uint8 {
	return chain
}
