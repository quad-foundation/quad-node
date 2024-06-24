package qtwidgets

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/quad-foundation/quad-node/account"
	"github.com/quad-foundation/quad-node/common"
	"github.com/quad-foundation/quad-node/core/stateDB"
	clientrpc "github.com/quad-foundation/quad-node/rpc/client"
	"github.com/quad-foundation/quad-node/services/transactionServices"
	"github.com/quad-foundation/quad-node/statistics"
	"github.com/quad-foundation/quad-node/transactionsDefinition"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
	"golang.org/x/exp/rand"
	"log"
	"math"
	"strconv"
	"strings"
)

var AmountLabelData *widgets.QPlainTextEdit
var PoolsSizeLabel *widgets.QPlainTextEdit
var TokenList = map[string]stateDB.TokenInfo{}
var poolTokensButton *widgets.QPushButton
var amountQAD *widgets.QLineEdit
var amountTokens *widgets.QLineEdit
var priceToken *widgets.QLineEdit
var poolPriceToken *widgets.QLineEdit
var humanReadable, humanReadableQAD, price, priceBid, priceAsk float64
var symbol string
var poolCoin, poolToken float64
var removePoolButton *widgets.QRadioButton
var removePoolButtonQAD *widgets.QRadioButton
var buyButton *widgets.QPushButton
var sellButton *widgets.QPushButton
var coinAddr = common.Address{}
var tradeButton *widgets.QRadioButton

//var PriceTokenSeries *charts.QLineSeries

//var ChartView *charts.QChartView

