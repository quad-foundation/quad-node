package qtwidgets

import (
	"encoding/hex"
	"fmt"
	"github.com/quad-foundation/quad-node/common"
	clientrpc "github.com/quad-foundation/quad-node/rpc/client"
	"github.com/quad-foundation/quad-node/services/transactionServices"
	"github.com/quad-foundation/quad-node/statistics"
	"github.com/quad-foundation/quad-node/transactionsDefinition"
	"github.com/therecipe/qt/widgets"
	"golang.org/x/exp/rand"
	"math"
	"strconv"
)

var ChainID = int16(23)
var SmartContractData *widgets.QTextEdit
var Recipient *widgets.QLineEdit
var Amount *widgets.QLineEdit

func ShowSendPage() *widgets.QTabWidget {

	// create a regular widget
	// give it a QVBoxLayout
	// and make it the central widget of the window
	widget := widgets.NewQTabWidget(nil)
	widget.SetLayout(widgets.NewQVBoxLayout())

	// create a line edit
	// with a custom placeholder text
	// and add it to the central widgets layout
	Recipient = widgets.NewQLineEdit(nil)
	Recipient.SetPlaceholderText("Address")
	widget.Layout().AddWidget(Recipient)

	Amount = widgets.NewQLineEdit(nil)
	Amount.SetPlaceholderText("Amount")
	widget.Layout().AddWidget(Amount)

	SmartContractData = widgets.NewQTextEdit(nil)
	SmartContractData.SetPlaceholderText("Smart Contract Data")
	widget.Layout().AddWidget(SmartContractData)

	pubkeyInclude := widgets.NewQCheckBox(nil)
	pubkeyInclude.SetText("Public key include in transaction")
	widget.Layout().AddWidget(pubkeyInclude)

	// connect the clicked signal
	// and add it to the central widgets layout
	button := widgets.NewQPushButton2("Send", nil)
	button.ConnectClicked(func(bool) {
		var info *string
		v := "Transaction sent"
		info = &v
		defer func(nfo *string) {
			widgets.QMessageBox_Information(nil, "Info", *nfo, widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
		}(info)

		if !MainWallet.Check() {
			v = fmt.Sprint("Load wallet first")
			info = &v
			return
		}
		ar := common.Address{}
		if len(Recipient.Text()) < 20 {
			i, err := strconv.Atoi(Recipient.Text())
			if err != nil || i > 255 {
				v = fmt.Sprint(err)
				info = &v
				return
			}
			ar = common.GetDelegatedAccountAddress(int16(i))
		} else {

			bar, err := hex.DecodeString(Recipient.Text())
			if err != nil {
				v = fmt.Sprint(err)
				info = &v
				return
			}
			err = ar.Init(bar)
			if err != nil {
				v = fmt.Sprint(err)
				info = &v
				return
			}
		}
		af, err := strconv.ParseFloat(Amount.Text(), 64)
		if err != nil {
			v = fmt.Sprint(err)
			info = &v
			return
		}
		if af < 0 {
			v = fmt.Sprint("Send Amount cannot be negative")
			info = &v
			return
		}
		am := int64(af * math.Pow10(int(common.Decimals)))
		if float64(am) != af*math.Pow10(int(common.Decimals)) {
			v = fmt.Sprint("Precision needs to be not larger than", common.Decimals, " digits")
			info = &v
			return
		}
		optData := SmartContractData.ToPlainText()

		scData := []byte{}
		if len(optData) > 0 {
			scData, err = hex.DecodeString(optData)
			if err != nil {
				scData = []byte{}
			}
		}

		pk := common.PubKey{}
		if pubkeyInclude.IsChecked() {
			pk = MainWallet.PublicKey
		}
		txd := transactionsDefinition.TxData{
			Recipient: ar,
			Amount:    am,
			OptData:   scData,
			Pubkey:    pk,
		}
		par := transactionsDefinition.TxParam{
			ChainID:     ChainID,
			Sender:      MainWallet.Address,
			SendingTime: common.GetCurrentTimeStampInSecond(),
			Nonce:       int16(rand.Intn(0xffff)),
		}
		tx := transactionsDefinition.Transaction{
			TxData:    txd,
			TxParam:   par,
			Hash:      common.Hash{},
			Signature: common.Signature{},
			Height:    0,
			GasPrice:  int64(rand.Intn(0x0000000f)),
			GasUsage:  0,
		}
		clientrpc.InRPC <- []byte("STAT")
		var reply []byte
		reply = <-clientrpc.OutRPC
		st := statistics.MainStats{}
		err = common.Unmarshal(reply, common.StatDBPrefix, &st)
		if err != nil {
			v = fmt.Sprint("Can not unmarshal statistics: ", err)
			info = &v
			return
		}
		tx.GasUsage = tx.GasUsageEstimate()
		tx.Height = st.Heights
		err = tx.CalcHashAndSet()
		if err != nil {
			v = fmt.Sprint("can not generate hash transaction: ", err)
			info = &v
			return
		}
		err = tx.Sign(MainWallet)
		if err != nil {
			v = fmt.Sprint(err)
			info = &v
			return
		}
		msg, err := transactionServices.GenerateTransactionMsg([]transactionsDefinition.Transaction{tx}, []byte("tx"), [2]byte{'T', 'T'})
		if err != nil {
			v = fmt.Sprint(err)
			info = &v
			return
		}
		tmm := msg.GetBytes()
		clientrpc.InRPC <- append([]byte("TRAN"), tmm...)
		<-clientrpc.OutRPC
		v = string(reply)
		info = &v
	})
	widget.Layout().AddWidget(button)

	return widget
}
