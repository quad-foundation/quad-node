package wallet

import (
	"github.com/chainpqc/chainpqc-node/common"
	"github.com/chainpqc/chainpqc-node/crypto/oqs"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

func init() {
	mainWallet = EmptyWallet()
	var err error
	HomePath, err = os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	HomePath += "db/wallet"
	mainWallet.SetPassword("a")
	err = mainWallet.Load()
	if err != nil {
		return
	}
}

func TestEmptyWallet(t *testing.T) {
	w := EmptyWallet()
	assert.Equal(t, "", w.password)
	assert.Nil(t, w.passwordBytes)
	assert.Nil(t, w.Iv)
	assert.Equal(t, common.PrivKey{}, w.secretKey)
	assert.Equal(t, common.PubKey{}, w.PublicKey)
	assert.Equal(t, common.Address{}, w.Address)
	assert.Equal(t, oqs.Signature{}, w.signer)
}
func TestSetPassword(t *testing.T) {
	w := EmptyWallet()
	password := "testpassword"
	w.SetPassword(password)
	assert.Equal(t, password, w.password)
	assert.NotNil(t, w.passwordBytes)
}
func TestPasswordToByte(t *testing.T) {
	password := "testpassword"
	passwordBytes := passwordToByte(password)
	assert.NotNil(t, passwordBytes)
	assert.Equal(t, 32, len(passwordBytes))
}
func TestGenerateNewWallet(t *testing.T) {
	password := "testpassword"
	w, err := GenerateNewWallet(password)
	assert.NoError(t, err)
	assert.NotNil(t, w)
	assert.Equal(t, password, w.password)
	assert.NotNil(t, w.passwordBytes)
	assert.NotNil(t, w.Iv)
	assert.NotNil(t, w.secretKey)
	assert.NotNil(t, w.PublicKey)
	assert.NotNil(t, w.Address)
	assert.NotNil(t, w.signer)
}
func TestStoreAndLoadWallet(t *testing.T) {
	// Generate a new wallet
	password := "testpassword"
	wallet, err := GenerateNewWallet(password)
	assert.NoError(t, err)
	// Store the wallet
	err = wallet.Store()
	assert.NoError(t, err)
	// Load the wallet
	loadedWallet := EmptyWallet()
	loadedWallet.SetPassword(password)
	err = loadedWallet.Load()
	assert.NoError(t, err)
	// Check if the loaded wallet is the same as the original wallet
	assert.Equal(t, wallet.PublicKey, loadedWallet.PublicKey)
	assert.Equal(t, wallet.Address, loadedWallet.Address)
	assert.Equal(t, wallet.secretKey, loadedWallet.secretKey)
}
func TestChangePassword(t *testing.T) {
	// Generate a new wallet
	password := "testpassword"
	newPassword := "newtestpassword"
	wallet, err := GenerateNewWallet(password)
	assert.NoError(t, err)
	// Change the password
	err = wallet.ChangePassword(password, newPassword)
	assert.NoError(t, err)
	// Load the wallet with the new password
	loadedWallet := EmptyWallet()
	loadedWallet.SetPassword(newPassword)
	err = loadedWallet.Load()
	assert.NoError(t, err)
	// Check if the loaded wallet is the same as the original wallet
	assert.Equal(t, wallet.PublicKey, loadedWallet.PublicKey)
	assert.Equal(t, wallet.Address, loadedWallet.Address)
	assert.Equal(t, wallet.secretKey, loadedWallet.secretKey)
}
func TestSignAndVerify(t *testing.T) {
	// Generate a new wallet
	password := "testpassword"
	wallet, err := GenerateNewWallet(password)
	assert.NoError(t, err)
	// Sign a message using the wallet
	message := []byte("Hello, world!")
	signature, err := wallet.Sign(message)
	if err != nil {
		log.Fatal(err)
	}
	// Verify the signature using the wallet's public key
	isVerified := Verify(message, signature.GetBytes(), wallet.PublicKey.GetBytes())
	assert.Equal(t, isVerified, true)
}
