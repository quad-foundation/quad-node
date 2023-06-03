package message

import (
	"github.com/chainpqc/chainpqc-node/message"
)

var chain uint8 = 0

type MainChainTransactionsMessage message.AnyTransactionsMessage

func (c MainChainTransactionsMessage) GetProperChain() uint8 {
	return chain
}
