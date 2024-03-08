package transactionType

import (
	"encoding/hex"
	"fmt"
	"github.com/quad/quad-node/account"
	"github.com/quad/quad-node/common"
	memDatabase "github.com/quad/quad-node/database"
	"github.com/quad/quad-node/wallet"
	"log"
	"strconv"
)

type TxData struct {
	Recipient common.Address `json:"recipient"`
	Amount    int64          `json:"amount"`
	OptData   []byte         `json:"opt_data,omitempty"`
}

type Transaction struct {
	TxData    TxData           `json:"tx_data"`
	TxParam   TxParam          `json:"tx_param"`
	Hash      common.Hash      `json:"hash"`
	Signature common.Signature `json:"signature"`
	Height    int64            `json:"height"`
	GasPrice  int64            `json:"gas_price"`
	GasUsage  int64            `json:"gas_usage"`
}

func (mt *Transaction) GetChain() uint8 {
	return mt.TxParam.Chain
}

func (mt *Transaction) GetData() TxData {
	return mt.TxData
}

func (mt *Transaction) GetParam() TxParam {
	return mt.TxParam
}

func (mt *Transaction) GasUsageEstimate() int64 {
	return 2100
}

func (mt *Transaction) GetGasUsage() int64 {
	return 2100
}

func (mt *Transaction) GetSignature() common.Signature {
	return mt.Signature
}

func (mt *Transaction) GetHeight() int64 {
	return mt.Height
}

func (mt *Transaction) GetHash() common.Hash {
	return mt.Hash
}

func (td TxData) GetString() string {
	t := "Recipient: " + td.Recipient.GetHex() + "\n"
	t += "Amount PQC: " + fmt.Sprintln(account.Int64toFloat64(td.Amount)) + "\n"
	t += "Opt Data: " + hex.EncodeToString(td.OptData) + "\n"
	return t
}

func (tx *Transaction) GetString() string {
	t := "Common parameters:\n" + tx.TxParam.GetString() + "\n"
	t += "Data:\n" + tx.TxData.GetString() + "\n"
	t += "Block Height: " + strconv.FormatInt(tx.Height, 10) + "\n"
	t += "Gas Price: " + strconv.FormatInt(tx.GasPrice, 10) + "\n"
	t += "Gas Usage: " + strconv.FormatInt(tx.GasUsage, 10) + "\n"
	t += "Hash: " + tx.Hash.GetHex() + "\n"
	t += "Signature: " + tx.Signature.GetHex() + "\n"
	return t
}

func (tx *Transaction) GetSenderAddress() common.Address {
	return tx.TxParam.Sender
}

func (md TxData) GetBytes() ([]byte, error) {
	b := md.Recipient.GetBytes()
	b = append(b, common.GetByteInt64(md.Amount)...)
	opt := common.BytesToLenAndBytes(md.OptData)
	b = append(b, opt...)
	return b, nil
}

func (tx *Transaction) GetFromBytes(b []byte) (Transaction, []byte, error) {

	if len(b) < 56+common.SignatureLength {
		return Transaction{}, nil, fmt.Errorf("Not enough bytes for transaction unmarshal")
	}
	tp := TxParam{}
	tp, b, err := tp.GetFromBytes(b)
	if err != nil {
		return Transaction{}, nil, err
	}
	td := TxData{}
	adata, b, err := td.GetFromBytes(b)
	if err != nil {
		return Transaction{}, nil, err
	}
	at := Transaction{
		TxData:    adata,
		TxParam:   tp,
		Hash:      common.Hash{},
		Signature: common.Signature{},
		Height:    common.GetInt64FromByte(b[:8]),
		GasPrice:  common.GetInt64FromByte(b[8:16]),
		GasUsage:  common.GetInt64FromByte(b[16:24]),
	}
	at.Hash = common.GetHashFromBytes(b[24:56])
	signature, err := common.GetSignatureFromBytes(b[56:], tp.Sender)
	if err != nil {
		return Transaction{}, nil, err
	}
	at.Signature = signature
	return at, b[56+signature.GetLength():], nil
}

func (TxData) GetFromBytes(data []byte) (TxData, []byte, error) {
	md := TxData{}
	address, err := common.BytesToAddress(data[:common.AddressLength])
	if err != nil {
		return TxData{}, []byte{}, err
	}
	md.Recipient = address
	amountBytes := data[common.AddressLength : common.AddressLength+8]
	md.Amount = common.GetInt64FromByte(amountBytes)
	opt, left, err := common.BytesWithLenToBytes(data[common.AddressLength+8:])
	if err != nil {
		return TxData{}, []byte{}, err
	}
	md.OptData = opt
	return md, left, nil
}

func (mt *Transaction) GetGasPrice() int64 {
	return mt.GasPrice
}

func (tx *Transaction) GetBytesWithoutSignature(withHash bool) []byte {

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

func (mt *Transaction) CalcHashAndSet() error {
	b := mt.GetBytesWithoutSignature(false)
	hash, err := common.CalcHashFromBytes(b)
	if err != nil {
		return err
	}
	mt.Hash = hash
	return nil
}

func (md TxData) GetOptData() []byte {
	return md.OptData
}
func (md TxData) GetAmount() int64 {
	return md.Amount
}
func (md TxData) GetRecipient() common.Address {
	return md.Recipient
}

func (mt *Transaction) GetBytes() []byte {
	if mt != nil {
		b := mt.GetBytesWithoutSignature(true)
		b = append(b, mt.GetSignature().GetBytes()...)
		return b
	}
	return nil
}

func (mt *Transaction) StoreToDBPoolTx(prefix []byte) error {
	prefix = append(prefix, mt.GetHash().GetBytes()...)
	bt := mt.GetBytes()
	err := (*memDatabase.MainDB).Put(prefix, bt)
	if err != nil {
		return err
	}
	return nil
}
func (mt *Transaction) RestoreFromDBPoolTx(prefix []byte, hashTransaction []byte) (Transaction, error) {
	prefix = append(prefix, hashTransaction...)
	bt, err := (*memDatabase.MainDB).Get(prefix)
	if err != nil {
		return Transaction{}, err
	}
	at, restb, err := mt.GetFromBytes(bt)
	if err != nil || len(restb) > 0 {
		return Transaction{}, err
	}
	return at, nil
}

func (tx *Transaction) Verify() bool {
	b := tx.GetHash().GetBytes()
	a := tx.GetSenderAddress()
	pk, err := (*memDatabase.MainDB).Get(append(common.PubKeyDBPrefix[:], a.GetBytes()...))
	if err != nil {
		return false
	}
	signature := tx.GetSignature()
	return wallet.Verify(b, signature.GetBytes(), pk)
}

func (tx *Transaction) Sign() error {
	b := tx.GetHash()
	w := wallet.GetActiveWallet()
	sign, err := w.Sign(b.GetBytes())
	if err != nil {
		return err
	}
	tx.Signature = *sign
	return nil
}

func EmptyTransaction() Transaction {
	tx := Transaction{
		TxData: TxData{
			Recipient: common.EmptyAddress(),
			Amount:    0,
			OptData:   []byte{},
		},
		TxParam: TxParam{
			ChainID:     0,
			Sender:      common.EmptyAddress(),
			SendingTime: 0,
			Nonce:       0,
			Chain:       0,
		},
		Hash:      common.EmptyHash(),
		Signature: common.Signature{},
		Height:    0,
		GasPrice:  0,
		GasUsage:  0,
	}
	err := tx.CalcHashAndSet()
	if err != nil {
		log.Println("empty transaction calc hash fails")
	}
	tx.Signature = common.EmptySignature()
	return tx
}
