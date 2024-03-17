package main

import (
	"github.com/quad/quad-node/account"
	"github.com/quad/quad-node/account/stake"
	"github.com/quad/quad-node/common"
	"github.com/quad/quad-node/database"
	"github.com/quad/quad-node/genesis"
	serverrpc "github.com/quad/quad-node/rpc/server"
	nonceService "github.com/quad/quad-node/services/nonceService"
	syncServices "github.com/quad/quad-node/services/syncService"
	"github.com/quad/quad-node/services/transactionServices"
	"github.com/quad/quad-node/statistics"
	"github.com/quad/quad-node/tcpip"
	"github.com/quad/quad-node/transactionsPool"
	"github.com/quad/quad-node/wallet"
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
	err = account.StoreAccounts()
	if err != nil {
		log.Fatal(err)
	}
	sa := stake.StakingAccount{
		StakedBalance:  0,
		StakingRewards: 0,
		Address:        addrbytes,
		StakingDetails: nil,
	}
	allStakingAccounts := map[[20]byte]stake.StakingAccount{}
	allStakingAccounts[addrbytes] = sa
	account.StakingAccounts = account.StakingAccountsType{AllStakingAccounts: allStakingAccounts}

	err = account.StoreStakingAccounts()
	if err != nil {
		log.Fatal(err)
	}
	transactionsPool.InitPermanentTrie()
	defer transactionsPool.GlobalMerkleTree.Destroy()
	statistics.InitGlobalMainStats()
	defer statistics.DestroyGlobalMainStats()
	err = account.LoadAccounts()
	if err != nil {
		log.Fatal(err)
	}
	defer account.StoreAccounts()
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

	for i := uint8(0); i < 5; i++ {
		go nonceService.StartSubscribingNonceMsgSelf(i)
		go nonceService.StartSubscribingNonceMsg(tcpip.MyIP, i)
		go transactionServices.StartSubscribingTransactionMsg(tcpip.MyIP, i)
		go syncServices.StartSubscribingSyncMsg(tcpip.MyIP, i)
	}
	time.Sleep(time.Second)
	if len(os.Args) > 1 {
		ip := os.Args[1]
		for i := uint8(0); i < 5; i++ {
			go transactionServices.StartSubscribingTransactionMsg(ip, i)
			go nonceService.StartSubscribingNonceMsg(ip, i)
			go syncServices.StartSubscribingSyncMsg(ip, i)
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
				go syncServices.StartSubscribingSyncMsg(ip, chain)
			}

		case <-tcpip.Quit:
			break QF
		}
	}

}
