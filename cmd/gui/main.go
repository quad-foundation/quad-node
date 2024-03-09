package main

import (
	"encoding/json"
	"github.com/quad/quad-node/cmd/gui/qtwidgets"
	"github.com/quad/quad-node/common"
	clientrpc "github.com/quad/quad-node/rpc/client"
	"github.com/quad/quad-node/statistics"
	"github.com/quad/quad-node/wallet"
	"github.com/therecipe/qt/widgets"
	"log"
	"os"
	"strconv"
	"time"
)

var MainWalllet *wallet.Wallet

func main() {
	var ip string
	if len(os.Args) > 1 {
		ip = os.Args[1]
	} else {
		ip = "127.0.0.1"
	}
	statistics.InitGlobalMainStats()
	defer statistics.DestroyGlobalMainStats()
	go clientrpc.ConnectRPC(ip)
	time.Sleep(time.Second)

	// needs to be called once before you can start using the QWidgets
	app := widgets.NewQApplication(len(os.Args), os.Args)

	// create a window
	window := widgets.NewQTabWidget(nil)
	window.SetMinimumSize2(800, 900)
	window.SetWindowTitle("QUAD Wallet - " + ip +
		" Node Account: " +
		strconv.Itoa(int(common.NumericalDelegatedAccountAddress(common.GetDelegatedAccount()))))

	MainWalllet = wallet.EmptyWallet(0)
	var reply []byte

	for {
		clientrpc.InRPC <- []byte("WALL")
		reply = <-clientrpc.OutRPC
		if string(reply) != "Timeout" {
			break
		}
		time.Sleep(time.Second)
	}
	err := json.Unmarshal(reply, MainWalllet)
	if err != nil {
		log.Println("Can not unmarshal wallet")
	}
	walletWidget := qtwidgets.ShowWalletPage(MainWalllet)
	accountWidget := qtwidgets.ShowAccountPage()
	//sendWidget := qtwidgets.ShowSendPage(&MainWalllet)
	//historyWidget := qtwidgets.ShowHistoryPage()
	//detailsWidget := qtwidgets.ShowDetailsPage()
	//stakingWidget := qtwidgets.ShowStakingPage(&MainWalllet)
	//smartContractWidget := qtwidgets.ShowSmartContractPage()
	//dexWidget := qtwidgets.ShowDexPage(&MainWalllet)
	window.AddTab(walletWidget, "Wallet")
	window.AddTab(accountWidget, "Account")
	//window.AddTab(sendWidget, "Send/Register")
	//window.AddTab(stakingWidget, "Staking/Unstaking")
	//window.AddTab(historyWidget, "Transactions history")
	//window.AddTab(detailsWidget, "Details")
	//window.AddTab(smartContractWidget, "Smart Contract")
	//window.AddTab(dexWidget, "DEX")
	// make the window visible
	window.Show()

	go func() {
		for range time.Tick(time.Second * 3) {
			qtwidgets.UpdateAccountStats(MainWalllet)
		}
	}()
	//go func() {
	//	for range time.Tick(time.Second) {
	//		qtwidgets2.GetAllPoolsInfo()
	//	}
	//}()

	// start the main Qt event loop
	// and block until app.Exit() is called
	// or the window is closed by the user
	app.Exec()
}
