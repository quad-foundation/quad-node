package message

import (
	"bytes"
	"fmt"
	"github.com/chainpqc/chainpqc-node/common"
	"github.com/chainpqc/chainpqc-node/transactionType"
	"log"
)

// tx - transaction, gt - get transaction, st - sync transaction
var validHeadTx = []string{"tx", "gt", "st"}
var validTopics = []string{
	"N0", "N1", "N2", "N3", "N4",
	"S0", "S1", "S2", "S3", "S4",
	"T0", "T1", "T2", "T3", "T4",
	"B0", "B1", "B2", "B3", "B4",
}

type AnyTransactionsMessage struct {
	BaseMessage       BaseMessage          `json:"base_message"`
	TransactionsBytes map[[2]byte][][]byte `json:"transactions_bytes"`
}

func (a AnyTransactionsMessage) GetTransactions() (map[[2]byte][]transactionType.AnyTransaction, error) {
	txn := map[[2]byte][]transactionType.AnyTransaction{}

	var t transactionType.AnyTransaction
	for topic, ret := range a.TransactionsBytes {
		chain := string(topic[1])
		for _, b := range ret {
			//b, rest, err := common.BytesWithLenToBytes(b)
			//if err != nil || len(rest) > 0 {
			//	return nil, err
			//}
			switch chain {
			case "0":
				tx := transactionType.MainChainTransaction{}
				at, rest, err := tx.GetFromBytes(b)
				if err != nil || len(rest) > 0 {
					return nil, err
				}
				t = at
			case "1":
				tx := transactionType.PubKeyChainTransaction{}
				at, rest, err := tx.GetFromBytes(b)
				if err != nil || len(rest) > 0 {
					return nil, err
				}
				t = at
			case "2":
				tx := transactionType.StakeChainTransaction{}
				at, rest, err := tx.GetFromBytes(b)
				if err != nil || len(rest) > 0 {
					return nil, err
				}
				t = at
			case "3":
				tx := transactionType.DexChainTransaction{}
				at, rest, err := tx.GetFromBytes(b)
				if err != nil || len(rest) > 0 {
					return nil, err
				}
				t = at
			case "4":
				tx := transactionType.ContractChainTransaction{}
				at, rest, err := tx.GetFromBytes(b)
				if err != nil || len(rest) > 0 {
					return nil, err
				}
				t = at
			default:
				return nil, fmt.Errorf("wrong chain number")
			}
			txn[topic] = append(txn[topic], t)
		}
	}
	return txn, nil
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

func (a AnyTransactionsMessage) GetBytes() []byte {
	tb := [2]byte{}

	b := a.BaseMessage.GetBytes()
	for _, t := range validTopics {
		copy(tb[:], t)
		if common.IsInKeysOfList(a.TransactionsBytes, tb) {
			for _, sb := range a.TransactionsBytes[tb] {
				b = append(b, t...)
				b = append(b, common.BytesToLenAndBytes(sb)...)
			}
		}
	}
	return b
}

func (a AnyTransactionsMessage) GetFromBytes(b []byte) (AnyMessage, error) {
	tb := [2]byte{}
	var err error
	var sb []byte
	a.BaseMessage.GetFromBytes(b[:5])
	a.TransactionsBytes = map[[2]byte][][]byte{}
	if len(b) > 7 {
		b = b[5:]
		for _, t := range validTopics {
			if len(b) == 0 {
				break
			}
			copy(tb[:], b[:2])
			if bytes.Equal(tb[:], []byte(t)) {
				a.TransactionsBytes[tb] = [][]byte{}
				b = b[2:]
				for len(b) > 0 {
					sb, b, err = common.BytesWithLenToBytes(b)
					if err != nil {
						log.Println("unmarshal AnyNonceMessage from bytes fails")
						return nil, err
					}
					a.TransactionsBytes[tb] = append(a.TransactionsBytes[tb], sb)
				}
			}
		}
	}
	return AnyMessage(a), nil
}
