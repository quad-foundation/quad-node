package qtwidgets

import (
	"bytes"
	"fmt"
	"github.com/quad-foundation/quad-node/account"
	"github.com/quad-foundation/quad-node/common"
	clientrpc "github.com/quad-foundation/quad-node/rpc/client"
	"github.com/quad-foundation/quad-node/statistics"
	"github.com/quad-foundation/quad-node/wallet"
	"github.com/therecipe/qt/widgets"
	"strconv"
	"strings"

	"log"
)

var StatsLabel *widgets.QLabel
var MainWallet *wallet.Wallet

func UpdateAccountStats() {
	clientrpc.InRPC <- []byte("STAT")
	var reply []byte
	reply = <-clientrpc.OutRPC
	if bytes.Equal(reply, []byte("Timeout")) {
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
	txt += fmt.Sprintln("Heights max:", st.HeightMax)
	txt += fmt.Sprintln("Time interval [sec.]:", st.TimeInterval)
	txt += fmt.Sprintln("Difficulty:", st.Difficulty)

	txt += fmt.Sprintln("Number of transactions : ", st.Transactions, "/", st.TransactionsPending)
	txt += fmt.Sprintln("Size of transactions [kB] : ", st.TransactionsSize/1024, "/", st.TransactionsPendingSize/1024)
	txt += fmt.Sprintln("TPS:", st.Tps)
	if st.Syncing {
		txt += fmt.Sprintln("Syncing...")
	}
	inb := append([]byte("ACCT"), MainWallet.Address.GetBytes()...)
	clientrpc.InRPC <- inb
	var re []byte
	var acc account.Account

	re = <-clientrpc.OutRPC
	if bytes.Equal(reply, []byte("Timeout")) {
		return
	}
	err = acc.Unmarshal(re)
	if err != nil {
		log.Println("cannot unmarshal account")
		return
	}
	conf := acc.GetBalanceConfirmedFloat()
	uncTx := 0.0

	stake := 0.0
	uncStake := 0.0

	rewards := 0.0
	uncRewards := 0.0

	var stakeAccs [256]account.StakingAccount
	for i := 1; i < 5; i++ { // should be 256
		inb = append([]byte("STAK"), MainWallet.Address.GetBytes()...)
		inb = append(inb, byte(i))
		clientrpc.InRPC <- inb
		re = <-clientrpc.OutRPC
		if string(reply) == "Timeout" {
			return
		}
		err = stakeAccs[i].Unmarshal(re)
		if err != nil {
			log.Println("cannot unmarshal stake account")
			return
		}
		stake += account.Int64toFloat64(stakeAccs[i].StakedBalance)
		rewards += account.Int64toFloat64(stakeAccs[i].StakingRewards)
	}

	txt += fmt.Sprintln("\n\nYour Address:", MainWallet.Address.GetHex())
	txt += fmt.Sprintf("Your holdings: %18.8f QAD\n", conf+stake+rewards+uncTx+uncStake+uncRewards)
	txt += fmt.Sprintf("Confirmed balance: %18.8f QAD\n", conf)
	//txt += fmt.Sprintf("Transactions unconfirmed balance: %18.8f QAD\n", uncTx)
	txt += fmt.Sprintf("Staked amount: %18.8f QAD\n", stake)
	//txt += fmt.Sprintf("Unconfirmed staked amount: %18.8f QAD\n", uncStake)
	txt += fmt.Sprintf("Rewards amount: %18.8f QAD\n", rewards)
	//txt += fmt.Sprintf("Unconfirmed rewards amount: %18.8f QAD\n", uncRewards)
	txt += fmt.Sprintf("\nStaking details:\n")
	for i, acc := range stakeAccs {
		if acc.StakedBalance == 0 && acc.StakingRewards == 0 {
			continue
		}
		a := common.Address{}
		a.Init(acc.DelegatedAccount[:])

		txt += fmt.Sprintf("Delegated Address: %v\n", a.GetHex())
		txt += fmt.Sprintf("Staked: %v = %v\n", i, account.Int64toFloat64(acc.StakedBalance))
		txt += fmt.Sprintf("Rewards: %v = %v\n", i, account.Int64toFloat64(acc.StakingRewards))
	}
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
	AddNewHistoryItem(txt)
	//lastSt = st
}

func ShowAccountPage() *widgets.QTabWidget {
	// create a regular widget
	// give it a QVBoxLayout
	// and make it the central widget of the window
	widget := widgets.NewQTabWidget(nil)
	widget.SetLayout(widgets.NewQVBoxLayout())

	ipLineEdit := widgets.NewQLineEdit(nil)
	ipLineEdit.SetText("192.168.8.104")
	widget.Layout().AddWidget(ipLineEdit)

	miningCheckBox := widgets.NewQCheckBox(nil)
	miningCheckBox.SetText("Start mining")
	widget.Layout().AddWidget(miningCheckBox)
	miningCheckBox.ConnectClicked(func(bool) {
		miningCheckBox.SetEnabled(false)
		var info *string
		v := "No reply"
		info = &v
		defer func(nfo *string) {
			widgets.QMessageBox_Information(nil, "Info", *nfo, widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
		}(info)
		ips := strings.Split(ipLineEdit.Text(), ".")
		if len(ips) != 4 {
			v = "Invalid IP address format"
			return
		}
		var ip [4]byte
		for i := 0; i < 4; i++ {
			num, err := strconv.Atoi(ips[i])
			if err != nil {
				v = fmt.Sprintf("Invalid IP address segment:", ips[i])
				return
			}
			ip[i] = byte(num)
		}
		v = startMining(ip)
		return
	})

	// create a line edit
	// with a custom placeholder text
	// and add it to the central widgets layout
	StatsLabel = widgets.NewQLabel2("Your holdings:", nil, widget.WindowType())
	widget.Layout().AddWidget(StatsLabel)

	return widget
}

func startMining(ip [4]byte) string {
	clientrpc.InRPC <- append([]byte("MINE"), ip[:]...)
	var reply []byte
	reply = <-clientrpc.OutRPC
	if string(reply) == "Timeout" {
		return "Timeout"
	}
	if len(reply) > 0 {
		return string(reply)
	}
	return "No reply"
}
