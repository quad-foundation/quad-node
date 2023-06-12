package main

import (
	"fmt"
	"github.com/chainpqc/chainpqc-node/common"
	"github.com/chainpqc/chainpqc-node/database"
	"github.com/chainpqc/chainpqc-node/genesis"
	serverrpc "github.com/chainpqc/chainpqc-node/rpc/server"
	nonceService "github.com/chainpqc/chainpqc-node/services/nonceService"
	"github.com/chainpqc/chainpqc-node/services/transactionServices"
	"github.com/chainpqc/chainpqc-node/statistics"
	"github.com/chainpqc/chainpqc-node/tcpip"
	"github.com/chainpqc/chainpqc-node/transactionType"
	"github.com/chainpqc/chainpqc-node/wallet"
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"os"
	"time"
)

func main() {
	fmt.Print("Enter password: ")
	password, err := terminal.ReadPassword(0)
	if err != nil {
		log.Fatal(err)
	}
	wallet.InitActiveWallet(0, string(password))

	memDatabase.Init()
	defer memDatabase.CloseDB()
	transactionType.InitPermanentTrie()
	defer transactionType.GlobalMerkleTree.Destroy()
	statistics.InitGlobalMainStats()
	defer statistics.DestroyGlobalMainStats()
	w := wallet.GetActiveWallet()

	err = memDatabase.MainDB.Put(append(common.PubKeyDBPrefix[:], w.Address.GetBytes()...),
		w.PublicKey.GetBytes())
	genesis.InitGenesis()

	//firstDel := common.GetDelegatedAccountAddress(1)
	//if firstDel.GetHex() == common.DelegatedAccount.GetHex() {
	//	common.IsSyncing.Put(false)
	//}

	transactionServices.InitTransactionService()
	nonceService.InitNonceService()
	//nonceMsg.InitSyncService()

	go serverrpc.ListenRPC()

	for i := uint8(0); i < 5; i++ {
		go nonceService.StartSubscribingNonceMsgSelf(i)
		go nonceService.StartSubscribingNonceMsg(tcpip.MyIP, i)
		go transactionServices.StartSubscribingTransactionMsg(tcpip.MyIP, i)
	}
	time.Sleep(time.Second)
	if len(os.Args) > 1 {
		ip := os.Args[1]
		for i := uint8(0); i < 5; i++ {
			go transactionServices.StartSubscribingTransactionMsg(ip, i)
			go nonceService.StartSubscribingNonceMsg(ip, i)
			//go nonceMsg.StartSubscribingSync(ip)
		}
	}
	time.Sleep(time.Second)

	chanPeer := make(chan string)
	go tcpip.LookUpForNewPeersToConnect(chanPeer)
	topic := [2]byte{}
QF:
	for {
		select {
		case topicip := <-chanPeer:
			copy(topic[:], topicip[:2])
			ip := topicip[2:]
			chain := topic[1]
			if topic[0] == 'T' {
				go transactionServices.StartSubscribingTransactionMsg(ip, chain)
			}
			if topic[0] == 'N' {
				go nonceService.StartSubscribingNonceMsg(ip, chain)
			}
			if topic[0] == 'S' {
				go nonceService.StartSubscribingNonceMsgSelf(chain)
			}
			if topic[0] == 'B' {
				// to be implemented
				//go StartSu
			}

		case <-tcpip.Quit:
			break QF
		}
	}

}

//func sendTransactionSideChain(t transaction.TxSideType) {
//
//	for range time.Tick(time.Second) {
//		m := message.GenerateTransactionMsg([]transaction2.Transaction{t}, "transaction")
//		m.BaseMessage.ChainID = 23
//		tmm, _ := json.Marshal(m)
//
//		//var r = make([]byte, 100)
//		clientrpc.InRPC <- append([]byte("TRAN"), tmm...)
//		<-clientrpc.OutRPC
//	}
//}
