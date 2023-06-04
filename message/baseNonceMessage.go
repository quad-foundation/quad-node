package message

import (
	"bytes"
	"github.com/chainpqc/chainpqc-node/common"
	"github.com/chainpqc/chainpqc-node/transactionType"
	"log"
)

// nn - nounce, bl - block, rb - reject block
var validHeadNonce = [][]byte{[]byte("nn"), []byte("bl"), []byte("rb")}

type AnyNonceMessage struct {
	BaseMessage BaseMessage          `json:"base_message"`
	NonceBytes  map[[2]byte][][]byte `json:"nonce_bytes"`
}

func (a AnyNonceMessage) GetTransactions() []transactionType.AnyTransaction {
	txn := []transactionType.AnyTransaction{}

	nnb := [2]byte{}
	copy(nnb[:], "nn")
	ret := a.NonceBytes[nnb]

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

func (b AnyNonceMessage) GetHead() []byte {
	return b.BaseMessage.Head
}

func (b AnyNonceMessage) GetChainID() int16 {
	return b.BaseMessage.ChainID
}

func (b AnyNonceMessage) GetValidHead() [][]byte {
	return validHeadNonce
}

func (a AnyNonceMessage) GetBytes() []byte {
	tb := [2]byte{}

	b := a.BaseMessage.GetBytes()
	for _, t := range validHeadNonce {
		copy(tb[:], t)
		if common.IsInKeysOfList(a.NonceBytes, tb) {
			for _, sb := range a.NonceBytes[tb] {
				b = append(b, common.BytesToLenAndBytes(sb)...)
			}
		}
	}
	return b
}

func (a *AnyNonceMessage) GetFromBytes(b []byte) error {
	tb := [2]byte{}
	var err error
	var sb []byte
	a.BaseMessage.GetFromBytes(b[:5])
	a.NonceBytes = map[[2]byte][][]byte{}
	if len(b) > 8 {
		b = b[6:]
		for _, t := range validHeadNonce {
			copy(tb[:], b[:2])
			if bytes.Equal(tb[:], t) {
				a.NonceBytes[tb] = [][]byte{}
				b = b[2:]
				for len(b) > 0 {
					sb, b, err = common.BytesWithLenToBytes(b)
					if err != nil {
						log.Println("unmarshal AnyNonceMessage from bytes fails")
						return err
					}
					a.NonceBytes[tb] = append(a.NonceBytes[tb], sb)
				}
			}
		}
	}
	return nil
}
