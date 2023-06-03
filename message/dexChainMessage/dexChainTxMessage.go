package message

import (
	"github.com/chainpqc/chainpqc-node/message"
)

var chain uint8 = 3

type DexChainTransactionsMessage message.AnyTransactionsMessage

func (c DexChainTransactionsMessage) GetProperChain() uint8 {
	return chain
}
