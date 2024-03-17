// Package genesis maintains access to the genesis file.
package genesis

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/quad/quad-node/account"
	"github.com/quad/quad-node/blocks"
	"github.com/quad/quad-node/common"
	"github.com/quad/quad-node/transactionsDefinition"
	"github.com/quad/quad-node/transactionsPool"
	"github.com/quad/quad-node/wallet"
	"log"
	"os"
)

// Genesis represents the genesis file.
type Genesis struct {
	Timestamp         int64            `json:"date"`
	ChainID           int16            `json:"chain_id"`   // The chain id represents an unique id for this running instance.
	Difficulty        int32            `json:"difficulty"` // How difficult it needs to be to solve the work problem.
	RewardRatio       float64          `json:"reward_ratio"`
	Decimals          uint8            `json:"decimals"`
	BlockTimeInterval float32          `json:"block_time_interval"`
	Balances          map[string]int64 `json:"balances"`
	Signature         string           `json:"signature"`
	OperatorPubKey    string           `json:"operator_pub_key"`
	DelegatedAccount  map[string]int   `json:"delegated_account"`
}

func CreateBlockFromGenesis(genesis Genesis) blocks.Block {

	myWallet := wallet.GetActiveWallet()

	//signature := common.Signature{}
	//err := signature.Set([]byte(genesis.Signature), myWallet.Address)
	//if err != nil {
	//	return nil, err
	//}
	accDel1 := account.Accounts.AllAccounts[myWallet.Address.ByteValue]
	accDel1.Balance = common.InitSupply
	account.Accounts.AllAccounts[myWallet.Address.ByteValue] = accDel1

	walletNonce := int16(0)
	blockTransactionsHashesBytes := [][]byte{}
	blockTransactionsHashes := []common.Hash{}
	genesisTxs := []transactionsDefinition.Transaction{}
	for addr, balance := range genesis.Balances {
		ab, err := hex.DecodeString(addr)
		if err != nil {
			log.Fatal("cannot decode address from string in genesis block")
		}
		a, err := common.BytesToAddress(ab)
		if err != nil {
			log.Fatal("cannot decode address from bytes in genesis block")
		}
		tx := GenesisTransaction(myWallet, a, balance, walletNonce, genesis)
		err = tx.CalcHashAndSet()
		if err != nil {
			log.Fatalf("cannot calculate hash of transaction in genesis block %v", err)
		}
		err = tx.StoreToDBPoolTx(common.TransactionDBPrefix[:])
		if err != nil {
			log.Fatalf("cannot store transaction of genesis block %v", err)
		}
		genesisTxs = append(genesisTxs, tx)
		blockTransactionsHashesBytes = append(blockTransactionsHashesBytes, tx.GetHash().GetBytes())
		blockTransactionsHashes = append(blockTransactionsHashes, tx.GetHash())
		walletNonce++
	}

	genesisMerkleTrie, err := transactionsPool.BuildMerkleTree(0, blockTransactionsHashesBytes)
	if err != nil {
		log.Fatalf("cannot generate genesis merkleTrie %v", err)
	}
	defer genesisMerkleTrie.Destroy()

	err = genesisMerkleTrie.StoreTree(0, blockTransactionsHashesBytes)
	if err != nil {
		log.Fatalf("cannot store genesis merkleTrie %v", err)
	}
	rootHash := common.Hash{}
	rootHash.Set(genesisMerkleTrie.GetRootHash())

	bh := blocks.BaseHeader{
		PreviousHash:     common.EmptyHash(),
		Difficulty:       genesis.Difficulty,
		Height:           0,
		DelegatedAccount: common.GetDelegatedAccountAddress(1),
		OperatorAccount:  myWallet.Address,
		RootMerkleTree:   rootHash,
		Signature:        common.Signature{},
		SignatureMessage: []byte{},
	}
	signatureBlockHeaderMessage := bh.GetBytesWithoutSignature()
	bh.SignatureMessage = signatureBlockHeaderMessage
	hashb, err := common.CalcHashToByte(signatureBlockHeaderMessage)
	if err != nil {
		log.Fatalf("cannot calculate hash of genesis block header %v", err)
	}

	sign, err := myWallet.Sign(hashb)
	if err != nil {
		log.Fatalf("cannot sign genesis block header %v", err)
	}
	bh.Signature = *sign

	bhHash, err := bh.CalcHash()
	if err != nil {
		log.Fatalf("cannot calculate hash of genesis block header %v", err)
	}
	bb := blocks.BaseBlock{
		BaseHeader:       bh,
		BlockHeaderHash:  bhHash,
		BlockTimeStamp:   genesis.Timestamp,
		RewardPercentage: 0,
		Supply:           common.InitSupply,
	}

	bl := blocks.Block{
		BaseBlock:          bb,
		Chain:              0,
		TransactionsHashes: blockTransactionsHashes,
		BlockHash:          common.EmptyHash(),
	}
	hash, err := bl.CalcBlockHash()
	if err != nil {
		log.Fatalf("cannot calculate hash of genesis block %v", err)
	}
	bl.BlockHash = hash

	return bl
}

func GenesisTransaction(w *wallet.Wallet, recipient common.Address, amount int64, walletNonce int16, genesis Genesis) transactionsDefinition.Transaction {

	sender := w.Address

	txdata := transactionsDefinition.TxData{
		Recipient: recipient,
		Amount:    amount,
		OptData:   nil,
	}
	txParam := transactionsDefinition.TxParam{
		ChainID:     common.GetChainID(),
		Sender:      sender,
		SendingTime: genesis.Timestamp,
		Nonce:       walletNonce,
		Chain:       0,
	}
	t := transactionsDefinition.Transaction{
		TxData:    txdata,
		TxParam:   txParam,
		Hash:      common.Hash{},
		Signature: common.Signature{},
		Height:    0,
		GasPrice:  0,
		GasUsage:  0,
	}

	err := t.CalcHashAndSet()
	if err != nil {
		log.Fatal("calc hash error", err)
	}
	err = t.Sign(w)
	if err != nil {
		log.Fatal("Signing error", err)
	}
	return t
}

// InitGenesis sets initial values written in genesis conf file
func InitGenesis() {
	pathhome, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	genesis, err := Load(pathhome + "/.quad/genesis/config/genesis.json")
	if err != nil {
		log.Fatal(err)
	}

	genesisBlock := CreateBlockFromGenesis(genesis)
	err = genesisBlock.StoreBlock()
	if err != nil {
		log.Fatal(err)
	}
	common.SetHeight(0)
	common.SetChainID(genesis.ChainID)

	common.BlockTimeInterval = genesis.BlockTimeInterval
	common.RewardRatio = genesis.RewardRatio
	common.Decimals = genesis.Decimals
}

// Load opens and consumes the genesis file.
func Load(path string) (Genesis, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Genesis{}, err
	}

	var genesis Genesis
	err = json.Unmarshal(content, &genesis)
	if err != nil {
		return Genesis{}, err
	}

	mainWallet := wallet.GetActiveWallet()
	fmt.Println(mainWallet.Address.GetHex())

	del1 := common.GetDelegatedAccountAddress(1)
	delegatedAccount := common.GetDelegatedAccount()
	if mainWallet.PublicKey.GetBytes() != nil &&
		genesis.OperatorPubKey[:100] != mainWallet.PublicKey.GetHex()[:100] &&
		delegatedAccount.GetHex() == del1.GetHex() {
		log.Fatal("Main Wallet address should be the same as in config genesis.json file")
	}
	return genesis, nil
}
