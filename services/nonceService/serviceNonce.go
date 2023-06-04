package nonceServices

import (
	"fmt"
	"github.com/chainpqc/chainpqc-node/common"
	"github.com/chainpqc/chainpqc-node/message"
	"github.com/chainpqc/chainpqc-node/services"
	"github.com/chainpqc/chainpqc-node/tcpip"
	"github.com/chainpqc/chainpqc-node/transactionType"
	transactionType6 "github.com/chainpqc/chainpqc-node/transactionType/contractChainTransaction"
	transactionType5 "github.com/chainpqc/chainpqc-node/transactionType/dexChainTransaction"
	transactionType2 "github.com/chainpqc/chainpqc-node/transactionType/mainChainTransaction"
	transactionType3 "github.com/chainpqc/chainpqc-node/transactionType/pubKeyChainTransaction"
	transactionType4 "github.com/chainpqc/chainpqc-node/transactionType/stakeChainTransaction"
	"github.com/chainpqc/chainpqc-node/wallet"
	"log"
	"time"
)

func InitNonceService() {
	services.SendMutex.Lock()
	services.SendChan = make(chan []byte)

	services.SendChanSelf = make(chan []byte)
	services.SendMutex.Unlock()
	startPublishingNonceMsg()
	time.Sleep(time.Second)
	go sendNonceMsgInLoop()
}

func generateNonceMsg(chain uint8) (message.AnyTransactionsMessage, error) {
	common.HeightMutex.RLock()
	h := common.GetHeight()
	common.HeightMutex.RUnlock()

	var nonceTransaction transactionType.AnyTransaction
	tp := transactionType.TxParam{
		ChainID:     common.GetChainID(),
		Sender:      wallet.EmptyWallet().GetWallet().Address,
		SendingTime: common.GetCurrentTimeStampInSecond(),
		Nonce:       0,
		Chain:       chain,
	}
	lastBlockHash := common.Hash{}.GetBytes() // To change
	optData := common.GetByteInt64(h)
	optData = append(optData, lastBlockHash...)
	topic := [2]byte{}

	switch chain {
	case 0:
		dataTx := transactionType2.MainChainTxData{
			Recipient: common.Address{},
			Amount:    0,
			OptData:   optData,
		}
		nonceTransaction = &transactionType2.MainChainTransaction{
			TxData:    dataTx,
			TxParam:   tp,
			Hash:      common.Hash{},
			Signature: common.Signature{},
			Height:    h + 1,
			GasPrice:  0,
			GasUsage:  0,
		}
		copy(topic[:], "N0")
	case 1:
		dataTx := transactionType3.PubKeyChainTxData{
			Recipient: common.Address{},
			Amount:    0,
			OptData:   optData,
		}
		nonceTransaction = &transactionType3.PubKeyChainTransaction{
			TxData:    dataTx,
			TxParam:   tp,
			Hash:      common.Hash{},
			Signature: common.Signature{},
			Height:    h + 1,
			GasPrice:  0,
			GasUsage:  0,
		}
		copy(topic[:], "N1")
	case 2:
		dataTx := transactionType4.StakeChainTxData{
			Recipient: common.Address{},
			Amount:    0,
			OptData:   optData,
		}
		nonceTransaction = &transactionType4.StakeChainTransaction{
			TxData:    dataTx,
			TxParam:   tp,
			Hash:      common.Hash{},
			Signature: common.Signature{},
			Height:    h + 1,
			GasPrice:  0,
			GasUsage:  0,
		}
		copy(topic[:], "N2")
	case 3:
		dataTx := transactionType5.DexChainTxData{
			Recipient: common.Address{},
			Amount:    0,
			OptData:   optData,
		}
		nonceTransaction = &transactionType5.DexChainTransaction{
			TxData:    dataTx,
			TxParam:   tp,
			Hash:      common.Hash{},
			Signature: common.Signature{},
			Height:    h + 1,
			GasPrice:  0,
			GasUsage:  0,
		}
		copy(topic[:], "N3")
	case 4:
		dataTx := transactionType6.ContractChainTxData{
			Recipient: common.Address{},
			Amount:    0,
			OptData:   optData,
		}
		nonceTransaction = &transactionType6.ContractChainTransaction{
			TxData:    dataTx,
			TxParam:   tp,
			Hash:      common.Hash{},
			Signature: common.Signature{},
			Height:    h + 1,
			GasPrice:  0,
			GasUsage:  0,
		}
		copy(topic[:], "N4")
	default:
		return message.AnyTransactionsMessage{}, fmt.Errorf("chain is not correct")
	}
	hash, err := nonceTransaction.CalcHash()
	if err != nil {
		return message.AnyTransactionsMessage{}, err
	}
	nonceTransaction.SetHash(hash)
	bm := message.BaseMessage{
		Head:    []byte("nn"),
		ChainID: common.GetChainID(),
		Chain:   chain,
	}

	bb, err := transactionType.SignTransactionAllToBytes(nonceTransaction)
	if err != nil {
		return message.AnyTransactionsMessage{}, fmt.Errorf("error signing transaction: %v", err)
	}
	n := message.AnyTransactionsMessage{
		BaseMessage:       bm,
		TransactionsBytes: map[[2]byte][][]byte{topic: {bb}},
	}

	return n, nil
}

