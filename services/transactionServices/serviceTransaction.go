package transactionServices

import (
	"github.com/chainpqc/chainpqc-node/common"
	"github.com/chainpqc/chainpqc-node/message"
	"github.com/chainpqc/chainpqc-node/services"
	"github.com/chainpqc/chainpqc-node/tcpip"
	"github.com/chainpqc/chainpqc-node/transactionType"
	"log"
	"time"
)

func InitTransactionService() {
	services.RecvMutex.Lock()
	services.RecvChan = make(chan []byte)

	services.RecvMutex.Unlock()
	startPublishingTransactionMsg()
	go broadcastTransactionsMsgInLoop(services.RecvChan)
}

func GenerateTransactionMsg(txs []transactionType.AnyTransaction, chain uint8, topic [2]byte) (message.AnyTransactionsMessage, error) {

	topic[1] = chain
	bm := message.BaseMessage{
		Head:    []byte("tx"),
		ChainID: common.GetChainID(),
		Chain:   chain,
	}
	bb := [][]byte{}
	for _, tx := range txs {
		b := transactionType.GetBytes(tx)
		bb = append(bb, b)
	}

	n := message.AnyTransactionsMessage{
		BaseMessage:       bm,
		TransactionsBytes: map[[2]byte][][]byte{topic: bb},
	}
	return n, nil
}

func broadcastTransactionsMsgInLoop(chanRecv chan []byte) {

Q:
	for range time.Tick(time.Second) {
		chain := common.GetChainForHeight(common.GetHeight() + 1)
		topic := [2]byte{'T', chain}

		SendTransactionMsg(tcpip.MyIP, chain, topic)
		select {
		case s := <-chanRecv:
			if len(s) == 4 && string(s) == "EXIT" {
				break Q
			}
		default:
		}
	}
}

func SendTransactionMsg(ip string, chain uint8, topic [2]byte) {
	isync := common.IsSyncing.Load()
	if isync == true {
		return
	}
	txs := transactionType.GetTransactionsFromToSendPool(int(common.MaxTransactionsPerBlock), chain)
	n, err := GenerateTransactionMsg(txs, chain, topic)
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

func startPublishingTransactionMsg() {
	services.SendMutex.Lock()
	for i := 0; i < 5; i++ {
		go tcpip.StartNewListener(services.SendChan, tcpip.TransactionTopic[i])
	}
	services.SendMutex.Unlock()
}

func StartSubscribingTransactionMsg(ip string, chain uint8) {
	recvChan := make(chan []byte)

	go tcpip.StartNewConnection(ip, recvChan, tcpip.TransactionTopic[chain])
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
