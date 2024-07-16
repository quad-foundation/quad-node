package main

import (
	"github.com/quad-foundation/quad-node/cmd/gui/qtwidgets"
	"github.com/quad-foundation/quad-node/common"
	clientrpc "github.com/quad-foundation/quad-node/rpc/client"
	"github.com/quad-foundation/quad-node/statistics"
	"github.com/quad-foundation/quad-node/tcpip"
	"github.com/quad-foundation/quad-node/wallet"
	"github.com/therecipe/qt/widgets"
	"net"
	"os"
	"strconv"
	"time"
)

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
	ip_this := tcpip.GetInternalIp()
	ip_str := net.IPv4(ip_this[0], ip_this[1], ip_this[2], ip_this[3])
	// create a window
	window := widgets.NewQTabWidget(nil)
	window.SetMinimumSize2(800, 900)
	window.SetWindowTitle("QUAD Wallet - " + ip_str.String() +
		" Node Account: " +
		strconv.Itoa(int(common.NumericalDelegatedAccountAddress(common.GetDelegatedAccount()))))

	qtwidgets.MainWallet = wallet.EmptyWallet(0)
	//var reply []byte
	//
	//for {
	//	clientrpc.InRPC <- []byte("WALL")
	//	reply = <-clientrpc.OutRPC
	//	if string(reply) != "Timeout" {
	//		break
	//	}
	//	time.Sleep(time.Second)
	//}
	//err := json.Unmarshal(reply, MainWallet)
	//if err != nil {
	//	log.Println("Can not unmarshal wallet")
	//}
	walletWidget := qtwidgets.ShowWalletPage()
	accountWidget := qtwidgets.ShowAccountPage()
	sendWidget := qtwidgets.ShowSendPage()
	historyWidget := qtwidgets.ShowHistoryPage()
	detailsWidget := qtwidgets.ShowDetailsPage()
	stakingWidget := qtwidgets.ShowStakingPage()
	smartContractWidget := qtwidgets.ShowSmartContractPage()
	dexWidget := qtwidgets.ShowDexPage()
	window.AddTab(walletWidget, "Wallet")
	window.AddTab(accountWidget, "Account")
	window.AddTab(sendWidget, "Send/Register")
	window.AddTab(stakingWidget, "Staking/Rewards")
	window.AddTab(historyWidget, "Transactions history")
	window.AddTab(detailsWidget, "Details")
	window.AddTab(smartContractWidget, "Smart Contract")
	window.AddTab(dexWidget, "DEX")
	// make the window visible
	window.Show()

	go func() {
		for range time.Tick(time.Second * 3) {
			qtwidgets.UpdateAccountStats()
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
