package main

import (
	"fmt"
	"github.com/quad/quad-node/common"
	"github.com/quad/quad-node/crypto/oqs/rand"
	clientrpc "github.com/quad/quad-node/rpc/client"
	"github.com/quad/quad-node/services/transactionServices"
	"github.com/quad/quad-node/transactionsDefinition"
	"github.com/quad/quad-node/wallet"
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

func SampleTransaction(w *wallet.Wallet, chain uint8) transactionsDefinition.Transaction {

	sender := w.Address
	recv := common.Address{}
	br := rand.RandomBytes(20)
	err := recv.Init(br)
	if err != nil {
		return transactionsDefinition.Transaction{}
	}

	txdata := transactionsDefinition.TxData{
		Recipient: recv,
		Amount:    int64(rand2.Intn(10000000)),
		OptData:   nil,
		Pubkey:    w.PublicKey,
	}
	txParam := transactionsDefinition.TxParam{
		ChainID:     common.GetChainID(),
		Sender:      sender,
		SendingTime: common.GetCurrentTimeStampInSecond(),
		Nonce:       int16(rand2.Intn(65000)),
		Chain:       chain,
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

	err = t.CalcHashAndSet()
	if err != nil {
		log.Println("calc hash error", err)
	}
	err = t.Sign()
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

	batchSize := 1
	count := int64(0)
	start := common.GetCurrentTimeStampInSecond()

	for range time.Tick(time.Millisecond) {
		var txs []transactionsDefinition.Transaction
		chain := uint8(rand2.Intn(5))
		for i := 0; i < batchSize; i++ {
			tx := SampleTransaction(w, chain)
			txs = append(txs, tx)
		}
		m, err := transactionServices.GenerateTransactionMsg(txs, chain, [2]byte{'T', chain})
		if err != nil {
			return
		}
		tmm := m.GetBytes()
		count += int64(batchSize)
		end := common.GetCurrentTimeStampInSecond()
		if count%100 == 0 && (end-start) > 0 {
			fmt.Println("tps=", count/(end-start))
		}
		clientrpc.InRPC <- append([]byte("TRAN"), tmm...)
		<-clientrpc.OutRPC
	}
}
