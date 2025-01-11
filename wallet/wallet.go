package wallet

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/wonabru/bip39"

	"github.com/quad-foundation/quad-node/common"
	"github.com/quad-foundation/quad-node/crypto/oqs"
	"golang.org/x/crypto/sha3"
	"io"
	"log"
	"sync"
)

type Wallet struct {
	password           string
	passwordBytes      []byte
	Iv                 []byte `json:"iv"`
	secretKey          common.PrivKey
	EncryptedSecretKey []byte         `json:"encrypted_secret_key"`
	PublicKey          common.PubKey  `json:"public_key"`
	Address            common.Address `json:"address"`
	signer             oqs.Signature
	mutexDb            sync.RWMutex
	HomePath           string `json:"home_path"`
	WalletNumber       uint8  `json:"wallet_number"`
}

var activeWallet *Wallet

type AnyWallet interface {
	GetWallet() Wallet
}

func InitActiveWallet(walletNumber uint8, password string) {
	var err error
	activeWallet, err = Load(walletNumber, password)
	if err != nil {
		log.Fatalf("wrong password")
	}
}

func (w *Wallet) SetPassword(password string) {
	(*w).password = password
	(*w).passwordBytes = passwordToByte(password)
}

func GetActiveWallet() *Wallet {
	return activeWallet
}

func (w *Wallet) ShowInfo() string {

	s := fmt.Sprintln("Length of public key:", w.PublicKey.GetLength())
	s += fmt.Sprintln("Beginning of public key:", w.PublicKey.GetHex()[:10])
	s += fmt.Sprintln("Address:", w.Address.GetHex())
	s += fmt.Sprintln("Length of private key:", w.GetSecretKey().GetLength())
	s += fmt.Sprintln("Wallet location", w.HomePath)
	s += fmt.Sprintln("Wallet Number", w.WalletNumber)
	fmt.Println(s)
	return s
}

func passwordToByte(password string) []byte {
	sh := make([]byte, 32)
	sha3.ShakeSum256(sh, []byte(password))
	return sh
}

func EmptyWallet(walletNumber uint8) *Wallet {
	return &Wallet{
		password:      "",
		passwordBytes: nil,
		Iv:            nil,
		secretKey:     common.PrivKey{},
		PublicKey:     common.PubKey{},
		Address:       common.Address{},
		signer:        oqs.Signature{},
		mutexDb:       sync.RWMutex{},
		HomePath:      common.DefaultWalletHomePath + common.GetSigName() + "/" + string(walletNumber+'0'),
		WalletNumber:  walletNumber,
	}
}
func GenerateNewWallet(walletNumber uint8, password string) (*Wallet, error) {
	if len(password) < 1 {
		return nil, fmt.Errorf("Password cannot be empty")
	}
	w := EmptyWallet(walletNumber)
	w.SetPassword(password)
	(*w).Iv = generateNewIv()
	var signer oqs.Signature
	//defer signer.Clean()

	// ignore potential errors everywhere
	err := signer.Init(common.GetSigName(), nil)
	if err != nil {
		return nil, err
	}
	pubKey, err := signer.GenerateKeyPair()
	if err != nil {
		return nil, err
	}
	mainAddress, err := common.PubKeyToAddress(pubKey)
	if err != nil {
		return nil, err
	}
	err = w.PublicKey.Init(pubKey, mainAddress)
	if err != nil {
		return nil, err
	}
	(*w).Address = w.PublicKey.GetAddress()
	err = w.secretKey.Init(signer.ExportSecretKey(), w.Address)
	if err != nil {
		return nil, err
	}
	(*w).signer = signer
	fmt.Print(signer.Details())
	return w, nil
}

func generateNewIv() []byte {
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}
	return iv
}

func (w *Wallet) encrypt(v []byte) ([]byte, error) {
	cb, err := aes.NewCipher(w.passwordBytes)
	if err != nil {
		log.Println("Can not create AES function")
		return []byte{}, err
	}
	v = append([]byte("validationTag"), v...)
	ciphertext := make([]byte, aes.BlockSize+len(v))
	stream := cipher.NewCTR(cb, w.Iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], v)
	return ciphertext, nil
}

func (w *Wallet) decrypt(v []byte) ([]byte, error) {
	cb, err := aes.NewCipher(w.passwordBytes)
	if err != nil {
		log.Println("Can not create AES function")
		return []byte{}, err
	}

	plaintext := make([]byte, aes.BlockSize+len(v))
	stream := cipher.NewCTR(cb, w.Iv)
	stream.XORKeyStream(plaintext, v[aes.BlockSize:])
	if !bytes.Equal(plaintext[:len(common.ValidationTag)], []byte(common.ValidationTag)) {
		return nil, fmt.Errorf("Wrong password")
	}
	return plaintext[len(common.ValidationTag):], nil
}

func (w *Wallet) GetMnemonicWords() (string, error) {
	if w.GetSecretKey().GetBytes() == nil {
		return "", fmt.Errorf("You need load wallet first")
	}
	mnemonic, _ := bip39.NewMnemonic(w.GetSecretKey().GetBytes())

	secretKey, _ := bip39.MnemonicToByteArray(mnemonic)
	if !bytes.Equal(secretKey[:w.secretKey.GetLength()], w.secretKey.GetBytes()) {
		log.Println("Can not restore secret key from mnemonic")
		return "", fmt.Errorf("Can not restore secret key from mnemonic")
	}
	return mnemonic, nil
}

