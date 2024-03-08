package transactionType

import (
	"fmt"
	"github.com/quad/quad-node/common"
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

var (
	PoolsTx [5]*TransactionPool
)

func init() {
	for c := 0; c < 5; c++ {
		PoolsTx[c] = NewTransactionPool(common.MaxTransactionInPool)
	}
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
