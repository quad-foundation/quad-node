package transactionType

import (
	"fmt"
	"github.com/chainpqc/chainpqc-node/common"
	memDatabase "github.com/chainpqc/chainpqc-node/database"
	"github.com/chainpqc/chainpqc-node/wallet"
	"log"
	"strconv"
	"time"
)

type TxParam struct {
	ChainID     int16          `json:"chain_id"`
	Sender      common.Address `json:"sender"`
	SendingTime int64          `json:"sending_time"`
	Nonce       int16          `json:"nonce"`
	Chain       uint8          `json:"chain"`
}

type AnyDataTransaction interface {
	GetBytes() ([]byte, error)
	GetFromBytes(data []byte) (AnyDataTransaction, []byte, error)
	GetAmount() int64
	GetOptData() []byte
	GetRecipient() common.Address
}

type AnyTransaction interface {
	GetHash() common.Hash
	GetParam() TxParam
	GetData() AnyDataTransaction
	GetSenderAddress() common.Address
	GetFromBytes([]byte) (AnyTransaction, []byte, error)
	//Store() error
	//StoreToPool(dbprefix string) error
	//DeleteFromPool(dbprefix string) error
	//LoadByHash(hash common.Hash, dbPrefix string) (AnyTransaction, error)
	//CheckTransaction(int64) (bool, int64)
	GetHeight() int64
	GetGasUsage() int64
	GetGasPrice() int64
	//FundsUsedForTx() (recipientFunds int64, senderCost int64)
	GetChain() uint8
	GetString() string
	GetSignature() common.Signature
	GetBytesWithoutSignature(bool) []byte
	CalcHash() (common.Hash, error)
	SetHash(h common.Hash)
	SetSignature(s common.Signature)
}

var (
	poolMainTx     PoolMainTx
	poolPubKeyTx   PoolPubKeyTx
	poolStakeTx    PoolStakeTx
	poolDexTx      PoolDexTx
	poolContractTx PoolContractTx
)

func init() {
	poolMainTx = make(PoolMainTx, 1)
	poolPubKeyTx = make(PoolPubKeyTx, 1)
	poolStakeTx = make(PoolStakeTx, 1)
	poolDexTx = make(PoolDexTx, 1)
	poolContractTx = make(PoolContractTx, 1)
}

func (tp TxParam) GetBytes() []byte {

	b := []byte{tp.Chain}
	b = append(b, common.GetByteInt16(tp.ChainID)...)
	b = append(b, tp.Sender.GetBytes()...)
	b = append(b, common.GetByteInt64(tp.SendingTime)...)
	b = append(b, common.GetByteInt16(tp.Nonce)...)
	return b
}

func (tp TxParam) GetFromBytes(b []byte) (TxParam, []byte, error) {
	var err error
	if len(b) < 33 {
		return TxParam{}, []byte{}, fmt.Errorf("not enough bytes in TxParam unmarshaling")
	}
	tp.Chain = b[0]
	tp.ChainID = common.GetInt16FromByte(b[1:3])
	tp.Sender, err = common.BytesToAddress(b[3:23])
	if err != nil {
		return TxParam{}, []byte{}, err
	}
	tp.SendingTime = common.GetInt64FromByte(b[23:31])
	tp.Nonce = common.GetInt16FromByte(b[31:33])
	return tp, b[33:], nil
}

func (tp TxParam) GetString() string {

	t := "Time: " + time.Unix(tp.SendingTime, 0).String() + "\n"
	t += "ChainID: " + strconv.Itoa(int(tp.ChainID)) + "\n"
	t += "Nonce: " + strconv.Itoa(int(tp.Nonce)) + "\n"
	t += "Sender Address: " + tp.Sender.GetHex() + "\n"
	t += "Chain: " + string(tp.Chain) + "\n"
	return t
}

func GetBytes(tx AnyTransaction) []byte {
	b := tx.GetBytesWithoutSignature(true)
	b = append(b, tx.GetSignature().GetBytes()...)
	return b
}

func VerifyTransaction(tx AnyTransaction) bool {
	b := tx.GetHash().GetBytes()
	a := tx.GetSenderAddress()
	pk, err := memDatabase.Load(append(common.PubKeyDBPrefix[:], a.GetBytes()...))
	if err != nil {
		return false
	}
	signature := tx.GetSignature()
	return wallet.Verify(b, signature.GetBytes(), pk)
}

func SignTransaction(tx AnyTransaction) (common.Signature, error) {
	b := tx.GetHash()

	w := wallet.EmptyWallet()
	w = w.GetWallet()
	return w.Sign(b.GetBytes())
}

func SignTransactionAllToBytes(tx AnyTransaction) ([]byte, error) {
	signature, err := SignTransaction(tx)
	if err != nil {
		return nil, err
	}
	b := tx.GetBytesWithoutSignature(true)
	b = append(b, signature.GetBytes()...)
	return b, nil
}

func GetBytesWithoutSignature(tx AnyTransaction, withHash bool) []byte {
	b := tx.GetParam().GetBytes()
	bd, err := tx.GetData().GetBytes()
	if err != nil {
		return nil
	}
	b = append(b, bd...)
	b = append(b, tx.GetHash().GetBytes()...)
	b = append(b, common.GetByteInt64(tx.GetHeight())...)
	b = append(b, common.GetByteInt64(tx.GetGasPrice())...)
	b = append(b, common.GetByteInt64(tx.GetGasUsage())...)
	if withHash {
		b = append(b, tx.GetHash().GetBytes()...)
	}
	return b
}

func AddTransactionsToPool(txs []AnyTransaction, chain uint8) {
	switch chain {
	case 0:
		poolMainTx.AddTransactions(txs)
	case 1:
		poolPubKeyTx.AddTransactions(txs)
	case 2:
		poolStakeTx.AddTransactions(txs)
	case 3:
		poolDexTx.AddTransactions(txs)
	case 4:
		poolContractTx.AddTransactions(txs)
	default:
		log.Println("tx was not added to pool")
	}
}

func GetTransactionsFromPool(number int, chain uint8) []AnyTransaction {
	switch chain {
	case 0:
		return poolMainTx.GetTransactionsFromPool(number)
	case 1:
		return poolPubKeyTx.GetTransactionsFromPool(number)
	case 2:
		return poolStakeTx.GetTransactionsFromPool(number)
	case 3:
		return poolDexTx.GetTransactionsFromPool(number)
	case 4:
		return poolContractTx.GetTransactionsFromPool(number)
	default:
		log.Println("tx was not added to pool")
	}
	return nil
}
