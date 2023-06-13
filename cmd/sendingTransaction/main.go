package main

import (
	"fmt"
	"github.com/chainpqc/chainpqc-node/common"
	"github.com/chainpqc/chainpqc-node/crypto/oqs/rand"
	clientrpc "github.com/chainpqc/chainpqc-node/rpc/client"
	"github.com/chainpqc/chainpqc-node/services/transactionServices"
	"github.com/chainpqc/chainpqc-node/transactionType"
	"github.com/chainpqc/chainpqc-node/wallet"
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

	go sendTransactions(mainWallet)
	chanPeer := make(chan string)

	<-chanPeer
}

func SampleTransaction(w *wallet.Wallet) transactionType.Transaction {

	sender := w.Address
	recv := common.Address{}
	br := rand.RandomBytes(20)
	err := recv.Init(br)
	if err != nil {
		return transactionType.Transaction{}
	}

	txdata := transactionType.TxData{
		Recipient: recv,
		Amount:    int64(rand2.Intn(10000000)),
		OptData:   nil,
	}
	txParam := transactionType.TxParam{
		ChainID:     common.GetChainID(),
		Sender:      sender,
		SendingTime: common.GetCurrentTimeStampInSecond(),
		Nonce:       int16(rand2.Intn(65000)),
		Chain:       0,
	}
	t := transactionType.Transaction{
		TxData:    txdata,
		TxParam:   txParam,
		Hash:      common.Hash{},
		Signature: common.Signature{},
		Height:    0,
		GasPrice:  0,
		GasUsage:  0,
	}

	err = t.CalcHashAndSet()
	if err != nil {
		log.Println("calc hash error", err)
	}
	//err = t.Sign()
	//if err != nil {
	//	log.Println("Signing error", err)
	//}
	s := rand.RandomBytes(common.SignatureLength)
	sig := common.Signature{}
	err = sig.Init(s, w.Address)
	if err != nil {
		return transactionType.Transaction{}
	}
	t.Signature = sig
	return t
}

func sendTransactions(w *wallet.Wallet) {

	chain := uint8(0)
	batchSize := 1
	count := int64(0)
	start := common.GetCurrentTimeStampInSecond()

	for range time.Tick(time.Microsecond) {
		var txs []transactionType.Transaction
		for i := 0; i < batchSize; i++ {
			tx := SampleTransaction(w)
			txs = append(txs, tx)
		}
		m, err := transactionServices.GenerateTransactionMsg(txs, chain, [2]byte{'T', chain})
		if err != nil {
			return
		}
		tmm := m.GetBytes()
		count += int64(batchSize)
		end := common.GetCurrentTimeStampInSecond()
		if count%2 == 0 && (end-start) > 0 {
			fmt.Println("tps=", count/(end-start))
		}
		clientrpc.InRPC <- append([]byte("TRAN"), tmm...)
		<-clientrpc.OutRPC
	}
}
