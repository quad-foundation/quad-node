package qtwidgets

import (
	"fmt"
	"github.com/quad/quad-node/common"
	clientrpc "github.com/quad/quad-node/rpc/client"
	"github.com/quad/quad-node/statistics"
	"github.com/quad/quad-node/wallet"
	"github.com/therecipe/qt/widgets"

	"log"
)

var StatsLabel *widgets.QLabel
var MainWalllet *wallet.Wallet

func UpdateAccountStats() {
	clientrpc.InRPC <- []byte("STAT")
	var reply []byte
	reply = <-clientrpc.OutRPC
	if string(reply) == "Timeout" {
		return
	}
	st := &statistics.MainStats{}
	err := common.Unmarshal(reply, common.StatDBPrefix, st)
	if err != nil {
		log.Println("Can not unmarshal statistics")
		return
	}
	if st.Heights == 0 {
		return
	}
	txt := fmt.Sprintln("Height:", st.Heights)
	txt += fmt.Sprintln("Chain:", st.Chain, " : ", common.NamesOfChains[st.Chain])
	txt += fmt.Sprintln("Heights max:", st.HeightMax)
	txt += fmt.Sprintln("Time interval [sec.]:", st.TimeInterval)
	txt += fmt.Sprintln("Difficulty:", st.Difficulty)

	for i := uint8(0); i < 5; i++ {
		txt += fmt.Sprintln("Number of transactions in ", common.NamesOfChains[i], " : ", st.Transactions[i], "/", st.TransactionsPending[i])
	}
	for i := uint8(0); i < 5; i++ {
		txt += fmt.Sprintln("Size of transactions [kB] in ", common.NamesOfChains[i], " : ", st.TransactionsSize[i]/1024, "/", st.TransactionsPendingSize[i]/1024)
	}
	txt += fmt.Sprintln("TPS:", st.Tps)
	if st.Syncing {
		txt += fmt.Sprintln("Syncing...")
	}
	//inb := append([]byte("ACCT"), w.Address.GetBytes()...)
	//clientrpc.InRPC <- inb
	//var re []byte
	//var acc account.Account
	//
	//re = <-clientrpc.OutRPC
	//if string(reply) == "Timeout" {
	//	return
	//}
	//err = json.Unmarshal(re, &acc)
	//if err != nil {
	//	log.Println("cannot unmarshal account")
	//	return
	//}
	//conf := acc.GetBalanceConfirmedFloat(st.Heights)
	//uncTx := acc.GetUnconfirmedTransactionFloat(st.Heights)
	//
	//stake := acc.GetStakeConfirmedFloat(st.Heights)
	//uncStake := acc.GetUnconfirmedStakeFloat(st.Heights)
	//
	//txt += fmt.Sprintln("\n\nYour Address:", w.Address.GetHex())
	//txt += fmt.Sprintf("Your holdings: %18.8f GXY\n", conf+uncTx+stake+uncStake)
	//txt += fmt.Sprintf("Confirmed balance: %18.8f GXY\n", conf)
	//txt += fmt.Sprintf("Transactions unconfirmed balance: %18.8f GXY\n", uncTx)
	//txt += fmt.Sprintf("Staked amount: %18.8f GXY\n", stake)
	//txt += fmt.Sprintf("Unconfirmed staked amount: %18.8f GXY\n", uncStake)
	//txt += fmt.Sprintf("\nStaking details:\n")
	//for k, v := range acc.StakingAddresses {
	//	if v == 0 {
	//		continue
	//	}
	//	ab, _ := hex.DecodeString(k)
	//	a := common.Address{}
	//	a.Init(ab[:])
	//	if n := common.NumericalDelegatedAccountAddress(a); n > 0 {
	//
	//		txt += fmt.Sprintf("Delegated Address: %v\n", a.GetHex())
	//		txt += fmt.Sprintf("Delegated Account Number: %v = %v\n", n, account.Int64toFloat64(v))
	//
	//	}
	//}
	//fmt.Println(txt)
	StatsLabel.SetText(txt)
	//txt2 := ""
	//if lastSt.Heights == 0 {
	//	lastSt = st
	//	lastSt.Heights = 1
	//}
	//histState := acc.GetHistoryState(lastSt.Heights, st.Heights)
	//histRewards := acc.GetHistoryUnconfirmedRewards(lastSt.Heights, st.Heights)
	//histConfirmed := acc.GetHistoryConfirmedTransaction(lastSt.Heights, st.Heights)
	//
	//histStake := acc.GetHistoryStake(lastSt.Heights, st.Heights)
	//for i := st.Heights - 1; i >= lastSt.Heights; i-- {
	//	txt2 += fmt.Sprintln(i, "/", st.Heights, ":")
	//	txt2 += fmt.Sprintln("Balance:", account.Int64toFloat64(histState[i]))
	//	txt2 += fmt.Sprintln("Staked:", account.Int64toFloat64(histStake[i]))
	//	txt2 += fmt.Sprintln("Unconfirmed reward:", account.Int64toFloat64(histRewards[i]))
	//
	//	for k, v := range histConfirmed[i] {
	//		if v != 0 {
	//			txt2 += fmt.Sprintln("Confirmed", k, account.Int64toFloat64(v))
	//		}
	//	}
	//}
	//AddNewHistoryItem(txt)
	//lastSt = st
}

