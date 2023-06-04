package transactionType

import (
	"github.com/chainpqc/chainpqc-node/common"
	"github.com/chainpqc/chainpqc-node/transactionType"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPubKeyChainTxData_GetBytes(t *testing.T) {
	// Create a sample PubKeyChainTxData object
	md := PubKeyChainTxData{
		Recipient: common.Address{},
		Amount:    1000,
		OptData:   []byte("optional data"),
	}
	// Call the GetBytes method
	b, err := md.GetBytes()
	if err != nil {
		return
	}
	// Check if the returned byte slice is correct
	expected := append(md.Recipient.GetBytes(), common.GetByteInt64(md.Amount)...)
	expected = append(expected, common.BytesToLenAndBytes(md.OptData)...)
	assert.Equal(t, expected, b)
}
func TestPubKeyChainTxData_GetFromBytes(t *testing.T) {
	// Create a sample byte slice representing PubKeyChainTxData
	md := PubKeyChainTxData{
		Recipient: common.Address{},
		Amount:    1000,
		OptData:   []byte("optional data"),
	}
	bd, err := md.GetBytes()
	if err != nil {
		return
	}
	// Call the GetFromBytes method
	adata, left, err := PubKeyChainTxData{}.GetFromBytes(bd)
	data, err := adata.GetBytes()
	if err != nil {
		return
	}
	// Check if the returned object and error are correct
	assert.Equal(t, len(left), 0)
	assert.NoError(t, err)
	assert.NotNil(t, adata)
	assert.Equal(t, bd, data)
}
func TestPubKeyChainTransaction_GetFromBytes(t *testing.T) {
	// Create a sample byte slice representing PubKeyChainTransaction
	md := PubKeyChainTxData{
		Recipient: common.Address{},
		Amount:    1000,
		OptData:   []byte("optional data"),
	}
	pk := PubKeyChainTransaction{
		TxData:    md,
		TxParam:   transactionType.TxParam{},
		Hash:      common.Hash{},
		Signature: common.Signature{},
		Height:    4,
		GasPrice:  2,
		GasUsage:  1,
	}
	zero32 := make([]byte, 32)
	zeroSig := make([]byte, common.SignatureLength)
	zero20 := make([]byte, 20)
	address, _ := common.BytesToAddress(zero20)
	pk.Signature, _ = common.GetSignatureFromBytes(zeroSig, address)
	pk.Hash, _ = common.GetHashFromBytes(zero32)
	data := transactionType.GetBytes(transactionType.AnyTransaction(&pk))
	// Call the GetFromBytes method
	at, left, err := PubKeyChainTransaction{}.GetFromBytes(data)
	// Check if the returned object and error are correct
	assert.Equal(t, len(left), 0)
	assert.NoError(t, err)
	assert.NotNil(t, at)
	assert.Equal(t, data, transactionType.GetBytes(at))
}