func ShowDexPage() *widgets.QTabWidget {

	// create a regular widget
	// give it a QVBoxLayout
	// and make it the central widget of the window
	widget := widgets.NewQTabWidget(nil)
	widget.SetLayout(widgets.NewQVBoxLayout())

	//ChartView = charts.NewQChartView(nil)
	//
	//ChartView.SetRenderHint(gui.QPainter__Antialiasing, true)
	//widget.Layout().AddWidget(ChartView)

	//PriceTokenSeries = charts.NewQLineSeries(nil)
	//PriceTokenSeries.SetName(fmt.Sprintf("Prices of Token in time"))

	// create a line edit
	// with a custom placeholder text
	// and add it to the central widgets layout
	tokens := widgets.NewQComboBox(nil)
	tokens.ConnectEnterEvent(func(event *core.QEvent) {
		if event.IsAccepted() {
			if !MainWallet.Check() {
				return
			}
			item := tokens.CurrentText()
			if item != "" {
				coin := strings.Split(item, ":")
				if len(coin) < 1 {
					return
				}
				coinAddr.Init(common.Hex2Bytes(coin[0]))
				addr := common.Address{}
				addr.Init(MainWallet.Address.GetBytes())
				GetAllTokensAccountInfo(addr, coinAddr)
				GetAllPoolsInfo()
				//PlotPrices(symbol)
				buyButton.SetText(fmt.Sprintf("BUY %s", symbol))
				sellButton.SetText(fmt.Sprintf("SELL %s", symbol))
			}

		}
	})
	widget.Layout().AddWidget(tokens)
	updateTokensButton := widgets.NewQPushButton2("Update Tokens list", nil)
	updateTokensButton.ConnectClicked(func(bool) {
		ts := GetAllTokens()
		if len(ts) > 0 {
			TokenList = ts
			ls := []string{}
			for addr, ti := range ts {
				ls = append(ls, addr+":"+ti.Name)
			}
			tokens.Clear()
			tokens.AddItems(ls)
			tokens.SetAcceptDrops(true)
		}
	})
	widget.Layout().AddWidget(updateTokensButton)

	addPoolButton := widgets.NewQRadioButton2("Add liquidity", nil)

	widget.Layout().AddWidget(addPoolButton)
	addPoolButton.ConnectClicked(func(bool) {
		poolTokensButton.SetText("Add liquidity to Pool")
		amountTokens.SetEnabled(true)
		amountQAD.SetEnabled(true)
		qad := amountQAD.Text()
		if qad != "" {
			amount := amountTokens.Text()
			if amount != "" {
				g, _ := strconv.ParseFloat(qad, 64)
				t, _ := strconv.ParseFloat(amount, 64)
				if g > 0 {
					price = common.RoundToken(g/t, int(common.Decimals+TokenList[coinAddr.GetHex()].Decimals))
					priceToken.SetText(fmt.Sprintf("My Price QAD/%s = %f", symbol, price))
					if poolCoin > 0 {
						priceBid = common.CalcNewDEXPrice(t, g, poolToken, poolCoin)
						poolPriceToken.SetText(fmt.Sprintf("New pool Price QAD/%s = %f", symbol, priceBid))
						if tradeButton.IsChecked() {
							amountQAD.SetText(fmt.Sprintf("Amount of QAD = %f", t*priceBid))
						}
					}
				}
			}
		}
	})
	removePoolButton = widgets.NewQRadioButton2("Withdraw Token", nil)
	widget.Layout().AddWidget(removePoolButton)
	removePoolButton.ConnectClicked(func(bool) {
		poolTokensButton.SetText(fmt.Sprintf("Withdraw %s from Pool", symbol))
		amountTokens.SetEnabled(true)
		amountQAD.SetEnabled(false)
		amount := amountTokens.Text()
		if amount != "" {
			t, _ := strconv.ParseFloat(amount, 64)
			g := common.RoundCoin(poolCoin / poolToken * t)
			amountQAD.SetText(fmt.Sprintf("%f", g))
			if g > 0 {
				//price = common.RoundToken(g / t, int(common.Decimals + TokenList[coinAddr.GetHex()].Decimals))
				priceToken.SetText(fmt.Sprintf("My Price QAD/%s = %f", symbol, priceAsk))
				if poolCoin > 0 {
					priceBid = common.CalcNewDEXPrice(t, g, poolToken, poolCoin)
					priceAsk = common.CalcNewDEXPrice(-t, -g, poolToken, poolCoin)
					poolPriceToken.SetText(fmt.Sprintf("New pool Price QAD/%s = %f/%f", symbol, priceAsk, priceBid))
					if tradeButton.IsChecked() {
						amountQAD.SetText(fmt.Sprintf("Amount of QAD = %f/%f", t*priceAsk, t*priceBid))
					}
				}
			}
		}
	})

	removePoolButtonQAD = widgets.NewQRadioButton2("Withdraw QAD", nil)
	widget.Layout().AddWidget(removePoolButtonQAD)
	removePoolButtonQAD.ConnectClicked(func(bool) {
		poolTokensButton.SetText("Withdraw QAD from Pool")
		amountTokens.SetEnabled(false)
		amountQAD.SetEnabled(true)
		qad := amountQAD.Text()
		if qad != "" {

			g, _ := strconv.ParseFloat(qad, 64)
			t := common.RoundCoin(poolCoin / poolToken * g)
			amountTokens.SetText(fmt.Sprintf("%f", t))
			if g > 0 {
				priceToken.SetText(fmt.Sprintf("My Price QAD/%s = %f", symbol, priceAsk))
				if poolCoin > 0 {
					priceBid = common.CalcNewDEXPrice(t, g, poolToken, poolCoin)
					priceAsk = common.CalcNewDEXPrice(-t, -g, poolToken, poolCoin)
					poolPriceToken.SetText(fmt.Sprintf("New pool Price QAD/%s = %f/%f", symbol, priceAsk, priceBid))
					if tradeButton.IsChecked() {
						amountQAD.SetText(fmt.Sprintf("Amount of QAD = %f/%f", t*priceAsk, t*priceBid))
					}
				}
			}
		}
	})

	tradeButton = widgets.NewQRadioButton2("BUY/SELL", nil)
	widget.Layout().AddWidget(tradeButton)
	tradeButton.ConnectClicked(func(bool) {
		amountTokens.SetEnabled(true)
		amountQAD.SetEnabled(true)
		amount := amountTokens.Text()
		qad := amountQAD.Text()
		if amount != "" {
			t, _ := strconv.ParseFloat(amount, 64)
			g, _ := strconv.ParseFloat(qad, 64)
			amountQAD.SetText(fmt.Sprintf("%f", g))
			if g > 0 {
				price = common.RoundCoin(g / t)
				priceToken.SetText(fmt.Sprintf("My Price QAD/%s = %f", symbol, price))
				if poolCoin > 0 {
					priceBid = common.CalcNewDEXPrice(t, g, poolToken, poolCoin)
					priceAsk = common.CalcNewDEXPrice(-t, -g, poolToken, poolCoin)
					poolPriceToken.SetText(fmt.Sprintf("New pool Price QAD/%s = %f/%f", symbol, priceAsk, priceBid))
					if tradeButton.IsChecked() {
						amountQAD.SetText(fmt.Sprintf("Amount of QAD = %f/%f", t*priceAsk, t*priceBid))
					}
				}
			}
		}
	})

	addPoolButton.SetChecked(true)

	amountTokens = widgets.NewQLineEdit(nil)
	amountTokens.SetPlaceholderText("Amount of token")
	widget.Layout().AddWidget(amountTokens)
	amountTokens.ConnectTextChanged(func(amount string) {
		if amount != "" {
			QAD := amountQAD.Text()
			if QAD != "" {
				g, _ := strconv.ParseFloat(QAD, 64)
				t, _ := strconv.ParseFloat(amount, 64)
				if g > 0 {
					price = common.RoundCoin(g / t)
					priceToken.SetText(fmt.Sprintf("My Price QAD/%s = %f", symbol, price))
					if poolCoin > 0 {
						priceBid = common.CalcNewDEXPrice(t, g, poolToken, poolCoin)
						priceAsk = common.CalcNewDEXPrice(-t, -g, poolToken, poolCoin)
						poolPriceToken.SetText(fmt.Sprintf("New pool Price QAD/%s = %f/%f", symbol, priceAsk, priceBid))
						if tradeButton.IsChecked() {
							amountQAD.SetText(fmt.Sprintf("Amount of QAD = %f/%f", t*priceAsk, t*priceBid))
						}
					}
				}
			}
		}
	})

	amountQAD = widgets.NewQLineEdit(nil)
	amountQAD.SetPlaceholderText("Amount of QAD")
	widget.Layout().AddWidget(amountQAD)
	amountQAD.ConnectTextChanged(func(QAD string) {
		if QAD != "" {
			amount := amountTokens.Text()
			if amount != "" {
				g, _ := strconv.ParseFloat(QAD, 64)
				t, _ := strconv.ParseFloat(amount, 64)
				if g > 0 {
					price = common.RoundCoin(g / t)
					priceToken.SetText(fmt.Sprintf("My Price QAD/%s = %f", symbol, price))
					if poolCoin > 0 {
						priceBid = common.CalcNewDEXPrice(t, g, poolToken, poolCoin)
						priceAsk = common.CalcNewDEXPrice(-t, -g, poolToken, poolCoin)
						poolPriceToken.SetText(fmt.Sprintf("New pool Price QAD/%s = %f/%f", symbol, priceAsk, priceBid))
						if tradeButton.IsChecked() {
							amountQAD.SetText(fmt.Sprintf("Amount of QAD = %f/%f", t*priceAsk, t*priceBid))
						}
					}
				}
			}
		}
	})
	priceToken = widgets.NewQLineEdit(nil)
	priceToken.SetPlaceholderText("Price of token in QAD")
	priceToken.SetEnabled(false)
	widget.Layout().AddWidget(priceToken)

	poolPriceToken = widgets.NewQLineEdit(nil)
	poolPriceToken.SetPlaceholderText("Price of token you get in QAD")
	widget.Layout().AddWidget(poolPriceToken)

	poolTokensButton = widgets.NewQPushButton2("Add liquidity to Pool", nil)

	poolPriceToken.SetEnabled(false)
	poolTokensButton.SetStyleSheet("background-color : yellow")
	poolTokensButton.ConnectClicked(func(bool) {

		item := tokens.CurrentText()
		if item != "" {
			coin := strings.Split(item, ":")
			if len(coin) < 1 {
				return
			}
			coinAddr := common.Address{}
			coinAddr.Init(common.Hex2Bytes(coin[0]))
			sender := common.Address{}
			sender.Init(MainWallet.Address.GetBytes())
			MakeTransaction(sender, coinAddr)
		}
	})
	widget.Layout().AddWidget(poolTokensButton)

	//groupTradePrice := widgets.NewQGroupBox(nil)
	//layoutPrice := widgets.NewQHBoxLayout()
	//widget.Layout().AddWidget(groupTradePrice)
	//
	//tradePrice := widgets.NewQLineEdit(nil)
	//tradePrice.SetPlaceholderText("Set price")
	//widget.Layout().AddWidget(tradePrice)

	//tradeAmount := widgets.NewQLineEdit(nil)
	//tradeAmount.SetPlaceholderText("Amount of token")
	//widget.Layout().AddWidget(tradeAmount)

	groupTrade := widgets.NewQGroupBox(nil)
	layout := widgets.NewQHBoxLayout()
	widget.Layout().AddWidget(groupTrade)

	buyButton = widgets.NewQPushButton2(fmt.Sprintf("BUY %s", symbol), nil)
	buyButton.SetStyleSheet("background-color : green")
	buyButton.ConnectClicked(func(bool) {

		item := tokens.CurrentText()
		if item != "" {
			coin := strings.Split(item, ":")
			if len(coin) < 1 {
				return
			}
			coinAddr := common.Address{}
			coinAddr.Init(common.Hex2Bytes(coin[0]))
			sender := common.Address{}
			sender.Init(MainWallet.Address.GetBytes())
			MakeTrade(sender, coinAddr, true)
		}

	})

	sellButton = widgets.NewQPushButton2(fmt.Sprintf("SELL %s", symbol), nil)
	sellButton.SetStyleSheet("background-color : red")
	sellButton.ConnectClicked(func(bool) {
		item := tokens.CurrentText()
		if item != "" {
			coin := strings.Split(item, ":")
			if len(coin) < 1 {
				return
			}
			coinAddr := common.Address{}
			coinAddr.Init(common.Hex2Bytes(coin[0]))
			sender := common.Address{}
			sender.Init(MainWallet.Address.GetBytes())
			MakeTrade(sender, coinAddr, false)
		}
	})
	layout.AddWidget(buyButton, 1, 0)
	layout.AddWidget(sellButton, 1, 0)
	groupTrade.SetLayout(layout)

	AmountLabelData = widgets.NewQPlainTextEdit(nil)

	widget.Layout().AddWidget(AmountLabelData)

	PoolsSizeLabel = widgets.NewQPlainTextEdit(nil)
	PoolsSizeLabel.SetLineWidth(1000)
	PoolsSizeLabel.SetWordWrapMode(0)
	widget.Layout().AddWidget(PoolsSizeLabel)

	return widget
}

