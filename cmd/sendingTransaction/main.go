package main

import (
	"fmt"
	"github.com/quad-foundation/quad-node/common"
	"github.com/quad-foundation/quad-node/crypto/oqs/rand"
	clientrpc "github.com/quad-foundation/quad-node/rpc/client"
	"github.com/quad-foundation/quad-node/services/transactionServices"
	"github.com/quad-foundation/quad-node/statistics"
	"github.com/quad-foundation/quad-node/transactionsDefinition"
	"github.com/quad-foundation/quad-node/wallet"
	rand2 "math/rand"

	"log"
	"os"
	"time"
)

func main() {
	var ip string
	if len(os.Args) > 1 {
		ip = os.Args[1]
	} else {
		ip = "127.0.0.1"
	}
	go clientrpc.ConnectRPC(ip)
	wallet.InitActiveWallet(0, "a")
	mainWallet := wallet.GetActiveWallet()

	for range 10 {
		go sendTransactions(mainWallet)
		time.Sleep(time.Second)
	}
	chanPeer := make(chan []byte)
	<-chanPeer
}

func SampleTransaction(w *wallet.Wallet) transactionsDefinition.Transaction {

	sender := w.MainAddress
	recv := common.Address{}
	br := rand.RandomBytes(20)
	err := recv.Init(append([]byte{0}, br...))
	if err != nil {
		return transactionsDefinition.Transaction{}
	}

	txdata := transactionsDefinition.TxData{
		Recipient: recv,
		Amount:    int64(rand2.Intn(1000000000)),
		OptData:   nil,
		Pubkey:    common.PubKey{}, //w.PublicKey2, //
	}
	txParam := transactionsDefinition.TxParam{
		ChainID:     common.GetChainID(),
		Sender:      sender,
		SendingTime: common.GetCurrentTimeStampInSecond(),
		Nonce:       int16(rand2.Intn(65000)),
	}
	t := transactionsDefinition.Transaction{
		TxData:    txdata,
		TxParam:   txParam,
		Hash:      common.Hash{},
		Signature: common.Signature{},
		Height:    0,
		GasPrice:  0,
		GasUsage:  0,
	}

	clientrpc.InRPC <- []byte("STAT")
	var reply []byte
	reply = <-clientrpc.OutRPC
	st := statistics.MainStats{}
	err = common.Unmarshal(reply, common.StatDBPrefix, &st)
	if err != nil {
		return transactionsDefinition.Transaction{}
	}
	t.Height = st.Heights

	err = t.CalcHashAndSet()
	if err != nil {
		log.Println("calc hash error", err)
	}
	err = t.Sign(w, w.PublicKey.Primary)
	if err != nil {
		log.Println("Signing error", err)
	}
	//s := rand.RandomBytes(common.SignatureLength)
	//sig := common.Signature{}
	//err = sig.Init(s, w.Address)
	//if err != nil {
	//	return transactionsDefinition.Transaction{}
	//}
	//t.Signature = sig
	return t
}

func sendTransactions(w *wallet.Wallet) {

	batchSize := 2000
	count := int64(0)
	start := common.GetCurrentTimeStampInSecond()

	for range time.Tick(time.Millisecond) {
		var txs []transactionsDefinition.Transaction
		for i := 0; i < batchSize; i++ {
			tx := SampleTransaction(w)
			txs = append(txs, tx)
			end := common.GetCurrentTimeStampInSecond()
			count++
			if count%1000 == 0 && (end-start) > 0 {
				fmt.Println("tps=", count/(end-start), " count: ", count)
			}
		}
		m, err := transactionServices.GenerateTransactionMsg(txs, []byte("tx"), [2]byte{'T', 'T'})
		if err != nil {
			return
		}
		tmm := m.GetBytes()
		//count += int64(batchSize)
		clientrpc.InRPC <- append([]byte("TRAN"), tmm...)
		//log.Printf("send batch %d transactions", batchSize)
		<-clientrpc.OutRPC
		//log.Println("transactions sent")
	}
}
