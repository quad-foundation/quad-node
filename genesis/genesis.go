// Package genesis maintains access to the genesis file.
package genesis

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/quad-foundation/quad-node/account"
	"github.com/quad-foundation/quad-node/blocks"
	"github.com/quad-foundation/quad-node/common"
	"github.com/quad-foundation/quad-node/transactionsDefinition"
	"github.com/quad-foundation/quad-node/transactionsPool"
	"github.com/quad-foundation/quad-node/wallet"
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
	TransactionsQueue      []string          `json:"transactions_queue"`
	Transactions           map[string]int64  `json:"transactions"`
	StakedBalances         map[string]int64  `json:"staked_balances"`
	TransactionsSignatures map[string]string `json:"transactions_signatures"`
	PubKeys                map[string]string `json:"pub_keys"`
	Signature              string            `json:"signature"`
	OperatorPubKey         string            `json:"operator_pub_key"`
}

func CreateBlockFromGenesis(genesis Genesis) blocks.Block {

	initSupplyWithoutStaked := common.InitSupply
	for _, balance := range genesis.StakedBalances {
		initSupplyWithoutStaked -= balance
	}
	pubKeyOpBytes, err := hex.DecodeString(genesis.OperatorPubKey)
	if err != nil {
		log.Fatal("cannot decode address from string in genesis block")
	}
	pubKeyOp1 := common.PubKey{}
	err = pubKeyOp1.Init(pubKeyOpBytes)
	if err != nil {
		log.Fatalf("cannot initialize operator pub key in genesis block %v", err)
	}
	err = blocks.StorePubKey(pubKeyOp1)
	if err != nil {
		log.Fatal("cannot store genesis operator pubkey", err)
	}
	addressOp1, err := common.PubKeyToAddress(pubKeyOp1)
	if err != nil {
		log.Fatalf("cannot retrieve operator address from pub key in genesis block %v", err)
	}
	accDel1 := account.Accounts.AllAccounts[addressOp1.ByteValue]
	accDel1.Balance = initSupplyWithoutStaked
	accDel1.Address = addressOp1.ByteValue
	account.Accounts.AllAccounts[addressOp1.ByteValue] = accDel1

	walletNonce := int16(0)
	blockTransactionsHashesBytes := [][]byte{}
	blockTransactionsHashes := []common.Hash{}
	genesisTxs := []transactionsDefinition.Transaction{}
	for _, addr := range genesis.TransactionsQueue {
		balance := genesis.Transactions[addr]
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
		err = tx.StoreToDBPoolTx(common.TransactionPoolHashesDBPrefix[:])
		if err != nil {
			log.Fatalf("cannot store transaction of genesis block %v", err)
		}
		genesisTxs = append(genesisTxs, tx)
		blockTransactionsHashesBytes = append(blockTransactionsHashesBytes, tx.GetHash().GetBytes())
		blockTransactionsHashes = append(blockTransactionsHashes, tx.GetHash())
		walletNonce++
	}
	for addr, balance := range genesis.StakedBalances {

		ab, err := hex.DecodeString(addr)
		if err != nil {
			log.Fatal("cannot decode address from string in genesis block")
		}
		addrb := [common.AddressLength]byte{}
		copy(addrb[:], ab)
		delAddrb := [common.AddressLength]byte{}
		firstDel := common.GetDelegatedAccountAddress(1)
		copy(delAddrb[:], firstDel.GetBytes())
		sd := account.StakingDetail{
			Amount:      balance,
			Reward:      0,
			LastUpdated: genesis.Timestamp,
		}
		sds := map[int64][]account.StakingDetail{}
		sds[0] = []account.StakingDetail{sd}
		as := account.StakingAccount{
			StakedBalance:      balance,
			StakingRewards:     0,
			DelegatedAccount:   delAddrb,
			Address:            addrb,
			OperationalAccount: true,
			StakingDetails:     sds,
		}
		account.StakingAccounts[1].AllStakingAccounts[addrb] = as
	}
	genesisMerkleTrie, err := transactionsPool.BuildMerkleTree(0, blockTransactionsHashesBytes)
	if err != nil {
		log.Fatalf("cannot generate genesis merkleTrie %v", err)
	}
	defer genesisMerkleTrie.Destroy()

	err = genesisMerkleTrie.StoreTree(0)
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
	pkb, err := hex.DecodeString(genesis.PubKeys[recipient.GetHex()])
	if err != nil {
		log.Fatal(err)
	}
	pk := common.PubKey{}
	err = pk.Init(pkb)
	if err != nil {
		log.Fatal(err)
	}
	err = blocks.StorePubKey(pk)
	if err != nil {
		log.Fatal(err)
	}
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

	err = t.CalcHashAndSet()
	if err != nil {
		log.Fatal("calc hash error", err)
	}

	signature, err := common.GetSignatureFromString(genesis.TransactionsSignatures[recipient.GetHex()], sender)

	if err != nil {
		log.Fatal(err)
	}
	t.Signature = signature

	//myWallet := wallet.GetActiveWallet()
	//err = t.Sign(myWallet)
	//if err != nil {
	//	log.Fatal("Signing error", err)
	//}
	//println(t.Signature.GetHex())

	if t.Verify() == false {
		log.Fatal("genesis transaction cannot be verified")
	}
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
	err = account.StoreAccounts(0)
	if err != nil {
		log.Fatal(err)
	}
	err = account.StoreStakingAccounts(0)
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
		log.Println(mainWallet.PublicKey.GetHex())
		log.Fatal("Main Wallet address should be the same as in config genesis.json file")
	}
	return genesis, nil
}
