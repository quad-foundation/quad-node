package main

import (
	"fmt"
	"github.com/quad-foundation/quad-node/account"
	"github.com/quad-foundation/quad-node/blocks"
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
	"net/http"
	"strconv"
	"strings"

	_ "net/http/pprof"
	"os"
	"time"
)

func main() {
	// profiling
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

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

	// DEX acount init

	allDexAccounts := map[[20]byte]account.DexAccount{}
	account.DexAccounts = account.DexAccountsType{AllDexAccounts: allDexAccounts}
	err = account.StoreDexAccounts(0)
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
	err = account.StoreStakingAccounts(0)
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

	err = account.LoadDexAccounts(-1)
	if err != nil {
		log.Fatal(err)
	}
	defer account.StoreDexAccounts(-1)

	err = account.LoadStakingAccounts(-1)
	if err != nil {
		log.Fatal(err)
	}
	defer account.StoreStakingAccounts(-1)

	w := wallet.GetActiveWallet()

	err = memDatabase.MainDB.Put(append(common.PubKeyDBPrefix[:], w.Address.GetBytes()...),
		w.PublicKey.GetBytes())

	blocks.InitStateDB()

	genesis.InitGenesis()

	transactionServices.InitTransactionService()
	syncServices.InitSyncService()

	go serverrpc.ListenRPC()

	firstDel := common.GetDelegatedAccountAddress(1)
	if firstDel.GetHex() == common.GetDelegatedAccount().Hex() {
		nonceService.InitNonceService()
		go nonceService.StartSubscribingNonceMsgSelf()
		go nonceService.StartSubscribingNonceMsg(tcpip.MyIP)
	}

	go transactionServices.StartSubscribingTransactionMsg(tcpip.MyIP)
	go syncServices.StartSubscribingSyncMsg(tcpip.MyIP)
	time.Sleep(time.Second)
	if len(os.Args) > 1 {
		ips := strings.Split(os.Args[1], ".")
		if len(ips) != 4 {
			fmt.Println("Invalid IP address format")
			return
		}
		var ip [4]byte
		for i := 0; i < 4; i++ {
			num, err := strconv.Atoi(ips[i])
			if err != nil {
				fmt.Println("Invalid IP address segment:", ips[i])
				return
			}
			ip[i] = byte(num)
		}
		go transactionServices.StartSubscribingTransactionMsg(ip)
		if firstDel.GetHex() == common.GetDelegatedAccount().Hex() {
			go nonceService.StartSubscribingNonceMsg(ip)
		}
		go syncServices.StartSubscribingSyncMsg(ip)
	}
	time.Sleep(time.Second)

	chanPeer := make(chan []byte)
	go tcpip.LookUpForNewPeersToConnect(chanPeer)
	topic := [2]byte{}
	ip := [4]byte{}
QF:
	for {

		select {

		case topicip := <-chanPeer:
			copy(topic[:], topicip[:2])
			copy(ip[:], topicip[2:])

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