func GetAllTokens() map[string]stateDB.TokenInfo {
	clientrpc.InRPC <- []byte("LTKN")
	var reply []byte
	reply = <-clientrpc.OutRPC
	if string(reply) == "Timeout" {
		return map[string]stateDB.TokenInfo{}
	}
	ts := map[string]stateDB.TokenInfo{}
	if len(reply) > 0 {
		err := json.Unmarshal(reply, &ts)
		if err != nil {
			log.Println("Can not unmarshal list of tokens", err)
			return map[string]stateDB.TokenInfo{}
		}
		return ts
	}
	return map[string]stateDB.TokenInfo{}
}

func GetBalance(addr common.Address, coin common.Address) int64 {

	m := []byte("GTBL")
	m = append(m, addr.GetBytes()...)
	m = append(m, coin.GetBytes()...)
	clientrpc.InRPC <- m
	var reply []byte
	reply = <-clientrpc.OutRPC
	if string(reply) == "Timeout" {
		return 0
	}
	if len(reply) == 32 {
		ts := common.GetInt64FromSCByteLittleEndian(reply)
		return ts
	}
	return 0
}

func GetAccountDex(coin common.Address) account.DexAccount {

	m := []byte("ADEX")
	m = append(m, coin.GetBytes()...)
	clientrpc.InRPC <- m
	var reply []byte
	reply = <-clientrpc.OutRPC
	if string(reply) == "Timeout" {
		return account.DexAccount{}
	}
	if len(reply) > 8 {
		dexAcc := account.DexAccount{}
		err := common.Unmarshal(reply, common.DexAccountsDBPrefix, &dexAcc)
		if err != nil {
			return account.DexAccount{}
		}

		return dexAcc
	}
	return account.DexAccount{}
}

