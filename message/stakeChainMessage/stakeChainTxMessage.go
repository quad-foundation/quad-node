package message

import (
	"github.com/chainpqc/chainpqc-node/message"
)

var chain uint8 = 2

type StakeChainTransactionsMessage message.AnyTransactionsMessage

func (c StakeChainTransactionsMessage) GetProperChain() uint8 {
	return chain
}
