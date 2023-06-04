package message

import (
	"encoding/json"
	"github.com/chainpqc/chainpqc-node/common"
	"github.com/chainpqc/chainpqc-node/transactionType"
	"log"
)

// tx - transaction, gt - get transaction, st - sync transaction
var validHeadTx = [][]byte{[]byte("tx"), []byte("gt"), []byte("st")}

type AnyTransactionsMessage struct {
	BaseMessage       BaseMessage          `json:"base_message"`
	TransactionsBytes map[[2]byte][][]byte `json:"transactions_bytes"`
}

func (a AnyTransactionsMessage) GetTransactions() []transactionType.AnyTransaction {
	txn := []transactionType.AnyTransaction{}

	tb := [2]byte{}
	copy(tb[:], "tx")
	ret := a.TransactionsBytes[tb]

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
	tb := [2]byte{}

	b := a.BaseMessage.GetBytes()
	for _, t := range validHeadTx {
		copy(tb[:], t)
		if common.IsInKeysOfList(a.TransactionsBytes, tb) {
			for _, sb := range a.TransactionsBytes[tb] {
				b = append(b, sb...)
			}
		}
	}
	return b
}

func (a *AnyTransactionsMessage) GetFromBytes(b []byte) error {
	tb := [2]byte{}

	a.BaseMessage.GetFromBytes(b[:5])
	for _, t := range validHeadTx {
		copy(tb[:], t)
		a.TransactionsBytes[tb] = [][]byte{}
		for _, sb := range a.TransactionsBytes[tb] {
			a.TransactionsBytes[tb] = append(a.TransactionsBytes[tb], sb)
		}
	}
	return nil
}

func (b AnyTransactionsMessage) GetChain() uint8 {
	return b.BaseMessage.Chain
}

func (b AnyTransactionsMessage) GetHead() []byte {
	return b.BaseMessage.Head
}

func (b AnyTransactionsMessage) GetChainID() int16 {
	return b.BaseMessage.ChainID
}

func (b AnyTransactionsMessage) GetValidHead() [][]byte {
	return validHeadTx
}

func (m AnyTransactionsMessage) Marshal() ([]byte, error) {
	mb, err := json.Marshal(m)
	if err != nil {
		log.Println("error unmarshalling message (nonceMsg)", err)
		return nil, err
	}
	return mb, nil
}

func (m AnyTransactionsMessage) Unmarshal(b []byte) (AnyTransactionsMessage, error) {
	err := json.Unmarshal(b, &m)
	if err != nil {
		log.Println("error unmarshalling message (nonceMsg)", err)
		return AnyTransactionsMessage{}, err
	}
	return m, nil
}
