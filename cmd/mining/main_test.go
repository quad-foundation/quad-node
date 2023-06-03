package main

import (
	"github.com/wonabru/galaxy/block"
	"github.com/wonabru/galaxy/common"
	"github.com/wonabru/galaxy/genesis"
	"github.com/wonabru/galaxy/nonceMsg"
	"github.com/wonabru/galaxy/nonceMsg/syncing"
	serverrpc "github.com/wonabru/galaxy/rpc/server"
	"github.com/wonabru/galaxy/tcpip"
	"github.com/wonabru/galaxy/transaction/broadcast"
	wallet "github.com/wonabru/galaxy/walletDB"
	"runtime"
	"testing"
)

func TestName(t *testing.T) {
	runtime.SetCPUProfileRate(500)
	wallet.MainWallet.SetPassword("a")
	wallet.MainWallet.Load()
	genesis.InitGenesis()

	firstDel := common.GetDelegatedAccountAddress(1)
	if firstDel.GetHex() == common.DelegatedAccount.GetHex() {

		common.IsSyncing.Store(false)
	}

	go broadcast.InitTxService(0)
	go broadcast.InitTxService(1)
	go nonceMsg.InitNonceService()
	go syncing.InitSyncService()

	go serverrpc.ListenRPC()

	go nonceMsg.StartSubscribingNonceMsg(tcpip.MyIP)
	go broadcast.StartSubscribingTransaction(tcpip.MyIP, 0)
	go broadcast.StartSubscribingTransaction(tcpip.MyIP, 1)

	tx := block.RegisterMyPubKey(common.GetCurrentTimeStampInSecond())
	go sendTransactionSideChain(tx)

	for {
		lrh, _ := block.GetHeight()
		if lrh > 50 {
			break
		}
	}

	//<-tcpip.Quit

}
