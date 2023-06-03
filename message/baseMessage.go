package message

import (
	"github.com/chainpqc/chainpqc-node/common"
	"github.com/chainpqc/chainpqc-node/transactionType"
)

type BaseMessage struct {
	Head    string `json:"head"`
	ChainID int16  `json:"chainID"`
	Chain   uint8  `json:"chain"`
}

type AnyMessage interface {
	GetHead() string
	GetValidHead() []string
	GetChainID() int16
	GetTransactions() []transactionType.AnyTransaction
	GetChain() uint8
	GetBytes() []byte
	Marshal() ([]byte, error)
	Unmarshal(b []byte) (AnyMessage, error)
}

func (m BaseMessage) GetBytes() []byte {
	b := []byte(m.Head)
	b = append(b, common.GetByteInt16(m.ChainID)...)
	b = append(b, m.Chain)
	return b
}

func CheckMessage(a AnyMessage) bool {
	isValidChainID := a.GetChainID() == common.GetChainID()

	isValidChain := false
	for _, chain := range common.ValidChains {
		if a.GetChain() == chain {
			isValidChain = true
			break
		}
	}
	isValidHead := false
	for _, key := range a.GetValidHead() {
		if a.GetHead() == key {
			isValidHead = true
			break
		}
	}
	return isValidHead && isValidChain && isValidChainID
}
