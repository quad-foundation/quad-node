package message

import (
	"bytes"
	"github.com/quad/quad-node/common"
	"github.com/quad/quad-node/transactionsDefinition"
	"log"
)

var validTopics = [][2]byte{{'N', 0}, {'S', 0}, {'T', 0}, {'B', 0}}

type TransactionsMessage struct {
	BaseMessage       BaseMessage          `json:"base_message"`
	TransactionsBytes map[[2]byte][][]byte `json:"transactions_bytes"`
}

func (a TransactionsMessage) GetTransactions() (map[[2]byte][]transactionsDefinition.Transaction, error) {
	txn := map[[2]byte][]transactionsDefinition.Transaction{}
	for _, topic := range validTopics {
		chain := a.GetChain()
		topic[1] = chain
		if common.IsInKeysOfList(a.TransactionsBytes, topic) {
			for _, tb := range a.TransactionsBytes[topic] {
				tx := transactionsDefinition.Transaction{}
				at, rest, err := tx.GetFromBytes(tb)
				if err != nil || len(rest) > 0 {
					return nil, err
				}
				txn[topic] = append(txn[topic], at)
			}
		}
	}

	return txn, nil
}

func (b TransactionsMessage) GetChain() uint8 {
	return b.BaseMessage.Chain
}

func (b TransactionsMessage) GetHead() []byte {
	return b.BaseMessage.Head
}

func (b TransactionsMessage) GetChainID() int16 {
	return b.BaseMessage.ChainID
}

func (a TransactionsMessage) GetBytes() []byte {

	b := a.BaseMessage.GetBytes()
	for _, topic := range validTopics {
		topic[1] = a.GetChain()
		if common.IsInKeysOfList(a.TransactionsBytes, topic) {
			for _, sb := range a.TransactionsBytes[topic] {
				b = append(b, topic[:]...)
				b = append(b, common.BytesToLenAndBytes(sb)...)
			}
		}
	}
	return b
}

func (a TransactionsMessage) GetFromBytes(b []byte) (AnyMessage, error) {

	var err error
	var sb []byte
	a.BaseMessage.GetFromBytes(b[:5])
	a.TransactionsBytes = map[[2]byte][][]byte{}
	if len(b) > 7 {
		b = b[5:]
		for _, topic := range validTopics {
			if len(b) == 0 {
				break
			}
			topic[1] = a.GetChain()
			if bytes.Equal(b[:2], topic[:]) {
				a.TransactionsBytes[topic] = [][]byte{}
				for len(b) > 0 {
					b = b[2:]
					sb, b, err = common.BytesWithLenToBytes(b)
					if err != nil {
						log.Println("unmarshal AnyNonceMessage from bytes fails")
						return nil, err
					}
					a.TransactionsBytes[topic] = append(a.TransactionsBytes[topic], sb)
				}
			}
		}
	}
	return AnyMessage(a), nil
}