func (w *Wallet) RestoreSecretKeyFromMnemonic(mnemonic string) error {
	secretKey, err := bip39.MnemonicToByteArray(mnemonic)
	if err != nil {
		log.Println("Can not restore secret key")
		return err
	}
	err = w.secretKey.Init(secretKey[:common.PrivateKeyLength], w.Address)
	if err != nil {
		return err
	}
	var signer oqs.Signature
	err = signer.Init(common.GetSigName(), w.secretKey.GetBytes())
	if err != nil {
		return err
	}
	(*w).signer = signer
	return nil
}

func (w *Wallet) Store() error {
	if w.GetSecretKey().GetBytes() == nil {
		return fmt.Errorf("You need load wallet first")
	}

	w.mutexDb.Lock()
	defer w.mutexDb.Unlock()
	walletDB, err := leveldb.OpenFile(w.HomePath, nil)
	if err != nil {
		return err
	}
	defer walletDB.Close()

	se, err := w.encrypt(w.secretKey.GetBytes())
	if err != nil {
		log.Println(err)
		return err
	}

	w2 := w
	(*w2).EncryptedSecretKey = make([]byte, len(se))
	copy((*w2).EncryptedSecretKey, se)
	if err != nil {
		log.Println(err)
		return err
	}
	wm, err := json.Marshal(&w2)
	if err != nil {
		log.Println(err)
		return err
	}
	prefix := common.WalletDBPrefix
	prefix[1] = w.WalletNumber
	// Put a key-value pair into the database
	err = walletDB.Put(prefix[:], wm, nil)
	if err != nil {
		return err
	}

	return nil
}
func Load(walletNumber uint8, password string) (*Wallet, error) {

	w := EmptyWallet(walletNumber)
	// Open the database with the provided options
	w.mutexDb.Lock()
	defer w.mutexDb.Unlock()
	walletDB, err := leveldb.OpenFile(w.HomePath, nil)
	if err != nil {
		return nil, err
	}
	defer walletDB.Close()
	prefix := common.WalletDBPrefix
	prefix[1] = walletNumber
	value, err := walletDB.Get(prefix[:], nil)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(value, w)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	w.SetPassword(password)
	ds, err := w.decrypt(w.EncryptedSecretKey)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	err = w.secretKey.Init(ds[:w.secretKey.GetLength()], w.Address)
	if err != nil {
		return nil, err
	}
	var signer oqs.Signature
	err = signer.Init(common.GetSigName(), w.secretKey.GetBytes())
	if err != nil {
		return nil, err
	}
	(*w).signer = signer

	return w, err
}

func (w *Wallet) ChangePassword(password, newPassword string) error {
	if w.passwordBytes == nil {
		return fmt.Errorf("You need load wallet first")
	}
	if !bytes.Equal(passwordToByte(password), w.passwordBytes) {
		return fmt.Errorf("Current password is not valid")
	}
	w2 := &Wallet{
		password:      newPassword,
		passwordBytes: passwordToByte(newPassword),
		Iv:            w.Iv,
		secretKey:     w.secretKey,
		PublicKey:     w.PublicKey,
		Address:       w.Address,
		signer:        w.signer,
		mutexDb:       sync.RWMutex{},
		HomePath:      w.HomePath,
		WalletNumber:  w.WalletNumber,
	}

	err := w2.Store()
	if err != nil {
		log.Println("Can not store new wallet")
		return err
	}
	w, err = Load(w2.WalletNumber, newPassword)
	if err != nil {
		return err
	}
	return nil
}

func (w *Wallet) Sign(data []byte) (*common.Signature, error) {
	if len(data) > 0 {
		signature, err := w.signer.Sign(data)
		if err != nil {
			return nil, err
		}

		sig := &common.Signature{}
		err = sig.Init(signature, w.Address)
		if err != nil {
			return nil, err
		}
		return sig, nil
	}
	return nil, fmt.Errorf("input data are empty")
}

func Verify(msg []byte, sig []byte, pubkey []byte) bool {
	var verifier oqs.Signature
	var err error
	if common.IsValid && common.IsPaused == false {
		err = verifier.Init(common.GetSigName(), nil)
		if err != nil {
			return false
		}
		if verifier.Details().LengthPublicKey == len(pubkey) {
			isVerified, err := verifier.Verify(msg, sig, pubkey)
			if err != nil {
				return false
			}
			return isVerified
		}
	}
	if common.IsValid2 && common.IsPaused2 == false {
		err = verifier.Init(common.GetSigName2(), nil)
		if err != nil {
			return false
		}
		if verifier.Details().LengthPublicKey == len(pubkey) {
			isVerified, err := verifier.Verify(msg, sig, pubkey)
			if err != nil {
				return false
			}
			return isVerified
		}
	}
	return false
}

func (w *Wallet) GetSecretKey() common.PrivKey {
	if w == nil {
		return common.PrivKey{}
	}
	return w.secretKey
}

func (w *Wallet) Check() bool {
	if len(w.GetSecretKey().GetBytes()) == w.GetSecretKey().GetLength() {
		return true
	}
	return false
}