func sendNonceMsgInLoopSelf(chanRecv chan []byte) {

Q:
	for range time.Tick(time.Second) {
		chain := common.GetChainForHeight(common.GetHeight() + 1)
		sendNonceMsg(tcpip.MyIP, chain)
		select {
		case s := <-chanRecv:
			if len(s) == 4 && string(s) == "EXIT" {
				break Q
			}
		default:
		}
	}
}

func sendNonceMsg(ip string, chain uint8) {
	isync := common.IsSyncing.Load()
	if isync == true {
		return
	}
	n, err := generateNonceMsg(chain)
	if err != nil {
		log.Println(err)
		return
	}
	Send(ip, n.GetBytes())
}

func Send(addr string, nb []byte) {
	bip := []byte(addr)
	lip := common.GetByteInt16(int16(len(bip)))
	lip = append(lip, bip...)
	nb = append(lip, nb...)
	services.SendMutex.Lock()
	services.SendChan <- nb
	services.SendMutex.Unlock()
}

func sendNonceMsgInLoop() {
	for range time.Tick(time.Second * 10) {
		chain := common.GetChainForHeight(common.GetHeight() + 1)
		sendNonceMsg("0.0.0.0", chain)
	}
}

func startPublishingNonceMsg() {
	services.SendMutex.Lock()
	for i := 0; i < 5; i++ {
		go tcpip.StartNewListener(services.SendChan, tcpip.NonceTopic[i])
		go tcpip.StartNewListener(services.SendChanSelf, tcpip.SelfNonceTopic[i])
	}
	services.SendMutex.Unlock()
}

func StartSubscribingNonceMsg(ip string, chain uint8) {
	recvChan := make(chan []byte)

	go tcpip.StartNewConnection(ip, recvChan, tcpip.NonceTopic[chain])
	log.Println("Enter connection receiving loop (nonce msg)", ip)
Q:

	for {
		select {
		case s := <-recvChan:
			if len(s) == 4 && string(s) == "EXIT" {
				break Q
			}
			if len(s) > 2 {
				l := common.GetInt16FromByte(s[:2])
				if len(s) > 2+int(l) {
					ipr := string(s[2 : 2+l])

					OnMessage(ipr, s[2+l:])
				}
			}

		case <-tcpip.Quit:
			break Q
		default:
		}

	}
	log.Println("Exit connection receiving loop (nonce msg)", ip)
}

func StartSubscribingNonceMsgSelf(chain uint8) {
	recvChanSelf := make(chan []byte)
	recvChanExit := make(chan []byte)

	go tcpip.StartNewConnection(tcpip.MyIP, recvChanSelf, tcpip.SelfNonceTopic[chain])
	go sendNonceMsgInLoopSelf(recvChanExit)
	log.Println("Enter connection receiving loop (nonce msg self)")
Q:

	for {
		select {
		case s := <-recvChanSelf:
			if len(s) == 4 && string(s) == "EXIT" {
				recvChanExit <- s
				break Q
			}
			if len(s) > 2 {
				l := common.GetInt16FromByte(s[:2])
				if len(s) > 2+int(l) {
					ipr := string(s[2 : 2+l])

					OnMessage(ipr, s[2+l:])
				}
			}
		case <-tcpip.Quit:
			break Q
		default:
		}

	}
	log.Println("Exit connection receiving loop (nonce msg self)")
}