func GetAccount(a common.Address) (account.Account, error) {
	inb := append([]byte("ACCT"), a.GetBytes()...)
	clientrpc.InRPC <- inb
	var re, reply []byte
	var acc account.Account

	re = <-clientrpc.OutRPC
	if string(reply) == "Timeout" {
		return account.Account{}, fmt.Errorf("timeout")
	}
	err = acc.Unmarshal(re)
	if err != nil {
		log.Println("cannot unmarshal account")
		return account.Account{}, err
	}
	return acc, nil
}

func GetAllPoolsInfo() string {
	txt := "Pools sizes:\n"
	ti := GetAllTokens()
	for addr, info := range ti {
		a := common.Address{}
		a.Init(common.Hex2Bytes(addr))
		accDex := GetAccountDex(a)

		tokenPoolFloat := account.Int64toFloat64ByDecimals(accDex.TokenPool, info.Decimals)
		coinPoolFloat := account.Int64toFloat64ByDecimals(accDex.CoinPool, common.Decimals)
		symb := strings.Trim(info.Symbols, string(byte(0)))
		tmptxt := fmt.Sprintln(addr, " = ", tokenPoolFloat, " ", symb)
		txt += tmptxt
		txt += "Users provided liquidity into pool:\n"

		if tokenPoolFloat > 0 {
			price = common.RoundCoin(coinPoolFloat / tokenPoolFloat)
		}
		tmptxt = fmt.Sprintf("Pool price QAD/%s = %f", symb, price)
		txt += tmptxt
		if bytes.Compare(a.GetBytes(), coinAddr.GetBytes()) == 0 {
			poolCoin = coinPoolFloat
			poolToken = tokenPoolFloat
		}

		//price for chart
		//PriceTokenSeries.Append(float64(time.Now().UTC().Unix()), price)
	}

	//log.Println(PriceTokenSeries)
	PoolsSizeLabel.SetPlainText(txt)
	return txt
}

