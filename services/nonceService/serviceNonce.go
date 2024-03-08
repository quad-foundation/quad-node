package nonceServices

import (
	"github.com/quad/quad-node/blocks"
	"github.com/quad/quad-node/common"
	"github.com/quad/quad-node/message"
	"github.com/quad/quad-node/services"
	"github.com/quad/quad-node/tcpip"
	"github.com/quad/quad-node/transactionType"
	"github.com/quad/quad-node/wallet"
	"log"
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

func generateNonceMsg(chain uint8, topic [2]byte) (message.TransactionsMessage, error) {
	h := common.GetHeight()

	var nonceTransaction transactionType.Transaction
	tp := transactionType.TxParam{
		ChainID:     common.GetChainID(),
		Sender:      wallet.GetActiveWallet().Address,
		SendingTime: common.GetCurrentTimeStampInSecond(),
		Nonce:       0,
		Chain:       chain,
	}
	lastBlockHash, err := blocks.LoadHashOfBlock(h)
	if err != nil {
		lastBlockHash = common.EmptyHash().GetBytes()
	}
	optData := common.GetByteInt64(h)
	optData = append(optData, lastBlockHash...)

	dataTx := transactionType.TxData{
		Recipient: common.EmptyAddress(),
		Amount:    0,
		OptData:   optData[:],
	}
	nonceTransaction = transactionType.Transaction{
		TxData:    dataTx,
		TxParam:   tp,
		Hash:      common.Hash{},
		Signature: common.Signature{},
		Height:    h + 1,
		GasPrice:  0,
		GasUsage:  0,
	}
	topic[1] = chain

	err = (&nonceTransaction).CalcHashAndSet()
	if err != nil {
		return message.TransactionsMessage{}, err
	}

	err = (&nonceTransaction).Sign()
	if err != nil {
		return message.TransactionsMessage{}, err
	}

	bm := message.BaseMessage{
		Head:    []byte("nn"),
		ChainID: common.GetChainID(),
		Chain:   chain,
	}
	bb := nonceTransaction.GetBytes()
	n := message.TransactionsMessage{
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
		topic[1] = chain
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
	services.SendMutexNonce.Lock()
	services.SendChanNonce <- nb
	services.SendMutexNonce.Unlock()
}

func sendNonceMsgInLoop() {
	for range time.Tick(time.Second * 10) {
		chain := common.GetChainForHeight(common.GetHeight() + 1)
		var topic = [2]byte{'N', chain}
		sendNonceMsg("0.0.0.0", chain, topic)
	}
}

func startPublishingNonceMsg() {
	services.SendMutexNonce.Lock()
	for i := 0; i < 5; i++ {
		go tcpip.StartNewListener(services.SendChanNonce, tcpip.NonceTopic[i])
		go tcpip.StartNewListener(services.SendChanSelfNonce, tcpip.SelfNonceTopic[i])
	}
	services.SendMutexNonce.Unlock()
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
