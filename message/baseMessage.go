package message

import (
	"bytes"
	"github.com/chainpqc/chainpqc-node/common"
	"github.com/chainpqc/chainpqc-node/transactionType"
)

type BaseMessage struct {
	Head    []byte `json:"head"`
	ChainID int16  `json:"chainID"`
	Chain   uint8  `json:"chain"`
}

type AnyMessage interface {
	GetHead() []byte
	GetValidHead() [][]byte
	GetChainID() int16
	GetTransactions() []transactionType.AnyTransaction
	GetChain() uint8
	GetBytes() []byte
	GetFromBytes([]byte) error
	//Marshal() ([]byte, error)
	//Unmarshal(b []byte) (AnyMessage, error)
}

func (m BaseMessage) GetBytes() []byte {
	b := m.Head[:]
	b = append(b, common.GetByteInt16(m.ChainID)...)
	b = append(b, m.Chain)
	return b
}

func (m *BaseMessage) GetFromBytes(b []byte) {
	if len(b) < 6 {
		return
	}
	m.Head = b[:2]
	m.ChainID = common.GetInt16FromByte(b[2:4])
	m.Chain = b[5]
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
		if bytes.Compare(a.GetHead(), key) == 0 {
			isValidHead = true
			break
		}
	}
	return isValidHead && isValidChain && isValidChainID
}
