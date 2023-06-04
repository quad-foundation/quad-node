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

	"github.com/chainpqc/chainpqc-node/common"
	"github.com/chainpqc/chainpqc-node/crypto/oqs"
	"github.com/chainpqc/chainpqc-node/database"
	"golang.org/x/crypto/sha3"
	"io"
	"log"
	"os"
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
}

var mainWallet Wallet
var mutexDb sync.Mutex
var HomePath string

type AnyWallet interface {
	GetWallet() Wallet
}

func init() {
	var err error
	//err = os.MkdirAll("/.chainpqc/db/wallet/"+common.GetSigName(), 0755)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//err = os.MkdirAll("/.chainpqc/db/blockchain/", 0755)
	//if err != nil {
	//	log.Fatal(err)
	//}
	mainWallet = EmptyWallet()

	HomePath, err = os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	HomePath += "/.chainpqc/db/wallet/"
	mainWallet.SetPassword("a")
	err = mainWallet.Load()
	if err != nil {
		return
	}
}

func (w *Wallet) SetPassword(password string) {
	w.password = password
	w.passwordBytes = passwordToByte(password)
}

func (w Wallet) GetWallet() Wallet {
	return mainWallet
}

func (w Wallet) ShowInfo() string {

	s := fmt.Sprintln("Length of public key:", w.PublicKey.GetLength())
	s += fmt.Sprintln("Beginning of public key:", w.PublicKey.GetHex()[:10])

	s += fmt.Sprintln("Address:", w.Address.GetHex())
	s += fmt.Sprintln("Length of private key:", w.secretKey.GetLength())
	s += fmt.Sprintln("Private key:", w.secretKey.GetHex())

	fmt.Println(s)
	return s
}

func passwordToByte(password string) []byte {
	sh := make([]byte, 32)
	sha3.ShakeSum256(sh, []byte(password))
	return sh
}

func EmptyWallet() Wallet {
	w := Wallet{
		password:      "",
		passwordBytes: nil,
		Iv:            nil,
		secretKey:     common.PrivKey{},
		PublicKey:     common.PubKey{},
		Address:       common.Address{},
		signer:        oqs.Signature{},
	}
	return w
}
func GenerateNewWallet(password string) (Wallet, error) {
	if len(password) < 1 {
		return EmptyWallet(), fmt.Errorf("Password cannot be empty")
	}
	w := EmptyWallet()
	w.password = password
	w.passwordBytes = passwordToByte(password)
	w.Iv = generateNewIv()
	var signer oqs.Signature
	//defer signer.Clean()

	// ignore potential errors everywhere
	err := signer.Init(common.GetSigName(), nil)
	if err != nil {
		return Wallet{}, err
	}
	pubKey, err := signer.GenerateKeyPair()
	if err != nil {
		return Wallet{}, err
	}
	err = w.PublicKey.Init(pubKey)
	if err != nil {
		return Wallet{}, err
	}
	w.Address = w.PublicKey.GetAddress()
	err = w.secretKey.Init(signer.ExportSecretKey(), w.Address)
	if err != nil {
		return Wallet{}, err
	}
	w.signer = signer
	return w, nil
}

func generateNewIv() []byte {
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}
	return iv
}

