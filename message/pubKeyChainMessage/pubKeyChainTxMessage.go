package message

import (
	"github.com/chainpqc/chainpqc-node/message"
)

var chain uint8 = 1

type PubKeyChainTransactionsMessage message.AnyTransactionsMessage

func (c PubKeyChainTransactionsMessage) GetProperChain() uint8 {
	return chain
}
