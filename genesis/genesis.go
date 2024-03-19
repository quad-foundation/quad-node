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
	Timestamp              int64             `json:"date"`
	ChainID                int16             `json:"chain_id"`   // The chain id represents an unique id for this running instance.
	Difficulty             int32             `json:"difficulty"` // How difficult it needs to be to solve the work problem.
	RewardRatio            float64           `json:"reward_ratio"`
	Decimals               uint8             `json:"decimals"`
	BlockTimeInterval      float32           `json:"block_time_interval"`
	Balances               map[string]int64  `json:"balances"`
	TransactionsSignatures map[string]string `json:"transactions_signatures"`
	Signature              string            `json:"signature"`
	OperatorPubKey         string            `json:"operator_pub_key"`
}

func CreateBlockFromGenesis(genesis Genesis) blocks.Block {

	pubKeyOpBytes, err := hex.DecodeString(genesis.OperatorPubKey)
	if err != nil {
		log.Fatal("cannot decode address from string in genesis block")
	}
	pubKeyOp1 := common.PubKey{}
	err = pubKeyOp1.Init(pubKeyOpBytes)
	if err != nil {
		log.Fatalf("cannot initialize operator pub key in genesis block %v", err)
	}

	addressOp1, err := common.PubKeyToAddress(pubKeyOp1)
	if err != nil {
		log.Fatalf("cannot retrieve operator address from pub key in genesis block %v", err)
	}
	accDel1 := account.Accounts.AllAccounts[addressOp1.ByteValue]
	accDel1.Balance = common.InitSupply
	account.Accounts.AllAccounts[addressOp1.ByteValue] = accDel1

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
		tx := GenesisTransaction(addressOp1, a, balance, walletNonce, genesis)
		err = tx.CalcHashAndSet()
		if err != nil {
			log.Fatalf("cannot calculate hash of transaction in genesis block %v", err)
		}
		prefix := []byte{common.TransactionDBPrefix[0], 0}
		err = tx.StoreToDBPoolTx(prefix)
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
		OperatorAccount:  addressOp1,
		RootMerkleTree:   rootHash,
		Signature:        common.Signature{},
		SignatureMessage: []byte{},
	}
	signatureBlockHeaderMessage := bh.GetBytesWithoutSignature()
	bh.SignatureMessage = signatureBlockHeaderMessage
	_, err = common.CalcHashToByte(signatureBlockHeaderMessage)
	if err != nil {
		log.Fatalf("cannot calculate hash of genesis block header %v", err)
	}

	//myWallet := wallet.GetActiveWallet()
	//sign, err := myWallet.Sign(hashb)
	//if err != nil {
	//	log.Fatalf("cannot sign genesis block header %v", err)
	//}
	//bh.Signature = *sign
	//log.Println("Block Signature:", bh.Signature.GetHex())

	signature, err := common.GetSignatureFromString(genesis.Signature, addressOp1)
	if err != nil {
		log.Fatal(err)
	}
	bh.Signature = signature
	bhHash, err := bh.CalcHash()
	if err != nil {
		log.Fatalf("cannot calculate hash of genesis block header %v", err)
	}
	if bh.Verify() == false {
		log.Fatal("Block Header signature fails to verify")
	}
	bb := blocks.BaseBlock{
		BaseHeader:       bh,
		BlockHeaderHash:  bhHash,
		BlockTimeStamp:   genesis.Timestamp,
		RewardPercentage: 0,
		Supply:           common.InitSupply + account.GetReward(common.InitSupply),
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

func GenesisTransaction(sender common.Address, recipient common.Address, amount int64, walletNonce int16, genesis Genesis) transactionsDefinition.Transaction {

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
	signature, err := common.GetSignatureFromString(genesis.TransactionsSignatures[recipient.GetHex()], sender)

	if err != nil {
		log.Fatal(err)
	}
	t.Signature = signature
	//err = t.Sign(w)
	//if err != nil {
	//	log.Fatal("Signing error", err)
	//}
	log.Println("transaction signature: ", t.Signature.GetHex())
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
	reward := account.GetReward(common.InitSupply)
	err = blocks.ProcessBlockTransfers(genesisBlock, reward)
	if err != nil {
		log.Fatalf("cannot process transactions in genesis block %v", err)
	}
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

func ResetAccountsAndBlocksSync(height int64) {
	common.IsSyncing.Store(true)
	account.Accounts.AllAccounts = map[[20]byte]account.Account{}
	InitGenesis()
}
