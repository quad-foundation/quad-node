package main

import (
	"github.com/chainpqc/chainpqc-node/database"
	"github.com/chainpqc/chainpqc-node/genesis"
	serverrpc "github.com/chainpqc/chainpqc-node/rpc/server"
	nonceService "github.com/chainpqc/chainpqc-node/services/nonceService"
	"github.com/chainpqc/chainpqc-node/services/transactionServices"
	"github.com/chainpqc/chainpqc-node/tcpip"
	"github.com/chainpqc/chainpqc-node/wallet"
	"log"
	"os"
	"time"
)

func main() {
	memDatabase.Init()
	defer memDatabase.CloseDB()
	mainWallet := wallet.EmptyWallet().GetWallet()
	mainWallet.SetPassword("a")
	err := mainWallet.Load()
	if err != nil {
		log.Println("Could not load wallet", err)
		return
	}
	genesis.InitGenesis()

	//firstDel := common.GetDelegatedAccountAddress(1)
	//if firstDel.GetHex() == common.DelegatedAccount.GetHex() {
	//	common.IsSyncing.Store(false)
	//}

	transactionServices.InitTransactionService()
	nonceService.InitNonceService()
	//nonceMsg.InitSyncService()
	//broadcastStaking.InitStakeService()

	go serverrpc.ListenRPC()

	for i := uint8(0); i < 5; i++ {
		go nonceService.StartSubscribingNonceMsgSelf(i)
		go nonceService.StartSubscribingNonceMsg(tcpip.MyIP, i)
	}
	time.Sleep(time.Second)
	if len(os.Args) > 1 {
		ip := os.Args[1]
		for i := uint8(0); i < 5; i++ {
			//go broadcast.StartSubscribingTransaction(ip, 0)
			//go broadcast.StartSubscribingTransaction(ip, 1)
			go nonceService.StartSubscribingNonceMsg(ip, i)
			//go nonceMsg.StartSubscribingSync(ip)
			//go broadcastStaking.StartSubscribingStakingTransaction(ip)
		}
	}
	time.Sleep(time.Second)

	chanPeer := make(chan string)
	go tcpip.LookUpForNewPeersToConnect(chanPeer)
QF:
	for {
		select {
		case topicip := <-chanPeer:
			topic := topicip[:2]
			ip := topicip[2:]
			switch topic {
			case tcpip.SelfNonceTopic[0]:
				go nonceService.StartSubscribingNonceMsgSelf(0)
			case tcpip.SelfNonceTopic[1]:
				go nonceService.StartSubscribingNonceMsgSelf(1)
			case tcpip.SelfNonceTopic[2]:
				go nonceService.StartSubscribingNonceMsgSelf(2)
			case tcpip.SelfNonceTopic[3]:
				go nonceService.StartSubscribingNonceMsgSelf(3)
			case tcpip.SelfNonceTopic[4]:
				go nonceService.StartSubscribingNonceMsgSelf(4)
			//case tcpip.SyncTopic:
			//	go nonceMsg.StartSubscribingSync(ip)
			case tcpip.NonceTopic[0]:
				go nonceService.StartSubscribingNonceMsg(ip, 0)
			case tcpip.NonceTopic[1]:
				go nonceService.StartSubscribingNonceMsg(ip, 1)
			case tcpip.NonceTopic[2]:
				go nonceService.StartSubscribingNonceMsg(ip, 2)
			case tcpip.NonceTopic[3]:
				go nonceService.StartSubscribingNonceMsg(ip, 3)
			case tcpip.NonceTopic[4]:
				go nonceService.StartSubscribingNonceMsg(ip, 4)
				//case tcpip.TransactionTopic[0]:
				//	go broadcast.StartSubscribingTransaction(ip, 0)
				//case tcpip.TransactionTopic[1]:
				//	go broadcast.StartSubscribingTransaction(ip, 1)
				//case tcpip.StakeTopic:
				//	go broadcastStaking.StartSubscribingStakingTransaction(ip)
			}
		case <-tcpip.Quit:
			break QF
		}
	}

}

//func sendTransactionSideChain(t transaction.TxSideType) {
//
//	for range time.Tick(time.Second) {
//		m := message.GenerateTransactionMsg([]transaction2.AnyTransaction{t}, "transaction")
//		m.BaseMessage.ChainID = 23
//		tmm, _ := json.Marshal(m)
//
//		//var r = make([]byte, 100)
//		clientrpc.InRPC <- append([]byte("TRAN"), tmm...)
//		<-clientrpc.OutRPC
//	}
//}
