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
	services.SendMutex.Lock()
	services.SendChan = make(chan []byte)

	services.SendChanSelf = make(chan []byte)
	services.SendMutex.Unlock()
	startPublishingTransactionMsg()
	time.Sleep(time.Second)
	go sendTransactionMsgInLoop()
}

func generateTransactionMsg(chain uint8, topic [2]byte) (message.AnyTransactionsMessage, error) {

	topic[1] = chain
	bm := message.BaseMessage{
		Head:    []byte("tx"),
		ChainID: common.GetChainID(),
		Chain:   chain,
	}
	txs := transactionType.GetTransactionsFromToSendPool(int(common.MaxTransactionsPerBlock), chain)
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

func sendTransactionMsgInLoopSelf(chanRecv chan []byte) {
	var topic = [2]byte{'S', '0'}
Q:
	for range time.Tick(time.Second) {
		chain := common.GetChainForHeight(common.GetHeight() + 1)
		sendTransactionMsg(tcpip.MyIP, chain, topic)
		select {
		case s := <-chanRecv:
			if len(s) == 4 && string(s) == "EXIT" {
				break Q
			}
		default:
		}
	}
}

func sendTransactionMsg(ip string, chain uint8, topic [2]byte) {
	isync := common.IsSyncing.Load()
	if isync == true {
		return
	}
	n, err := generateTransactionMsg(chain, topic)
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

func sendTransactionMsgInLoop() {
	for range time.Tick(time.Second * 10) {
		chain := common.GetChainForHeight(common.GetHeight() + 1)
		var topic = [2]byte{'N', '0'}
		sendTransactionMsg("0.0.0.0", chain, topic)
	}
}

func startPublishingTransactionMsg() {
	services.SendMutex.Lock()
	for i := 0; i < 5; i++ {
		go tcpip.StartNewListener(services.SendChan, tcpip.TransactionTopic[i])
		go tcpip.StartNewListener(services.SendChanSelf, tcpip.TransactionTopic[i])
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

func StartSubscribingTransactionMsgSelf(chain uint8) {
	recvChanSelf := make(chan []byte)
	recvChanExit := make(chan []byte)

	go tcpip.StartNewConnection(tcpip.MyIP, recvChanSelf, tcpip.TransactionTopic[chain])
	go sendTransactionMsgInLoopSelf(recvChanExit)
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