func GetAllTokensAccountInfo(a common.Address, symbolAddr common.Address) string {
	txt := "My Address:\n" + a.GetHex() + "\n\nMy Holding:\n"
	myacc, _ := GetAccount(a)
	myBal := myacc.GetBalanceConfirmedFloat()
	txt += fmt.Sprintln(myBal, " QAD\n\nTokens:")
	ti := GetAllTokens()
	for addr, info := range ti {

		coinAddr := common.Address{}
		coinAddr.Init(common.Hex2Bytes(addr[:]))
		if bytes.Compare(coinAddr.GetBytes(), symbolAddr.GetBytes()) == 0 {
			symbol = strings.Trim(info.Symbols, string(byte(0)))
		}
		balCoin := GetBalance(a, coinAddr)
		humanReadable := account.Int64toFloat64ByDecimals(balCoin, info.Decimals)
		symb := strings.Trim(info.Symbols, string(byte(0)))
		tmptxt := fmt.Sprintln(addr, " = ", humanReadable, " ", symb)
		txt += tmptxt
	}
	txt += "\nMy Tokens in Pools:\n"

	for addr, info := range ti {
		a := common.Address{}
		a.Init(common.Hex2Bytes(addr))
		accDex := GetAccountDex(a)
		coinAddr := common.Address{}
		coinAddr.Init(common.Hex2Bytes(addr[:]))
		symb := strings.Trim(info.Symbols, string(byte(0)))

		if bal, ok := accDex.Balances[a.ByteValue]; ok {
			humanReadable = account.Int64toFloat64ByDecimals(bal.TokenBalance, info.Decimals)

			tmptxt := fmt.Sprintln(addr, " = ", humanReadable, " ", symb)
			txt += tmptxt
		}
		if bal, ok := accDex.Balances[a.ByteValue]; ok {
			humanReadableQAD = account.Int64toFloat64ByDecimals(bal.CoinBalance, common.Decimals)
			tmptxt := fmt.Sprintln(addr, " = ", humanReadableQAD, " QAD")
			txt += tmptxt
		}
		//if humanReadableQAD > 0 {
		//	price = common.RoundQAD(humanReadableQAD / humanReadable)
		//}
		//tmptxt := fmt.Sprintf("My price QAD/%s = %f", symb, price)
		//txt += tmptxt
	}
	AmountLabelData.SetPlainText(txt)
	return txt
}