func ShowAccountPage() *widgets.QTabWidget {
	// create a regular widget
	// give it a QVBoxLayout
	// and make it the central widget of the window
	widget := widgets.NewQTabWidget(nil)
	widget.SetLayout(widgets.NewQVBoxLayout())

	// create a line edit
	// with a custom placeholder text
	// and add it to the central widgets layout
	StatsLabel = widgets.NewQLabel2("Your holdings:", nil, widget.WindowType())
	widget.Layout().AddWidget(StatsLabel)

	// connect the clicked signal
	// and add it to the central widgets layout
	//buttonMining := widgets.NewQPushButton2("Show balances", nil)
	//buttonMining.ConnectClicked(func(bool) {
	//	clientrpc.InRPC <- []byte("STAT")
	//	var reply []byte
	//	reply = <-clientrpc.OutRPC
	//	st := statistics.MainStats{}
	//	err := json.Unmarshal(reply, &st)
	//	if err != nil {
	//		log.Println("Can not unmarshal statistics")
	//		return
	//	}
	//inb := []byte("ACCS")
	//clientrpc.InRPC <- inb
	//var re []byte
	//var accs [][]byte
	//
	//re = <-clientrpc.OutRPC
	//err = json.Unmarshal(re, &accs)
	//if err != nil {
	//	log.Println("cannot unmarshal account")
	//	return
	//}
	//txt := ""
	//for _, a := range accs {
	//	inb := append([]byte("ACCT"), a...)
	//	clientrpc.InRPC <- inb
	//	var re []byte
	//	var acc account.Account
	//
	//	re = <-clientrpc.OutRPC
	//	err = json.Unmarshal(re, &acc)
	//	if err != nil {
	//		log.Println("cannot unmarshal account")
	//		return
	//	}
	//	conf := acc.GetBalanceConfirmedFloat(st.Heights)
	//	uncTx := acc.GetUnconfirmedTransactionFloat(st.Heights)
	//	stake := acc.GetStakeConfirmedFloat(st.Heights)
	//	uncStake := acc.GetUnconfirmedStakeFloat(st.Heights)
	//
	//	txt += fmt.Sprintln("Address:", hex.EncodeToString(a))
	//	txt += fmt.Sprintln("Holdings:", conf+uncTx, "GXY")
	//	txt += fmt.Sprintln("Confirmed balance:", conf, "GXY")
	//	txt += fmt.Sprintln("Transactions unconfirmed balance:", uncTx, "GXY")
	//	txt += fmt.Sprintf("Staked amount: %18.8f GXY\n", stake)
	//	txt += fmt.Sprintf("Unconfirmed staked amount: %18.8f GXY\n", uncStake)
	//}
	//widgets.QMessageBox_Information(nil, "Balances", txt, widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
	//})
	//widget.Layout().AddWidget(buttonMining)
	return widget
}
