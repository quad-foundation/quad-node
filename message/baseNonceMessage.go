package message

import (
	"encoding/json"
	"github.com/chainpqc/chainpqc-node/common"
	"github.com/chainpqc/chainpqc-node/transactionType"
	"log"
)

// nn - nounce, bl - block, rb - reject block
var validHeadNonce = []string{"nn", "bl", "rb"}

type AnyNonceMessage struct {
	BaseMessage BaseMessage         `json:"base_message"`
	NonceBytes  map[string][][]byte `json:"nonce_bytes"`
}

func (a AnyNonceMessage) GetTransactions() []transactionType.AnyTransaction {
	txn := []transactionType.AnyTransaction{}

	ret := a.NonceBytes["nn"]

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

func (b AnyNonceMessage) GetChain() uint8 {
	return b.BaseMessage.Chain
}

func (b AnyNonceMessage) GetHead() string {
	return b.BaseMessage.Head
}

func (b AnyNonceMessage) GetChainID() int16 {
	return b.BaseMessage.ChainID
}

func (b AnyNonceMessage) GetValidHead() []string {
	return validHeadNonce
}

func (a AnyNonceMessage) GetBytes() []byte {
	b := a.BaseMessage.GetBytes()
	for _, t := range validHeadNonce {
		if common.IsInKeysOfList(a.NonceBytes, t) {
			for _, sb := range a.NonceBytes[t] {
				b = append(b, sb...)
			}
		}
	}
	return b
}

func (m *AnyNonceMessage) UnMarshal(b []byte) error {
	err := json.Unmarshal(b, m)
	if err != nil {
		log.Println("error unmarshalling message (nonceMsg)", err)
		return err
	}
	return nil
}
