package transactionType

import (
	"encoding/hex"
	"fmt"
	"github.com/chainpqc/chainpqc-node/account"
	"github.com/chainpqc/chainpqc-node/common"
	"github.com/chainpqc/chainpqc-node/transactionType"
	"strconv"
)

type DexChainTxData struct {
	Recipient common.Address `json:"recipient"`
	Amount    int64          `json:"amount"`
	OptData   []byte         `json:"opt_data,omitempty"`
}

type DexChainTransaction struct {
	TxData    DexChainTxData          `json:"tx_data"`
	TxParam   transactionType.TxParam `json:"tx_param"`
	Hash      common.Hash             `json:"hash"`
	Signature common.Signature        `json:"signature"`
	Height    int64                   `json:"height"`
	GasPrice  int64                   `json:"gas_price"`
	GasUsage  int64                   `json:"gas_usage"`
}

func (mt DexChainTransaction) GetChain() uint8 {
	return mt.TxParam.Chain
}

func (mt DexChainTransaction) GetData() transactionType.AnyDataTransaction {
	return mt.TxData
}

func (mt DexChainTransaction) GetParam() transactionType.TxParam {
	return mt.TxParam
}

func (mt DexChainTransaction) GasUsageEstimate() int64 {
	return 2100
}

func (mt DexChainTransaction) GetGasUsage() int64 {
	return 2100
}

func (td DexChainTxData) GetString() string {
	t := "Recipient: " + td.Recipient.GetHex() + "\n"
	t += "Amount PQC: " + fmt.Sprintln(account.Int64toFloat64(td.Amount)) + "\n"
	t += "Opt Data: " + hex.EncodeToString(td.OptData) + "\n"
	return t
}

func (tx DexChainTransaction) GetString() string {
	t := "Common parameters:\n" + tx.TxParam.GetString() + "\n"
	t += "Data:\n" + tx.TxData.GetString() + "\n"
	t += "Block Height: " + strconv.FormatInt(tx.Height, 10) + "\n"
	t += "Gas Price: " + strconv.FormatInt(tx.GasPrice, 10) + "\n"
	t += "Gas Usage: " + strconv.FormatInt(tx.GasUsage, 10) + "\n"
	t += "Hash: " + tx.Hash.GetHex() + "\n"
	t += "Signature: " + tx.Signature.GetHex() + "\n"
	return t
}

func (tx DexChainTransaction) GetSenderAddress() common.Address {
	return tx.TxParam.Sender
}

func (mt DexChainTransaction) GetSignature() common.Signature {
	return mt.Signature
}

func (mt DexChainTransaction) GetHeight() int64 {
	return mt.Height
}

func (mt DexChainTransaction) GetHash() common.Hash {
	return mt.Hash
}

func (md DexChainTxData) GetBytes() ([]byte, error) {
	b := md.Recipient.GetBytes()
	b = append(b, common.GetByteInt64(md.Amount)...)
	opt := common.BytesToLenAndBytes(md.OptData)
	b = append(b, opt...)
	return b, nil
}

func (tx DexChainTransaction) GetFromBytes(b []byte) (transactionType.AnyTransaction, []byte, error) {

	if len(b) < 56+common.SignatureLength {
		return nil, nil, fmt.Errorf("Not enough bytes for transaction unmarshal")
	}
	tp := transactionType.TxParam{}
	tp, b, err := tp.GetFromBytes(b)
	if err != nil {
		return nil, nil, err
	}
	td := DexChainTxData{}
	adata, b, err := td.GetFromBytes(b)
	if err != nil {
		return nil, nil, err
	}
	at := DexChainTransaction{
		TxData:    adata.(DexChainTxData),
		TxParam:   tp,
		Hash:      common.Hash{},
		Signature: common.Signature{},
		Height:    common.GetInt64FromByte(b[:8]),
		GasPrice:  common.GetInt64FromByte(b[8:16]),
		GasUsage:  common.GetInt64FromByte(b[16:24]),
	}
	hash, err := common.GetHashFromBytes(b[24:56])
	if err != nil {
		return nil, nil, err
	}
	at.Hash = hash

	signature, err := common.GetSignatureFromBytes(b[56:], tp.Sender)
	if err != nil {
		return nil, nil, err
	}
	at.Signature = signature
	return transactionType.AnyTransaction(&at), b[56+common.SignatureLength:], nil
}

func (DexChainTxData) GetFromBytes(data []byte) (transactionType.AnyDataTransaction, []byte, error) {
	md := DexChainTxData{}
	address, err := common.BytesToAddress(data[:common.AddressLength])
	if err != nil {
		return nil, []byte{}, err
	}
	md.Recipient = address
	amountBytes := data[common.AddressLength : common.AddressLength+8]
	md.Amount = common.GetInt64FromByte(amountBytes)
	opt, left, err := common.BytesWithLenToBytes(data[common.AddressLength+8:])
	if err != nil {
		return nil, []byte{}, err
	}
	md.OptData = opt
	return transactionType.AnyDataTransaction(md), left, nil
}

func (mt DexChainTransaction) GetPrice() int64 {
	return mt.GasPrice
}

func (tx DexChainTransaction) GetBytesWithoutSignature(withHash bool) []byte {

	b := tx.TxParam.GetBytes()
	bd, err := tx.TxData.GetBytes()
	if err != nil {
		return nil
	}
	b = append(b, bd...)
	b = append(b, common.GetByteInt64(tx.Height)...)
	b = append(b, common.GetByteInt64(tx.GasPrice)...)
	b = append(b, common.GetByteInt64(tx.GasUsage)...)
	if withHash {
		b = append(b, tx.GetHash().GetBytes()...)
	}
	return b
}

func (mt DexChainTransaction) CalcHash() (common.Hash, error) {
	b := mt.GetBytesWithoutSignature(false)
	hash, err := common.CalcHashFromBytes(b)
	if err != nil {
		return common.Hash{}, err
	}
	return hash, nil
}

func (mt *DexChainTransaction) SetHash(h common.Hash) {
	mt.Hash = h
}

func (md DexChainTxData) GetOptData() []byte {
	return md.OptData
}
func (md DexChainTxData) GetAmount() int64 {
	return md.Amount
}
func (md DexChainTxData) GetRecipient() common.Address {
	return md.Recipient
}

func (mt *DexChainTransaction) SetSignature(s common.Signature) {
	mt.Signature = s
}