func MakeTransaction(sender, coinAddr common.Address) {
	balance := GetBalance(sender, coinAddr)
	//myAcc, _ := GetAccount(sender)
	ti, ok := TokenList[coinAddr.GetHex()]
	if ok {
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
		af, err := strconv.ParseFloat(amountTokens.Text(), 64)
		if err != nil {
			v = fmt.Sprint(err)
			info = &v
			return
		}
		am := int64(af * math.Pow10(int(ti.Decimals)))
		if float64(am) != af*math.Pow10(int(ti.Decimals)) {
			v = fmt.Sprint("Precision needs to be not larger than", ti.Decimals, " digits")
			info = &v
			return
		}

		QAD, err := strconv.ParseFloat(amountQAD.Text(), 64)
		if err != nil {
			v = fmt.Sprint(err)
			info = &v
			return
		}
		QADam := int64(QAD * math.Pow10(int(common.Decimals)))
		if float64(QADam) != QAD*math.Pow10(int(common.Decimals)) {
			v = fmt.Sprint("Precision needs to be not larger than", common.Decimals, " digits")
			info = &v
			return
		}
		operation := 2
		if removePoolButton.IsChecked() {
			am *= -1
			af *= -1
			QADam = 0
			operation = 5
		}
		if removePoolButtonQAD.IsChecked() {
			QADam *= -1
			am = 0
			af = 0
			operation = 6
		}
		//if -QADam > myAcc.GetBalance() {
		//	v = fmt.Sprint("Not enough QAD balance at account")
		//	info = &v
		//	return
		//}
		if af > float64(balance) {
			v = fmt.Sprint("Not enough balance at account")
			info = &v
			return
		}
		ar := common.GetDelegatedAccountAddress(int16(512 + operation))
		txd := transactionsDefinition.TxData{
			Recipient: ar,
			Amount:    QADam,
			OptData:   common.GetByteInt64(am),
			Pubkey:    common.PubKey{},
		}

		par := transactionsDefinition.TxParam{
			ChainID:     ChainID,
			Sender:      sender,
			SendingTime: common.GetCurrentTimeStampInSecond(),
			Nonce:       int16(rand.Intn(0xffff)),
		}
		tx := transactionsDefinition.Transaction{
			TxData:          txd,
			TxParam:         par,
			Hash:            common.Hash{},
			Signature:       common.Signature{},
			Height:          0,
			GasPrice:        0,
			GasUsage:        0,
			ContractAddress: coinAddr,
			OutputLogs:      nil,
		}
		clientrpc.InRPC <- []byte("STAT")
		var reply []byte
		reply = <-clientrpc.OutRPC
		st := &statistics.MainStats{}
		err = common.Unmarshal(reply, common.StatDBPrefix, st)
		if err != nil {
			v = fmt.Sprint("Can not unmarshal statistics", err)
			info = &v
			return
		}
		tx.Height = st.Heights
		tx.GasUsage = tx.GasUsageEstimate()
		err = tx.CalcHashAndSet()
		if err != nil {
			v = fmt.Sprint("Can not calculate hash", err)
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

	}
}

func MakeTrade(sender, coinAddr common.Address, isBuy bool) {
	balance := GetBalance(sender, coinAddr)
	ti, ok := TokenList[coinAddr.GetHex()]
	if ok {
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
		af, err := strconv.ParseFloat(amountTokens.Text(), 64)
		if err != nil {
			v = fmt.Sprint(err)
			info = &v
			return
		}
		am := int64(af * math.Pow10(int(ti.Decimals)))
		if float64(am) != af*math.Pow10(int(ti.Decimals)) {
			v = fmt.Sprint("Precision needs to be not larger than", ti.Decimals, " digits")
			info = &v
			return
		}
		var operation int
		operation = 3
		if !isBuy {
			am *= -1
			af *= -1
			operation = 4
		}
		if -af > float64(balance) {
			v = fmt.Sprint("Not enough balance at account")
			info = &v
			return
		}
		ar := common.GetDelegatedAccountAddress(int16(512 + operation))
		txd := transactionsDefinition.TxData{
			Recipient: ar,
			Amount:    am,
			OptData:   common.GetByteInt64(am),
			Pubkey:    common.PubKey{},
		}

		par := transactionsDefinition.TxParam{
			ChainID:     ChainID,
			Sender:      sender,
			SendingTime: common.GetCurrentTimeStampInSecond(),
			Nonce:       int16(rand.Intn(0xffff)),
		}
		tx := transactionsDefinition.Transaction{
			TxData:          txd,
			TxParam:         par,
			Hash:            common.Hash{},
			Signature:       common.Signature{},
			Height:          0,
			GasPrice:        0,
			GasUsage:        0,
			ContractAddress: coinAddr,
			OutputLogs:      nil,
		}
		clientrpc.InRPC <- []byte("STAT")
		var reply []byte
		reply = <-clientrpc.OutRPC
		st := &statistics.MainStats{}
		err = common.Unmarshal(reply, common.StatDBPrefix, st)
		if err != nil {
			v = fmt.Sprint("Can not unmarshal statistics", err)
			info = &v
			return
		}
		tx.Height = st.Heights
		tx.GasUsage = tx.GasUsageEstimate()
		err = tx.CalcHashAndSet()
		if err != nil {
			v = fmt.Sprint("Can not calculate hash", err)
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

	}
}

//
//func PlotPrices(symbol string) *charts.QChart {
//
//	chart := charts.NewQChart(nil, 0)
//	chart.AddSeries(PriceTokenSeries)
//	chart.SetTitle(fmt.Sprintf("Prices of QAD/%s from DEX", symbol))
//	chart.SetAnimationOptions(charts.QChart__SeriesAnimations)
//
//	chart.CreateDefaultAxes()
//	//
//	//axisX := chart.Axes(core.Qt__Horizontal, PriceTokenSeries)[0]
//	//
//	//
//	//axisY := chart.Axes(core.Qt__Vertical, PriceTokenSeries)[0]
//	//axisY.SetMax(axisY.Max() * 1.01)
//	//axisY.SetMin(axisY.Min() * 0.99);
//
//	chart.Legend().SetVisible(true)
//	chart.Legend().SetAlignment(core.Qt__AlignBottom)
//	ChartView.SetChart(chart)
//	return chart
//
//}
