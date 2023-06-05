// Package genesis maintains access to the genesis file.
package genesis

import (
	"encoding/json"
	"fmt"
	"github.com/chainpqc/chainpqc-node/blocks"
	"github.com/chainpqc/chainpqc-node/common"
	"github.com/chainpqc/chainpqc-node/wallet"
	"log"
	"os"
)

// Genesis represents the genesis file.
type Genesis struct {
	Timestamp            int64            `json:"date"`
	ChainID              int16            `json:"chain_id"`   // The chain id represents an unique id for this running instance.
	Difficulty           int32            `json:"difficulty"` // How difficult it needs to be to solve the work problem.
	InitialReward        int64            `json:"initial_reward"`
	RewardChangeInterval int64            `json:"reward_change_interval"`
	Decimals             uint8            `json:"decimals"`
	BlockTimeInterval    float32          `json:"block_time_interval"`
	Balances             map[string]int64 `json:"balances"`
	Signature            string           `json:"signature"`
	OperatorPubKey       string           `json:"operator_pub_key"`
	DelegatedAccount     map[string]int   `json:"delegated_account"`
}

func CreateBlockFromGenesis(genesis Genesis) (blocks.AnyBlock, error) {

	myWallet := wallet.EmptyWallet().GetWallet()

	//signature := common.Signature{}
	//err := signature.Init([]byte(genesis.Signature), myWallet.Address)
	//if err != nil {
	//	return nil, err
	//}

	hashZero := common.Hash{}
	hashZero, err := hashZero.Init(make([]byte, 32))
	if err != nil {
		return nil, err
	}
	bh := blocks.BaseHeader{
		PreviousHash:     hashZero,
		Difficulty:       genesis.Difficulty,
		Height:           0,
		DelegatedAccount: common.GetDelegatedAccountAddress(1),
		OperatorAccount:  myWallet.Address,
		Signature:        common.Signature{},
		SignatureMessage: make([]byte, 32),
	}
	signatureBlockHeaderMessage := bh.GetBytesWithoutSignature()
	hashb, err := common.CalcHashToByte(signatureBlockHeaderMessage)
	if err != nil {
		return nil, err
	}

	sign, err := myWallet.Sign(hashb)
	if err != nil {
		return nil, err
	}
	bh.Signature = sign

	bh.SignatureMessage = signatureBlockHeaderMessage
	bhHash, err := bh.CalcHash()
	if err != nil {
		return nil, err
	}
	bb := blocks.BaseBlock{
		BaseHeader:       bh,
		BlockHeaderHash:  bhHash,
		BlockTimeStamp:   genesis.Timestamp,
		RewardPercentage: 0,
	}
	var anyBlock blocks.AnyBlock

	bl := blocks.TransactionsBlock{
		BaseBlock:        bb,
		Chain:            0,
		TransactionsHash: hashZero,
		BlockHash:        common.Hash{},
	}
	hash, err := bl.CalcBlockHash()
	if err != nil {
		return nil, err
	}
	bl.BlockHash = hash
	anyBlock = blocks.AnyBlock(bl)

	return anyBlock, nil
}

// InitGenesis sets initial values written in genesis conf file
func InitGenesis() {
	pathhome, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	genesis, err := Load(pathhome + "/chainpqc-node/genesis/config/genesis.json")
	if err != nil {
		log.Fatal(err)
	}

	genesisBlock, err := CreateBlockFromGenesis(genesis)
	if err != nil {
		log.Fatal(err)
	}
	err = blocks.StoreBlock(genesisBlock)
	if err != nil {
		log.Fatal(err)
	}
	common.SetHeight(0)
	common.SetChainID(genesis.ChainID)

	common.BlockTimeInterval = genesis.BlockTimeInterval
	common.InitialReward = genesis.InitialReward
	common.RewardChangeInterval = genesis.RewardChangeInterval
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

	mainWallet := wallet.EmptyWallet().GetWallet()
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
