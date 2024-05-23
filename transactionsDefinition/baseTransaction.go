package transactionsDefinition

import (
	"fmt"
	"github.com/quad-foundation/quad-node/common"
	"strconv"
	"time"
)

type TxParam struct {
	ChainID     int16          `json:"chain_id"`
	Sender      common.Address `json:"sender"`
	SendingTime int64          `json:"sending_time"`
	Nonce       int16          `json:"nonce"`
}

func (tp TxParam) GetBytes() []byte {

	b := common.GetByteInt16(tp.ChainID)
	b = append(b, tp.Sender.GetBytes()...)
	b = append(b, common.GetByteInt64(tp.SendingTime)...)
	b = append(b, common.GetByteInt16(tp.Nonce)...)
	return b
}

func (tp TxParam) GetFromBytes(b []byte) (TxParam, []byte, error) {
	var err error
	if len(b) < 32 {
		return TxParam{}, []byte{}, fmt.Errorf("not enough bytes in TxParam unmarshaling")
	}
	tp.ChainID = common.GetInt16FromByte(b[:2])
	tp.Sender, err = common.BytesToAddress(b[2:22])
	if err != nil {
		return TxParam{}, []byte{}, err
	}
	tp.SendingTime = common.GetInt64FromByte(b[22:30])
	tp.Nonce = common.GetInt16FromByte(b[30:32])
	return tp, b[32:], nil
}

func (tp TxParam) GetString() string {

	t := "Time: " + time.Unix(tp.SendingTime, 0).String() + "\n"
	t += "ChainID: " + strconv.Itoa(int(tp.ChainID)) + "\n"
	t += "Nonce: " + strconv.Itoa(int(tp.Nonce)) + "\n"
	t += "Sender Address: " + tp.Sender.GetHex() + "\n"
	return t
}
