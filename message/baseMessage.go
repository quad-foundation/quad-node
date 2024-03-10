package message

import (
	"bytes"
	"github.com/quad/quad-node/common"
	"log"
)

// tx - transaction, gt - get transaction, st - sync transaction, "nn" - nonce, "bl" - block, "rb" - reject block, "hi" - GetHeight, "gh" - GetHeaders, "sh" - SendHeaders
var validHead = []string{"nn", "bl", "rb", "tx", "gt", "st", "hi", "gh", "sh"}

type BaseMessage struct {
	Head    []byte `json:"head"`
	ChainID int16  `json:"chainID"`
	Chain   uint8  `json:"chain"`
}

type AnyMessage interface {
	GetHead() []byte
	GetChainID() int16
	GetTransactionsBytes() map[[2]byte][][]byte
	GetChain() uint8
	GetBytes() []byte
	GetFromBytes([]byte) (AnyMessage, error)
}

func (m BaseMessage) GetBytes() []byte {
	b := m.Head[:]
	b = append(b, common.GetByteInt16(m.ChainID)...)
	b = append(b, m.Chain)
	return b
}

func (m *BaseMessage) GetFromBytes(b []byte) {
	if len(b) != 5 {
		log.Println("bytes length should be 5")
		return
	}
	m.Head = b[:2]
	if !common.ContainsKey(validHead, string(m.Head)) {
		log.Println("Head not in valid heads keys")
		return
	}
	m.ChainID = common.GetInt16FromByte(b[2:4])
	if m.ChainID != common.GetChainID() {
		log.Println("Wrong Chain ID")
		return
	}
	m.Chain = b[4]
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
	for _, key := range validHead {
		if bytes.Compare(a.GetHead(), []byte(key)) == 0 {
			isValidHead = true
			break
		}
	}
	return isValidHead && isValidChain && isValidChainID
}
