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

func (md DexChainTxData) GetBytes() []byte {
	b := md.Recipient.GetBytes()
	b = append(b, common.GetByteInt64(md.Amount)...)
	b = append(b, md.OptData...)
	return b
}

func (mt DexChainTransaction) GetPrice() int64 {
	return mt.GasPrice
}

func (tx DexChainTransaction) GetBytesWithoutSignature(withHash bool) []byte {

	b := tx.TxParam.GetBytes()
	b = append(b, tx.TxData.GetBytes()...)
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
	hash, err := common.GetHashFromBytes(b)
	if err != nil {
		return common.Hash{}, err
	}
	return hash, nil
}

func (mt *DexChainTransaction) SetHash(h common.Hash) {
	mt.Hash = h
}
