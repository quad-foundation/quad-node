package message

import (
	"github.com/chainpqc/chainpqc-node/common"
	"github.com/chainpqc/chainpqc-node/transactionType"
	"log"
)

// tx - transaction, gt - get transaction, st - sync transaction
var validHeadTx = []string{"tx", "gt", "st"}

type AnyTransactionsMessage struct {
	BaseMessage       BaseMessage         `json:"base_message"`
	TransactionsBytes map[string][][]byte `json:"transactions_bytes"`
}

func (a AnyTransactionsMessage) GetTransactions() []transactionType.AnyTransaction {
	txn := []transactionType.AnyTransaction{}

	ret := a.TransactionsBytes["tx"]

	var t *transactionType.AnyTransaction
	for _, b := range ret {
		err := common.Unmarshal(b, "T"+string(a.GetChain()), t)
		if err != nil {
			log.Println(err)
			continue
		}
		txn = append(txn, *t)
	}
	return txn
}

func (a AnyTransactionsMessage) GetBytes() []byte {
	b := a.BaseMessage.GetBytes()
	for _, t := range validHeadTx {
		for _, sb := range a.TransactionsBytes[t] {
			b = append(b, sb...)
		}
	}
	return b
}

func (b AnyTransactionsMessage) GetChain() uint8 {
	return b.BaseMessage.Chain
}

func (b AnyTransactionsMessage) GetHead() string {
	return b.BaseMessage.Head
}

func (b AnyTransactionsMessage) GetChainID() int16 {
	return b.BaseMessage.ChainID
}

func (b AnyTransactionsMessage) GetValidHead() []string {
	return validHeadTx
}
