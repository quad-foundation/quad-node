package qtwidgets

import (
	"fmt"
	"github.com/chainpqc/chainpqc-node/wallet"
	"github.com/therecipe/qt/widgets"
)

var ReserveWallet wallet.Wallet

func ShowWalletPage(w *wallet.Wallet) *widgets.QTabWidget {

	// create a regular widget
	// give it a QVBoxLayout
	// and make it the central widget of the window
	widget := widgets.NewQTabWidget(nil)
	widget.SetLayout(widgets.NewQVBoxLayout())

	// create a line edit
	// with a custom placeholder text
	// and add it to the central widgets layout
	input := widgets.NewQLineEdit(nil)
	input.SetEchoMode(widgets.QLineEdit__Password)
	input.SetPlaceholderText("Password:")
	widget.Layout().AddWidget(input)

	// connect the clicked signal
	// and add it to the central widgets layout
	button := widgets.NewQPushButton2("Load wallet", nil)
	button.ConnectClicked(func(bool) {
		w, err := wallet.Load(0, input.Text())
		if err != nil {
			return
		}

		var info string
		if err != nil {
			info = fmt.Sprintf("%v", err)
		} else {
			info = w.ShowInfo()
			fmt.Println(w.PublicKey.GetHex())
		}
		widgets.QMessageBox_Information(nil, "OK", info, widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
	})
	widget.Layout().AddWidget(button)
	buttonNewWallet := widgets.NewQPushButton2("Generate new wallet", nil)
	buttonNewWallet.ConnectReleased(func() {
		info := "Creating reserve wallet success"

		err := ReserveWallet.Store()

		if err != nil {
			info = fmt.Sprintf("Can not store wallet. Error %v", err)
		} else {
			err := ReserveWallet.ChangePassword("a", input.Text())
			if err != nil {
				info = fmt.Sprintf("Can not change password. Error %v", err)
			} else {
				w, err = wallet.Load(0, input.Text())
				if err != nil {
					return
				}
				if err != nil {
					info = fmt.Sprintf("Can not load wallet. Error %v", err)
				} else {
					info = w.ShowInfo()
				}
			}
		}

		widgets.QMessageBox_Information(nil, "OK", info, widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
		buttonNewWallet.SetDisabled(true)
	})

	widget.Layout().AddWidget(buttonNewWallet)
	newPassword := widgets.NewQLineEdit(nil)
	newPassword.SetEchoMode(widgets.QLineEdit__Password)
	newPassword.SetPlaceholderText("New password:")
	widget.Layout().AddWidget(newPassword)
	repeatPassword := widgets.NewQLineEdit(nil)
	repeatPassword.SetEchoMode(widgets.QLineEdit__Password)
	repeatPassword.SetPlaceholderText("Repeat password:")
	widget.Layout().AddWidget(repeatPassword)
	buttonChangePassword := widgets.NewQPushButton2("Change password", nil)
	buttonChangePassword.ConnectClicked(func(bool) {
		if w.GetSecretKey().GetLength() == 0 {
			widgets.QMessageBox_Information(nil, "Error", "Load wallet first", widgets.QMessageBox__Close, widgets.QMessageBox__Close)
			return
		}
		if newPassword.Text() != repeatPassword.Text() {

			widgets.QMessageBox_Information(nil, "Error", "Passwords do not match", widgets.QMessageBox__Close, widgets.QMessageBox__Close)
			return
		}
		w.SetPassword(newPassword.Text())
		err := w.Store()
		if err != nil {
			widgets.QMessageBox_Information(nil, "Error", fmt.Sprintf("%v", err), widgets.QMessageBox__Close, widgets.QMessageBox__Close)
			return
		}
		widgets.QMessageBox_Information(nil, "OK", "Password changed", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
	})
	widget.Layout().AddWidget(buttonChangePassword)
	buttonMnemonic := widgets.NewQPushButton2("Show mnemonic words", nil)
	buttonMnemonic.ConnectClicked(func(bool) {
		mnemonic, _ := w.GetMnemonicWords()
		widgets.QMessageBox_Information(nil, "OK", fmt.Sprintf("Mnemonic words:\n%v", mnemonic), widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
	})
	widget.Layout().AddWidget(buttonMnemonic)

	inputRestoreMnemonic := widgets.NewQLineEdit(nil)
	inputRestoreMnemonic.SetPlaceholderText("Mnemonic words seperated by space:")
	widget.Layout().AddWidget(inputRestoreMnemonic)
	buttonRestoreMnemonic := widgets.NewQPushButton2("Restore private key from mnemonic words", nil)
	buttonRestoreMnemonic.ConnectClicked(func(bool) {
		err := w.RestoreSecretKeyFromMnemonic(inputRestoreMnemonic.Text())
		if err != nil {
			widgets.QMessageBox_Information(nil, "OK", fmt.Sprintf("Can not restore Private key from mnemonic words:\n%v", err), widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
			return
		}
		sec := w.GetSecretKey()
		widgets.QMessageBox_Information(nil, "OK", fmt.Sprintf("Private Key:\n%v", sec.GetHex()), widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
	})
	widget.Layout().AddWidget(buttonRestoreMnemonic)
	return widget
}
