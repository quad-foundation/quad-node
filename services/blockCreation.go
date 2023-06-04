package services

import (
	"bytes"
	"fmt"
	"github.com/chainpqc/chainpqc-node/blocks"
	blocks6 "github.com/chainpqc/chainpqc-node/blocks/contracts"
	blocks5 "github.com/chainpqc/chainpqc-node/blocks/dex"
	blocks3 "github.com/chainpqc/chainpqc-node/blocks/pubKey"
	blocks4 "github.com/chainpqc/chainpqc-node/blocks/stake"
	blocks2 "github.com/chainpqc/chainpqc-node/blocks/transactions"
	"github.com/chainpqc/chainpqc-node/common"
	"github.com/chainpqc/chainpqc-node/message"
	"github.com/chainpqc/chainpqc-node/tcpip"
	"github.com/chainpqc/chainpqc-node/transactionType"
	"github.com/chainpqc/chainpqc-node/wallet"
	"sync"
)

var (
	SendChan     chan []byte
	SendChanSelf chan []byte
	SendMutex    sync.RWMutex
)

func CreateBlockFromNonceMessage(at transactionType.AnyTransaction, lastBlock blocks.AnyBlock) (blocks.AnyBlock, error) {

	if lastBlock == nil {
		return nil, fmt.Errorf("last block is nil")
	}

	myWallet := wallet.EmptyWallet().GetWallet()
	heightTransaction := at.GetHeight()
	sendingTimeTransaction := at.GetParam().SendingTime
	transactionChain := at.GetChain()
	heightLastBlocktransaction := common.GetInt64FromByte(at.GetData().GetOptData()[:8])
	hashLastBlocktransaction := at.GetData().GetOptData()[8:]

	if bytes.Equal(hashLastBlocktransaction, lastBlock.GetBlockHash().GetBytes()) {
		return nil, fmt.Errorf("last block hash and nonce hash do not match")
	}
	if heightTransaction != heightLastBlocktransaction {
		return nil, fmt.Errorf("last block height and nonce height do not match")
	}

	ti := sendingTimeTransaction - lastBlock.GetBlockTimeStamp()
	bblock := lastBlock.GetBaseBlock()

	diff := blocks.AdjustDifficulty(bblock.BaseHeader.Difficulty, ti)

	bh := blocks.BaseHeader{
		PreviousHash:     lastBlock.GetBlockHash(),
		Difficulty:       diff,
		Height:           heightTransaction,
		DelegatedAccount: common.GetDelegatedAccount(),
		OperatorAccount:  myWallet.Address,
		Signature:        common.Signature{},
		SignatureMessage: []byte{},
	}
	signatureBlockHeaderMessage := bh.GetBytesWithoutSignature()

	sign, err := myWallet.Sign(signatureBlockHeaderMessage)
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
		BlockTimeStamp:   common.GetCurrentTimeStampInSecond(),
		RewardPercentage: common.GetRewardPercentage(),
	}
	var anyBlock blocks.AnyBlock
	switch transactionChain {
	case 0:
		bl := blocks2.TransactionsBlock{
			BaseBlock:        bb,
			Chain:            transactionChain,
			TransactionsHash: common.Hash{},
			BlockHash:        common.Hash{},
		}
		hash, err := bl.CalcBlockHash()
		if err != nil {
			return nil, err
		}
		bl.BlockHash = hash
		anyBlock = blocks.AnyBlock(bl)
	case 1:
		bl := blocks3.PubKeysBlock{
			BaseBlock:        bb,
			Chain:            transactionChain,
			TransactionsHash: common.Hash{},
			BlockHash:        common.Hash{},
		}
		hash, err := bl.CalcBlockHash()
		if err != nil {
			return nil, err
		}
		bl.BlockHash = hash
		anyBlock = blocks.AnyBlock(bl)
	case 2:
		bl := blocks4.StakesBlock{
			BaseBlock:        bb,
			Chain:            transactionChain,
			TransactionsHash: common.Hash{},
			BlockHash:        common.Hash{},
		}
		hash, err := bl.CalcBlockHash()
		if err != nil {
			return nil, err
		}
		bl.BlockHash = hash
		anyBlock = blocks.AnyBlock(bl)
	case 3:
		bl := blocks5.DexBlock{
			BaseBlock:        bb,
			Chain:            transactionChain,
			TransactionsHash: common.Hash{},
			BlockHash:        common.Hash{},
		}
		hash, err := bl.CalcBlockHash()
		if err != nil {
			return nil, err
		}
		bl.BlockHash = hash
		anyBlock = blocks.AnyBlock(bl)
	case 4:
		bl := blocks6.ContractsBlock{
			BaseBlock:        bb,
			Chain:            transactionChain,
			TransactionsHash: common.Hash{},
			BlockHash:        common.Hash{},
		}
		hash, err := bl.CalcBlockHash()
		if err != nil {
			return nil, err
		}
		bl.BlockHash = hash
		anyBlock = blocks.AnyBlock(bl)
	default:
		return nil, fmt.Errorf("Chain is not valid")
	}

	return anyBlock, nil
}

func GenerateBlockMessage(bl blocks.AnyBlock) message.AnyTransactionsMessage {

	bm := message.BaseMessage{
		Head:    []byte("bl"),
		ChainID: common.GetChainID(),
		Chain:   bl.GetChain(),
	}
	txm := [2]byte{}
	copy(txm[:], "bs")
	atm := message.AnyTransactionsMessage{
		BaseMessage:       bm,
		TransactionsBytes: map[[2]byte][][]byte{},
	}
	atm.TransactionsBytes[txm] = [][]byte{bl.GetBytes()}

	return atm
}

func SendNonce(ip string, nb []byte) {
	bip := []byte(ip)
	lip := common.GetByteInt16(int16(len(bip)))
	lip = append(lip, bip...)
	nb = append(lip, nb...)
	SendMutex.Lock()
	SendChan <- nb
	SendMutex.Unlock()
}

func BroadcastBlock(bl blocks.AnyBlock) {
	atm := GenerateBlockMessage(bl)
	nb := atm.GetBytes()
	SendNonce("0.0.0.0", nb)
	SendNonce(tcpip.MyIP, nb)
}
