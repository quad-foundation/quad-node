package transactionServices

import (
	"github.com/quad/quad-node/common"
	"github.com/quad/quad-node/message"
	"github.com/quad/quad-node/services"
	"github.com/quad/quad-node/tcpip"
	"github.com/quad/quad-node/transactionType"
	"log"
	"time"
)

func InitTransactionService() {
	services.SendMutexTx.Lock()
	services.SendChanTx = make(chan []byte)

	services.SendMutexTx.Unlock()
	startPublishingTransactionMsg()
	go broadcastTransactionsMsgInLoop(services.SendChanTx)
}

func GenerateTransactionMsg(txs []transactionType.Transaction, chain uint8, topic [2]byte) (message.TransactionsMessage, error) {

	topic[1] = chain
	bm := message.BaseMessage{
		Head:    []byte("tx"),
		ChainID: common.GetChainID(),
		Chain:   chain,
	}
	bb := [][]byte{}
	for _, tx := range txs {
		b := tx.GetBytes()
		bb = append(bb, b)
	}

	n := message.TransactionsMessage{
		BaseMessage:       bm,
		TransactionsBytes: map[[2]byte][][]byte{topic: bb},
	}
	return n, nil
}

func broadcastTransactionsMsgInLoop(chanRecv chan []byte) {

Q:
	for range time.Tick(time.Second) {
		//chain := common.GetChainForHeight(common.GetHeight() + 1)
		for chain := uint8(0); chain < 5; chain++ {
			topic := [2]byte{'T', chain}

			SendTransactionMsg(tcpip.MyIP, chain, topic)
		}
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
	txs := transactionType.PoolsTx[chain].PeekTransactions(int(common.MaxTransactionsPerBlock))
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
	services.SendMutexTx.Lock()
	services.SendChanTx <- nb
	services.SendMutexTx.Unlock()
}

func startPublishingTransactionMsg() {
	services.SendMutexTx.Lock()
	for i := 0; i < 5; i++ {
		go tcpip.StartNewListener(services.SendChanTx, tcpip.TransactionTopic[i])
	}
	services.SendMutexTx.Unlock()
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
