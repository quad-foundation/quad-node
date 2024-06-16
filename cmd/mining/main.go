package main

import (
	"github.com/quad-foundation/quad-node/account"
	"github.com/quad-foundation/quad-node/common"
	memDatabase "github.com/quad-foundation/quad-node/database"
	"github.com/quad-foundation/quad-node/genesis"
	serverrpc "github.com/quad-foundation/quad-node/rpc/server"
	nonceService "github.com/quad-foundation/quad-node/services/nonceService"
	syncServices "github.com/quad-foundation/quad-node/services/syncService"
	"github.com/quad-foundation/quad-node/services/transactionServices"
	"github.com/quad-foundation/quad-node/statistics"
	"github.com/quad-foundation/quad-node/tcpip"
	"github.com/quad-foundation/quad-node/transactionsPool"
	"github.com/quad-foundation/quad-node/wallet"
	"log"
	"os"
	"time"
)

func main() {
	//fmt.Print("Enter password: ")
	//password, err := terminal.ReadPassword(0)
	//if err != nil {
	//	log.Fatal(err)
	//}
	var err error
	password := "a"
	wallet.InitActiveWallet(0, string(password))
	addrbytes := [common.AddressLength]byte{}
	copy(addrbytes[:], wallet.GetActiveWallet().Address.GetBytes())
	memDatabase.Init()
	defer memDatabase.CloseDB()

	a := account.Account{
		Balance: 0,
		Address: addrbytes,
	}
	allAccounts := map[[20]byte]account.Account{}
	allAccounts[addrbytes] = a
	account.Accounts = account.AccountsType{AllAccounts: allAccounts}
	err = account.StoreAccounts(0)
	if err != nil {
		log.Fatal(err)
	}

	for i := 1; i < 256; i++ {
		del := common.GetDelegatedAccountAddress(int16(i))
		delbytes := [common.AddressLength]byte{}
		copy(delbytes[:], del.GetBytes())
		sa := account.StakingAccount{
			StakedBalance:    0,
			StakingRewards:   0,
			DelegatedAccount: delbytes,
			StakingDetails:   nil,
		}
		allStakingAccounts := map[[20]byte]account.StakingAccount{}
		allStakingAccounts[addrbytes] = sa
		account.StakingAccounts[i] = account.StakingAccountsType{AllStakingAccounts: allStakingAccounts}
	}
	err = account.StoreStakingAccounts()
	if err != nil {
		log.Fatal(err)
	}
	transactionsPool.InitPermanentTrie()
	defer transactionsPool.GlobalMerkleTree.Destroy()
	statistics.InitGlobalMainStats()
	defer statistics.DestroyGlobalMainStats()
	err = account.LoadAccounts(-1)
	if err != nil {
		log.Fatal(err)
	}
	defer account.StoreAccounts(-1)
	err = account.LoadStakingAccounts()
	if err != nil {
		log.Fatal(err)
	}
	defer account.StoreStakingAccounts()

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
	syncServices.InitSyncService()

	go serverrpc.ListenRPC()

	go nonceService.StartSubscribingNonceMsgSelf()
	go nonceService.StartSubscribingNonceMsg(tcpip.MyIP)
	go transactionServices.StartSubscribingTransactionMsg(tcpip.MyIP)
	go syncServices.StartSubscribingSyncMsg(tcpip.MyIP)
	time.Sleep(time.Second)
	if len(os.Args) > 1 {
		ip := os.Args[1]

		go transactionServices.StartSubscribingTransactionMsg(ip)
		go nonceService.StartSubscribingNonceMsg(ip)
		go syncServices.StartSubscribingSyncMsg(ip)
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

			if topic[0] == 'T' {
				go transactionServices.StartSubscribingTransactionMsg(ip)
			}
			if topic[0] == 'N' {
				go nonceService.StartSubscribingNonceMsg(ip)
			}
			if topic[0] == 'S' {
				go nonceService.StartSubscribingNonceMsgSelf()
			}
			if topic[0] == 'B' {
				go syncServices.StartSubscribingSyncMsg(ip)
			}

		case <-tcpip.Quit:
			break QF
		}
	}

}