func (w Wallet) encrypt(v []byte) ([]byte, error) {
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

func (w Wallet) GetMnemonicWords() (string, error) {
	if w.secretKey.GetBytes() == nil {
		return "", fmt.Errorf("You need load wallet first")
	}
	mnemonic, _ := bip39.NewMnemonic(w.secretKey.GetBytes())

	secretKey, _ := bip39.MnemonicToByteArray(mnemonic)
	if bytes.Compare(secretKey[:w.secretKey.GetLength()], w.secretKey.GetBytes()) != 0 {
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
	w.secretKey.Init(secretKey[:common.PrivateKeyLength], w.Address)
	var signer oqs.Signature
	signer.Init(common.GetSigName(), w.secretKey.GetBytes())
	w.signer = signer
	return nil
}

func (w Wallet) decrypt(v []byte) ([]byte, error) {
	cb, err := aes.NewCipher(w.passwordBytes)
	if err != nil {
		log.Println("Can not create AES function")
		return []byte{}, err
	}

	plaintext := make([]byte, aes.BlockSize+len(v))
	stream := cipher.NewCTR(cb, w.Iv)
	stream.XORKeyStream(plaintext, v[aes.BlockSize:])
	if bytes.Compare(plaintext[:len(common.ValidationTag)], []byte(common.ValidationTag)) != 0 {
		return nil, fmt.Errorf("Wrong password")
	}
	return plaintext[len(common.ValidationTag):], nil
}

func (w Wallet) Store() error {
	if w.secretKey.GetBytes() == nil {
		return fmt.Errorf("You need load wallet first")
	}

	mutexDb.Lock()
	walletDB, err := leveldb.OpenFile(HomePath+common.GetSigName(), nil)
	if err != nil {
		return err
	}
	defer walletDB.Close()
	defer mutexDb.Unlock()

	se, err := w.encrypt(w.secretKey.GetBytes())
	if err != nil {
		log.Println(err)
		return err
	}

	w2 := w
	w2.EncryptedSecretKey = make([]byte, len(se))
	copy(w2.EncryptedSecretKey, se)
	if err != nil {
		log.Println(err)
		return err
	}
	wm, err := json.Marshal(&w2)
	if err != nil {
		log.Println(err)
		return err
	}

	// Put a key-value pair into the database
	err = walletDB.Put([]byte("main_account"), wm, nil)
	if err != nil {
		return err
	}

	return nil
}

func (w Wallet) Sign(data []byte) (sig common.Signature, err error) {
	if len(data) > 0 {
		signature, err := w.signer.Sign(data)
		if err != nil {
			return common.Signature{}, err
		}

		//signature := rand2.RandomBytes(common.SignatureLength)

		err = sig.Init(signature, w.Address)
		if err != nil {
			return common.Signature{}, err
		}
		return sig, nil
	}
	return common.Signature{}, fmt.Errorf("input data are empty")
}

func (w Wallet) GetSecretKey() common.PrivKey {
	return w.secretKey
}

func (w Wallet) Check() bool {
	if len(w.secretKey.GetBytes()) == w.secretKey.GetLength() {
		return true
	}
	return false
}

func (w *Wallet) Load() error {

	// Open the database with the provided options
	mutexDb.Lock()
	walletDB, err := leveldb.OpenFile(HomePath+common.GetSigName(), nil)
	if err != nil {
		return err
	}
	defer walletDB.Close()

	defer mutexDb.Unlock()

	value, err := walletDB.Get([]byte("main_account"), nil)
	if err != nil {
		return err
	}

	err = json.Unmarshal(value, w)
	if err != nil {
		log.Println(err)
		return err
	}
	ds, err := w.decrypt(w.EncryptedSecretKey)
	if err != nil {
		log.Println(err)
		*w = EmptyWallet()
		return err
	}
	w.secretKey.Init(ds[:w.secretKey.GetLength()], w.Address)
	var signer oqs.Signature
	signer.Init(common.GetSigName(), (*w).secretKey.GetBytes())
	w.signer = signer

	mainWallet = *w
	return err
}

func (w *Wallet) ChangePassword(password, newPassword string) error {
	if w.passwordBytes == nil {
		return fmt.Errorf("You need load wallet first")
	}
	w2 := Wallet{
		password:      password,
		passwordBytes: nil,
		Iv:            nil,
		secretKey:     common.PrivKey{},
		PublicKey:     common.PubKey{},
		Address:       common.Address{},
		signer:        oqs.Signature{},
	}
	w2.passwordBytes = passwordToByte(password)
	err := w2.Load()
	if err != nil {
		log.Println("Current password not valid")
		return err
	}
	w.password = newPassword
	w.passwordBytes = passwordToByte(newPassword)
	err = w.Store()
	if err != nil {
		log.Println("Can not store new wallet")
		return err
	}
	return nil
}

func Verify(msg []byte, sig common.Signature, pubkey common.PubKey) bool {
	var verifier oqs.Signature
	err := verifier.Init(common.GetSigName(), nil)
	if err != nil {
		return false
	}

	isVerified, err := verifier.Verify(msg, sig.GetBytes(), pubkey.GetBytes())
	if err != nil {
		return false
	}
	return isVerified
	//return true
}

func LoadPubKey(addr common.Address) (pk common.PubKey, err error) {
	val, err := memDatabase.LoadBytes(append([]byte(common.PubKeyDBPrefix), addr.GetBytes()...))
	if err != nil {
		return pk, err
	}
	err = pk.Init(val)
	if err != nil {
		return pk, err
	}
	return pk, nil
}
