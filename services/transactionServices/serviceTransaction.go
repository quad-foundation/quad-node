package transactionServices

import (
	"bytes"
	"github.com/quad-foundation/quad-node/common"
	"github.com/quad-foundation/quad-node/message"
	"github.com/quad-foundation/quad-node/services"
	"github.com/quad-foundation/quad-node/tcpip"
	"github.com/quad-foundation/quad-node/transactionsDefinition"
	"github.com/quad-foundation/quad-node/transactionsPool"
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

func GenerateTransactionMsg(txs []transactionsDefinition.Transaction, mesgHead []byte, topic [2]byte) (message.TransactionsMessage, error) {

	bm := message.BaseMessage{
		Head:    mesgHead,
		ChainID: common.GetChainID(),
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

func GenerateTransactionMsgGT(txsHashes [][]byte, mesgHead []byte, topic [2]byte) (message.TransactionsMessage, error) {

	bm := message.BaseMessage{
		Head:    mesgHead,
		ChainID: common.GetChainID(),
	}

	n := message.TransactionsMessage{
		BaseMessage:       bm,
		TransactionsBytes: map[[2]byte][][]byte{topic: txsHashes},
	}
	return n, nil
}

func broadcastTransactionsMsgInLoop(chanRecv chan []byte) {

Q:
	for range time.Tick(time.Second) {

		topic := [2]byte{'T', 'T'}

		SendTransactionMsg(tcpip.MyIP, topic)

		select {
		case s := <-chanRecv:
			if len(s) == 4 && string(s) == "EXIT" {
				break Q
			}
		default:
		}
	}
}

func SendTransactionMsg(ip [4]byte, topic [2]byte) {
	isync := common.IsSyncing.Load()
	if isync == true {
		return
	}
	txs := transactionsPool.PoolsTx.PeekTransactions(int(common.MaxTransactionsPerBlock))
	n, err := GenerateTransactionMsg(txs, []byte("tx"), topic)
	if err != nil {
		log.Println(err)
		return
	}
	Send(ip, n.GetBytes())
}

func SendGT(ip [4]byte, txsHashes [][]byte) {
	topic := tcpip.TransactionTopic
	transactionMsg, err := GenerateTransactionMsgGT(txsHashes, []byte("st"), topic)
	if err != nil {
		log.Println("cannot generate transaction msg", err)
	}
	Send(ip, transactionMsg.GetBytes())
}

func Send(addr [4]byte, nb []byte) {

	nb = append(addr[:], nb...)
	services.SendMutexTx.Lock()
	services.SendChanTx <- nb
	services.SendMutexTx.Unlock()
}

func Spread(ignoreAddr [4]byte, nb []byte) {
	var ip [4]byte
	var peers = tcpip.GetPeersConnected(tcpip.TransactionTopic)
	for topicip, _ := range peers {
		copy(ip[:], topicip[2:])
		if bytes.Compare(ip[:], ignoreAddr[:]) != 0 && bytes.Compare(ip[:], tcpip.MyIP[:]) != 0 {
			Send(ip, nb)
		}
	}
}

func startPublishingTransactionMsg() {
	go tcpip.StartNewListener(services.SendChanTx, tcpip.TransactionTopic)
}

func StartSubscribingTransactionMsg(ip [4]byte) {
	recvChan := make(chan []byte, 10) // Use a buffered channel
	quit := false
	var ipr [4]byte
	go tcpip.StartNewConnection(ip, recvChan, tcpip.TransactionTopic)
	log.Println("Enter connection receiving loop (nonce msg)", ip)
	for !quit {
		select {
		case s := <-recvChan:
			if len(s) == 4 && bytes.Compare(s, []byte("EXIT")) == 0 {
				quit = true
				break
			}
			if len(s) > 4 {
				copy(ipr[:], s[:4])
				OnMessage(ipr, s[4:])
			}
		case <-tcpip.Quit:
			quit = true
		default:
			// Optional: Add a small sleep to prevent busy-waiting
			time.Sleep(time.Millisecond)
		}
	}
	log.Println("Exit connection receiving loop (nonce msg)", ip)
}
