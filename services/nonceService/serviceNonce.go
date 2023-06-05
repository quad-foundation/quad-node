package nonceServices

import (
	"fmt"
	"github.com/chainpqc/chainpqc-node/blocks"
	"github.com/chainpqc/chainpqc-node/common"
	"github.com/chainpqc/chainpqc-node/message"
	"github.com/chainpqc/chainpqc-node/services"
	"github.com/chainpqc/chainpqc-node/tcpip"
	"github.com/chainpqc/chainpqc-node/transactionType"
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

func generateNonceMsg(chain uint8, topic [2]byte) (message.AnyTransactionsMessage, error) {
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
	lastBlockHash, err := blocks.LoadHashOfBlock(h)
	if err != nil {
		return message.AnyTransactionsMessage{}, err
	}
	optData := common.GetByteInt64(h)
	optData = append(optData, lastBlockHash.GetBytes()...)

	switch chain {
	case 0:
		dataTx := transactionType.MainChainTxData{
			Recipient: common.EmptyAddress(),
			Amount:    0,
			OptData:   optData,
		}
		nonceTransaction = &transactionType.MainChainTransaction{
			TxData:    dataTx,
			TxParam:   tp,
			Hash:      common.Hash{},
			Signature: common.Signature{},
			Height:    h + 1,
			GasPrice:  0,
			GasUsage:  0,
		}
		topic[1] = byte('0')
	case 1:
		dataTx := transactionType.PubKeyChainTxData{
			Recipient: common.EmptyAddress(),
			Amount:    0,
			OptData:   optData,
		}
		nonceTransaction = &transactionType.PubKeyChainTransaction{
			TxData:    dataTx,
			TxParam:   tp,
			Hash:      common.Hash{},
			Signature: common.Signature{},
			Height:    h + 1,
			GasPrice:  0,
			GasUsage:  0,
		}
		topic[1] = byte('1')
	case 2:
		dataTx := transactionType.StakeChainTxData{
			Recipient: common.EmptyAddress(),
			Amount:    0,
			OptData:   optData,
		}
		nonceTransaction = &transactionType.StakeChainTransaction{
			TxData:    dataTx,
			TxParam:   tp,
			Hash:      common.Hash{},
			Signature: common.Signature{},
			Height:    h + 1,
			GasPrice:  0,
			GasUsage:  0,
		}
		topic[1] = byte('2')
	case 3:
		dataTx := transactionType.DexChainTxData{
			Recipient: common.EmptyAddress(),
			Amount:    0,
			OptData:   optData,
		}
		nonceTransaction = &transactionType.DexChainTransaction{
			TxData:    dataTx,
			TxParam:   tp,
			Hash:      common.Hash{},
			Signature: common.Signature{},
			Height:    h + 1,
			GasPrice:  0,
			GasUsage:  0,
		}
		topic[1] = byte('3')
	case 4:
		dataTx := transactionType.ContractChainTxData{
			Recipient: common.EmptyAddress(),
			Amount:    0,
			OptData:   optData,
		}
		nonceTransaction = &transactionType.ContractChainTransaction{
			TxData:    dataTx,
			TxParam:   tp,
			Hash:      common.Hash{},
			Signature: common.Signature{},
			Height:    h + 1,
			GasPrice:  0,
			GasUsage:  0,
		}
		topic[1] = byte('4')
	default:
		return message.AnyTransactionsMessage{}, fmt.Errorf("chain is not correct")
	}
	hash, err := nonceTransaction.CalcHash()
	if err != nil {
		return message.AnyTransactionsMessage{}, err
	}
	nonceTransaction.SetHash(hash)

	signature, err := transactionType.SignTransaction(nonceTransaction)
	if err != nil {
		return message.AnyTransactionsMessage{}, err
	}
	nonceTransaction.SetSignature(signature)

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
	var topic = [2]byte{'S', '0'}
Q:
	for range time.Tick(time.Second) {
		chain := common.GetChainForHeight(common.GetHeight() + 1)
		sendNonceMsg(tcpip.MyIP, chain, topic)
		select {
		case s := <-chanRecv:
			if len(s) == 4 && string(s) == "EXIT" {
				break Q
			}
		default:
		}
	}
}

func sendNonceMsg(ip string, chain uint8, topic [2]byte) {
	isync := common.IsSyncing.Load()
	if isync == true {
		return
	}
	n, err := generateNonceMsg(chain, topic)
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
		var topic = [2]byte{'N', '0'}
		sendNonceMsg("0.0.0.0", chain, topic)
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
