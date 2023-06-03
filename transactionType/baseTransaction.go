package transactionType

import (
	"github.com/chainpqc/chainpqc-node/common"
	"github.com/chainpqc/chainpqc-node/wallet"
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

type AnyTransaction interface {
	GetHash() common.Hash
	GetParam() TxParam
	GetSenderAddress() common.Address
	//Store() error
	//StoreToPool(dbprefix string) error
	//DeleteFromPool(dbprefix string) error
	//LoadByHash(hash common.Hash, dbPrefix string) (AnyTransaction, error)
	//CheckTransaction(int64) (bool, int64)
	GasUsageEstimate() int64
	GetHeight() int64
	GetGasUsage() int64
	//FundsUsedForTx() (recipientFunds int64, senderCost int64)
	GetChain() uint8
	GetString() string
	GetSignature() common.Signature
	GetBytesWithoutSignature() []byte
}

func (tp TxParam) GetBytes() []byte {
	b := []byte{tp.Chain}
	b = append(b, common.GetByteInt16(tp.ChainID)...)
	b = append(b, tp.Sender.GetBytes()...)
	b = append(b, common.GetByteInt64(tp.SendingTime)...)
	b = append(b, common.GetByteInt16(tp.Nonce)...)
	return b
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
	b := tx.GetSignature().GetBytes()
	b = append(b, tx.GetBytesWithoutSignature()...)
	return b
}

func VerifyTransaction(tx AnyTransaction) bool {
	b, err := common.CalcHashToByte(tx.GetBytesWithoutSignature())
	if err != nil {
		return false
	}
	a := tx.GetSenderAddress()
	pk, err := wallet.LoadPubKey(a)
	if err != nil {
		return false
	}

	return wallet.Verify(b, tx.GetSignature(), pk)
}

func SignTransaction(tx AnyTransaction) (common.Signature, error) {
	b, err := common.CalcHashToByte(tx.GetBytesWithoutSignature())
	if err != nil {
		return common.Signature{}, err
	}
	w := wallet.EmptyWallet()
	w = w.GetWallet()
	return w.Sign(b)
}

//
//func MarshalTx(tx AnyTransaction) []byte {
//
//	hash := tx.Hash()
//	sig := tx.GetSignature()
//	res := hash.GetByte()
//	res = append(res, common.GetByteInt64(tx.GetHeightFinal())...)
//	ln := common.GetByteInt16(int16(sig.GetLen()))
//	res = append(res, ln...)
//	res = append(res, sig.GetByte()...)
//	res = append(res, tx.GetByte()...)
//	return res
//}
//
//func UnmarshalTxParam(b []byte) (TxParam, error) {
//	var txParam TxParam
//
//	if len(b) != 13+common.GVMAddressLength {
//		return txParam, fmt.Errorf("wrong number of bytes in unmarshal txParam")
//	}
//	txParam.ChainID = common.GetInt16FromByte(b[:2])
//	b = b[2:]
//
//	sender := common.Address{}
//	err := sender.Init(b[:sender.GetLen()])
//	if err != nil {
//		return txParam, err
//	}
//	txParam.Sender = sender
//	b = b[sender.GetLen():]
//
//	txParam.Time = common.GetInt64FromByte(b[:8])
//	b = b[8:]
//
//	txParam.Nonce = common.GetInt16FromByte(b[:2])
//	b = b[2:]
//	txParam.Chain = b[0]
//	b = b[1:]
//	return txParam, nil
//}
//
//func UnmarshalTx(b []byte, chain uint8, v any) error {
//
//	hash := common.HHash{}
//	hash, err := hash.Init(b[:hash.GetLen()])
//	if err != nil {
//		return err
//	}
//	heightFinal := common.GetInt64FromByte(b[hash.GetLen() : hash.GetLen()+8])
//
//	ln := int(common.GetInt16FromByte(b[hash.GetLen()+8 : hash.GetLen()+10]))
//	if len(b) < ln {
//		return fmt.Errorf("not enough bytes to unmarshal tx signature (nonce msg)")
//	}
//	sig := common.Signature{}
//	btx := b[hash.GetLen()+ln+10:]
//	switch chain {
//	case 0:
//
//		txParam, err := UnmarshalTxParam(btx[:13+common.GVMAddressLength])
//		if err != nil {
//			return err
//		}
//		btx = btx[13+common.GVMAddressLength:]
//
//		height := common.GetInt64FromByte(btx[:8])
//		btx = btx[8:]
//		gasPrice := common.GetInt64FromByte(btx[:8])
//		btx = btx[8:]
//		gasUsage := common.GetInt64FromByte(btx[:8])
//		btx = btx[8:]
//
//		recipient := common.Address{}
//		err = recipient.Init(btx[:common.GVMAddressLength])
//		if err != nil {
//			return err
//		}
//		btx = btx[common.GVMAddressLength:]
//
//		amount := common.GetInt64FromByte(btx[:8])
//		btx = btx[8:]
//		l := common.GetInt32FromByte(btx[:4])
//		btx = btx[4:]
//		optData := btx[:l]
//		btx = btx[l:]
//		l = common.GetInt32FromByte(btx[:4])
//		btx = btx[4:]
//		outPutLogs := btx[:l]
//		btx = btx[l:]
//		contractAddress := common.Address{}
//		contractAddress.Init(btx[:common.GVMAddressLength])
//
//		err = sig.Init(b[hash.GetLen()+10:hash.GetLen()+ln+10], txParam.Sender)
//		if err != nil {
//			return err
//		}
//		td := TxDataMain{
//			Recipient: recipient,
//			Amount:    amount,
//			OptData:   optData,
//		}
//		t := TxMain{
//			TxData:          td,
//			TxParam:         txParam,
//			HHash:           hash,
//			Signature:       sig,
//			Height:          height,
//			GasPrice:        gasPrice,
//			GasUsage:        gasUsage,
//			HeightFinal:     heightFinal,
//			OutputLogs:      outPutLogs,
//			ContractAddress: contractAddress,
//		}
//		*v.(*any) = t
//	case 1:
//
//		txParam, err := UnmarshalTxParam(btx[:13+common.GVMAddressLength])
//		if err != nil {
//			return err
//		}
//		btx = btx[13+common.GVMAddressLength:]
//
//		height := common.GetInt64FromByte(btx[:8])
//		btx = btx[8:]
//		gasPrice := common.GetInt64FromByte(btx[:8])
//		btx = btx[8:]
//		gasUsage := common.GetInt64FromByte(btx[:8])
//		btx = btx[8:]
//		pk := common.PubKey{}
//		err = pk.Init(btx[:pk.GetLen()])
//		if err != nil {
//			return err
//		}
//
//		err = sig.Init(b[hash.GetLen()+10:hash.GetLen()+ln+10], txParam.Sender)
//		if err != nil {
//			return err
//		}
//		td := TxDataSide{Pubkey: pk}
//		t := TxSide{
//			TxData:      td,
//			TxParam:     txParam,
//			HHash:       hash,
//			Signature:   sig,
//			Height:      height,
//			GasUsage:    gasUsage,
//			GasPrice:    gasPrice,
//			HeightFinal: heightFinal,
//		}
//		*v.(*any) = t
//	}
//
//	return nil
//}
