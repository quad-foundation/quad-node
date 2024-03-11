// Package genesis maintains access to the genesis file.
package genesis

import (
	"encoding/json"
	"fmt"
	"github.com/quad/quad-node/blocks"
	"github.com/quad/quad-node/common"
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

func CreateBlockFromGenesis(genesis Genesis) (blocks.Block, error) {

	myWallet := wallet.GetActiveWallet()

	//signature := common.Signature{}
	//err := signature.Set([]byte(genesis.Signature), myWallet.Address)
	//if err != nil {
	//	return nil, err
	//}

	bh := blocks.BaseHeader{
		PreviousHash:     common.EmptyHash(),
		Difficulty:       genesis.Difficulty,
		Height:           0,
		DelegatedAccount: common.GetDelegatedAccountAddress(1),
		OperatorAccount:  myWallet.Address,
		RootMerkleTree:   common.EmptyHash(),
		Signature:        common.Signature{},
		SignatureMessage: []byte{},
	}
	signatureBlockHeaderMessage := bh.GetBytesWithoutSignature()
	bh.SignatureMessage = signatureBlockHeaderMessage
	hashb, err := common.CalcHashToByte(signatureBlockHeaderMessage)
	if err != nil {
		return blocks.Block{}, err
	}

	sign, err := myWallet.Sign(hashb)
	if err != nil {
		return blocks.Block{}, err
	}
	bh.Signature = *sign

	bhHash, err := bh.CalcHash()
	if err != nil {
		return blocks.Block{}, err
	}
	bb := blocks.BaseBlock{
		BaseHeader:       bh,
		BlockHeaderHash:  bhHash,
		BlockTimeStamp:   genesis.Timestamp,
		RewardPercentage: 0,
	}

	tempInstance, err := transactionsPool.BuildMerkleTree(0, [][]byte{})
	if err != nil {
		return blocks.Block{}, err
	}
	defer tempInstance.Destroy()
	rmthash := []common.Hash{}

	bl := blocks.Block{
		BaseBlock:          bb,
		Chain:              0,
		TransactionsHashes: rmthash,
		BlockHash:          common.EmptyHash(),
	}
	hash, err := bl.CalcBlockHash()
	if err != nil {
		return blocks.Block{}, err
	}
	bl.BlockHash = hash

	return bl, nil
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

	genesisBlock, err := CreateBlockFromGenesis(genesis)
	if err != nil {
		log.Fatal(err)
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
	//fmt.Println(mainWallet.PublicKey.GetHex())
	fmt.Println(mainWallet.Address.GetHex())

	//del1 := common.GetDelegatedAccountAddress(1)
	//delegatedAccount := common.GetDelegatedAccount()
	//if mainWallet.PublicKey.GetBytes() != nil &&
	//	genesis.OperatorPubKey[:100] != mainWallet.PublicKey.GetHex()[:100] &&
	//	delegatedAccount.GetHex() == del1.GetHex() {
	//	log.Fatal("Main Wallet address should be the same as in config genesis.json file")
	//}
	return genesis, nil
}
