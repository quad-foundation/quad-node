package nonceServices

import (
	"github.com/quad-foundation/quad-node/blocks"
	"github.com/quad-foundation/quad-node/common"
	"github.com/quad-foundation/quad-node/message"
	"github.com/quad-foundation/quad-node/services"
	"github.com/quad-foundation/quad-node/tcpip"
	"github.com/quad-foundation/quad-node/transactionsDefinition"
	"github.com/quad-foundation/quad-node/wallet"
	"log"
	"net"
	"time"
)

func InitNonceService() {
	services.SendMutexNonce.Lock()
	services.SendChanNonce = make(chan []byte)

	services.SendChanSelfNonce = make(chan []byte)
	services.SendMutexNonce.Unlock()
	startPublishingNonceMsg()
	time.Sleep(time.Second)
	go sendNonceMsgInLoop()
}

func generateNonceMsg(topic [2]byte) (message.TransactionsMessage, error) {
	h := common.GetHeight()

	var nonceTransaction transactionsDefinition.Transaction
	tp := transactionsDefinition.TxParam{
		ChainID:     common.GetChainID(),
		Sender:      wallet.GetActiveWallet().Address,
		SendingTime: common.GetCurrentTimeStampInSecond(),
		Nonce:       0,
	}
	lastBlockHash, err := blocks.LoadHashOfBlock(h)
	if err != nil {
		lastBlockHash = common.EmptyHash().GetBytes()
	}
	optData := common.GetByteInt64(h)
	optData = append(optData, lastBlockHash...)

	dataTx := transactionsDefinition.TxData{
		Recipient: common.EmptyAddress(),
		Amount:    0,
		OptData:   optData[:],
	}
	nonceTransaction = transactionsDefinition.Transaction{
		TxData:    dataTx,
		TxParam:   tp,
		Hash:      common.Hash{},
		Signature: common.Signature{},
		Height:    h + 1,
		GasPrice:  0,
		GasUsage:  0,
	}

	err = (&nonceTransaction).CalcHashAndSet()
	if err != nil {
		return message.TransactionsMessage{}, err
	}

	err = (&nonceTransaction).Sign(wallet.GetActiveWallet())
	if err != nil {
		return message.TransactionsMessage{}, err
	}

	bm := message.BaseMessage{
		Head:    []byte("nn"),
		ChainID: common.GetChainID(),
	}
	bb := nonceTransaction.GetBytes()
	n := message.TransactionsMessage{
		BaseMessage:       bm,
		TransactionsBytes: map[[2]byte][][]byte{topic: {bb}},
	}

	return n, nil
}

func sendNonceMsgInLoopSelf(chanRecv chan []byte) {
	var topic = [2]byte{'S', 'S'}
Q:
	for range time.Tick(time.Second) {
		sendNonceMsg(tcpip.MyIP, topic)
		select {
		case s := <-chanRecv:
			if len(s) == 4 && string(s) == "EXIT" {
				break Q
			}
		default:
		}
	}
}

func sendNonceMsg(ip string, topic [2]byte) {
	isync := common.IsSyncing.Load()
	if isync == true {
		return
	}
	n, err := generateNonceMsg(topic)
	if err != nil {
		log.Println(err)
		return
	}
	Send(ip, n.GetBytes())
}

func Send(addr string, nb []byte) {
	nb = append(net.ParseIP(addr).To4(), nb...)
	services.SendMutexNonce.Lock()
	services.SendChanNonce <- nb
	services.SendMutexNonce.Unlock()
}

func sendNonceMsgInLoop() {
	for range time.Tick(time.Second * 10) {
		var topic = [2]byte{'N', 'N'}
		sendNonceMsg("0.0.0.0", topic)
	}
}

func startPublishingNonceMsg() {

	go tcpip.StartNewListener(services.SendChanNonce, tcpip.NonceTopic)
	go tcpip.StartNewListener(services.SendChanSelfNonce, tcpip.SelfNonceTopic)

}

func StartSubscribingNonceMsg(ip string) {
	recvChan := make(chan []byte)

	go tcpip.StartNewConnection(ip, recvChan, tcpip.NonceTopic)
	log.Println("Enter connection receiving loop (nonce msg)", ip)
Q:

	for {
		select {
		case s := <-recvChan:
			if len(s) == 4 && string(s) == "EXIT" {
				break Q
			}
			if len(s) > 4 {
				ipr := net.IPv4(s[0], s[1], s[2], s[3]).String()
				OnMessage(ipr, s[4:])
			}

		case <-tcpip.Quit:
			break Q
		default:
		}

	}
	log.Println("Exit connection receiving loop (nonce msg)", ip)
}

func StartSubscribingNonceMsgSelf() {
	recvChanSelf := make(chan []byte)
	recvChanExit := make(chan []byte)

	go tcpip.StartNewConnection(tcpip.MyIP, recvChanSelf, tcpip.SelfNonceTopic)
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
			if len(s) > 4 {
				ipr := net.IPv4(s[0], s[1], s[2], s[3]).String()
				OnMessage(ipr, s[4:])
			}
		case <-tcpip.Quit:
			break Q
		default:
		}

	}
	log.Println("Exit connection receiving loop (nonce msg self)")
}
