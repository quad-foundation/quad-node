package nonceServices

import (
	"bytes"
	"github.com/quad-foundation/quad-node/blocks"
	"github.com/quad-foundation/quad-node/common"
	"github.com/quad-foundation/quad-node/message"
	"github.com/quad-foundation/quad-node/services"
	"github.com/quad-foundation/quad-node/tcpip"
	"github.com/quad-foundation/quad-node/transactionsDefinition"
	"github.com/quad-foundation/quad-node/wallet"
	"golang.org/x/exp/rand"
	"log"
	"time"
)

func InitNonceService() {
	services.SendMutexNonce.Lock()
	services.SendChanNonce = make(chan []byte, 10)

	services.SendChanSelfNonce = make(chan []byte, 10)
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

	//TODO Price oracle currently is random: 0.9 - 1.1 QAD/USD
	priceOracle := int64(rand.Intn(10000000) - 5000000 + 100000000)
	randOracle := rand.Int63()
	optData = append(optData, common.GetByteInt64(priceOracle)...)
	optData = append(optData, common.GetByteInt64(randOracle)...)

	// be2, _ := oqs.GenerateBytesFromParams(common.SigName2, common.PubKeyLength2, common.PrivateKeyLength2, common.SignatureLength2, true, true)
	// Encryption1 and Encryption2 when changed than needs to add bytes
	encryption1 := common.BytesToLenAndBytes([]byte{})
	encryption2 := common.BytesToLenAndBytes([]byte{})
	optData = append(optData, encryption1...)
	optData = append(optData, encryption2...)

	dataTx := transactionsDefinition.TxData{
		Recipient: common.GetDelegatedAccount(), // will be delegated account temporary
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
			if len(s) == 4 && bytes.Equal(s, []byte("EXIT")) {
				break Q
			}
		default:
		}
	}
}

func sendNonceMsg(ip [4]byte, topic [2]byte) {
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

func Send(addr [4]byte, nb []byte) {
	nb = append(addr[:], nb...)
	services.SendMutexNonce.Lock()
	services.SendChanNonce <- nb
	services.SendMutexNonce.Unlock()
}

func sendNonceMsgInLoop() {
	for range time.Tick(time.Second * 3) {
		var topic = [2]byte{'N', 'N'}
		sendNonceMsg([4]byte{0, 0, 0, 0}, topic)
	}
}

func startPublishingNonceMsg() {
	go tcpip.StartNewListener(services.SendChanNonce, tcpip.NonceTopic)
	go tcpip.StartNewListener(services.SendChanSelfNonce, tcpip.SelfNonceTopic)
}

func StartSubscribingNonceMsg(ip [4]byte) {
	recvChan := make(chan []byte, 10) // Use a buffered channel
	quit := false
	var ipr [4]byte
	go tcpip.StartNewConnection(ip, recvChan, tcpip.NonceTopic)
	log.Println("Enter connection receiving loop (nonce msg)", ip)
	for !quit {
		select {
		case s := <-recvChan:
			if len(s) == 4 && bytes.Equal(s, []byte("EXIT")) {
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

func StartSubscribingNonceMsgSelf() {
	recvChanSelf := make(chan []byte, 10) // Use a buffered channel
	recvChanExit := make(chan []byte, 10) // Use a buffered channel
	quit := false
	var ip [4]byte
	go tcpip.StartNewConnection(tcpip.MyIP, recvChanSelf, tcpip.SelfNonceTopic)
	go sendNonceMsgInLoopSelf(recvChanExit)
	log.Println("Enter connection receiving loop (nonce msg self)")
	for !quit {
		select {
		case s := <-recvChanSelf:
			if len(s) == 4 && bytes.Equal(s, []byte("EXIT")) {
				recvChanExit <- s
				quit = true
				break
			}
			if len(s) > 4 {
				copy(ip[:], s[:4])
				OnMessage(ip, s[4:])
			}
		case <-tcpip.Quit:
			quit = true
		default:

			// Optional: Add a small sleep to prevent busy-waiting
			time.Sleep(time.Millisecond)
		}
	}
	log.Println("Exit connection receiving loop (nonce msg self)")
}
