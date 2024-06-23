package transactionsDefinition

import (
	"bytes"
	"fmt"
	"github.com/quad-foundation/quad-node/account"
	"github.com/quad-foundation/quad-node/common"
	memDatabase "github.com/quad-foundation/quad-node/database"
	"github.com/quad-foundation/quad-node/wallet"
	"log"
	"strconv"
)

type Transaction struct {
	TxData          TxData           `json:"tx_data"`
	TxParam         TxParam          `json:"tx_param"`
	Hash            common.Hash      `json:"hash"`
	Signature       common.Signature `json:"signature"`
	Height          int64            `json:"height"`
	GasPrice        int64            `json:"gas_price"`
	GasUsage        int64            `json:"gas_usage"`
	OutputLogs      []byte           `json:"outputLogs,omitempty"`
	ContractAddress common.Address   `json:"contractAddress,omitempty"`
}

func (mt *Transaction) GetData() TxData {
	return mt.TxData
}

func (mt *Transaction) GetParam() TxParam {
	return mt.TxParam
}

func (mt *Transaction) GasUsageEstimate() int64 {
	gas := len(mt.TxData.OptData) * 100
	gas += len(mt.TxData.Pubkey.GetBytes()) * 10
	return int64(gas) + 21000
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

func (tx *Transaction) GetString() string {
	t := "Common parameters:\n" + tx.TxParam.GetString() + "\n"
	t += "Data:\n" + tx.TxData.GetString() + "\n"
	t += "Block Height: " + strconv.FormatInt(tx.Height, 10) + "\n"
	t += "Gas Price: " + strconv.FormatInt(tx.GasPrice, 10) + "\n"
	t += "Gas Usage: " + strconv.FormatInt(tx.GasUsage, 10) + "\n"
	t += "Hash: " + tx.Hash.GetHex() + "\n"
	t += "Signature: " + tx.Signature.GetHex() + "\n"
	t += "Contract Address: " + tx.ContractAddress.GetHex() + "\n"
	t += "Contract Logs:\n" + string(tx.OutputLogs) + "\n"
	return t
}

func (tx *Transaction) GetSenderAddress() common.Address {
	return tx.TxParam.Sender
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
	vb, leftb, err := common.BytesWithLenToBytes(b[56:])
	if err != nil {
		return Transaction{}, nil, err
	}
	signature, err := common.GetSignatureFromBytes(vb, tp.Sender)
	if err != nil {
		return Transaction{}, nil, err
	}
	at.Signature = signature
	err = at.ContractAddress.Init(leftb[:20])
	if err != nil {
		return Transaction{}, nil, err
	}
	toBytes, leftb2, err := common.BytesWithLenToBytes(leftb[20:])
	if err != nil {
		return Transaction{}, nil, err
	}
	at.OutputLogs = toBytes[:]
	return at, leftb2, nil
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

func (mt *Transaction) GetBytes() []byte {
	if mt != nil {
		b := mt.GetBytesWithoutSignature(true)
		sb := common.BytesToLenAndBytes(mt.GetSignature().GetBytes())
		b = append(b, sb...)
		b = append(b, mt.ContractAddress.GetBytes()...)
		olb := common.BytesToLenAndBytes(mt.OutputLogs)
		b = append(b, olb...)

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

func (mt *Transaction) RemoveFromDBPoolTx(prefix []byte) error {
	prefix = append(prefix, mt.GetHash().GetBytes()...)
	err := (*memDatabase.MainDB).Delete(prefix)
	if err != nil {
		return err
	}
	return nil
}

func RemoveTransactionFromDBbyHash(prefix []byte, hash []byte) error {
	prefix = append(prefix, hash...)
	err := (*memDatabase.MainDB).Delete(prefix)
	if err != nil {
		return err
	}
	return nil
}

func LoadFromDBPoolTx(prefix []byte, hashTransaction []byte) (Transaction, error) {
	prefix = append(prefix, hashTransaction...)
	bt, err := (*memDatabase.MainDB).Get(prefix)
	if err != nil {
		return Transaction{}, err
	}
	mt := &Transaction{}
	at, restb, err := mt.GetFromBytes(bt)
	if err != nil || len(restb) > 0 {
		return Transaction{}, err
	}
	return at, nil
}

func CheckFromDBPoolTx(prefix []byte, hashTransaction []byte) bool {
	prefix = append(prefix, hashTransaction...)
	isKey, err := (*memDatabase.MainDB).IsKey(prefix)
	if err != nil {
		return false
	}
	return isKey
}

// Verify - checking if hash is correct and signature
func (tx *Transaction) Verify() bool {
	recipientAddress := tx.TxData.Recipient
	n, err := account.IntDelegatedAccountFromAddress(recipientAddress)
	if tx.GetData().Amount < 0 && err != nil && n < 512 {
		log.Println("transaction amount has to be larger or equal 0")
		return false
	}
	b := tx.GetHash().GetBytes()
	err = tx.CalcHashAndSet()
	if err != nil {
		return false
	}
	if bytes.Compare(b, tx.GetHash().GetBytes()) != 0 {
		log.Println("check transaction hash fails")
		return false
	}
	a := tx.GetSenderAddress()
	pk := tx.TxData.GetPubKey().GetBytes()
	if len(pk) == 0 {
		pk, err = (*memDatabase.MainDB).Get(append(common.PubKeyDBPrefix[:], a.GetBytes()...))
		if err != nil {
			return false
		}
	}
	signature := tx.GetSignature()
	return wallet.Verify(b, signature.GetBytes(), pk)
}

func (tx *Transaction) Sign(w *wallet.Wallet) error {
	b := tx.GetHash()
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
