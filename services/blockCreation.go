package services

import (
	"bytes"
	"fmt"
	"github.com/quad/quad-node/account"
	"github.com/quad/quad-node/blocks"
	"github.com/quad/quad-node/common"
	"github.com/quad/quad-node/message"
	"github.com/quad/quad-node/tcpip"
	"github.com/quad/quad-node/transactionsDefinition"
	"github.com/quad/quad-node/transactionsPool"
	"github.com/quad/quad-node/wallet"
	"sync"
)

var (
	SendChanNonce     chan []byte
	SendChanSelfNonce chan []byte
	SendMutexNonce    sync.RWMutex
	SendChanTx        chan []byte
	SendMutexTx       sync.RWMutex
	SendChanSync      chan []byte
	SendMutexSync     sync.RWMutex
)

func CreateBlockFromNonceMessage(nonceTx []transactionsDefinition.Transaction,
	lastBlock blocks.Block,
	merkleTrie *transactionsPool.MerkleTree,
	txs []common.Hash) (blocks.Block, error) {

	myWallet := wallet.GetActiveWallet()
	transactionChain := nonceTx[0].GetChain()
	heightTransaction := nonceTx[0].GetHeight()
	totalFee := int64(0)
	for _, at := range nonceTx {
		heightLastBlocktransaction := common.GetInt64FromByte(at.GetData().GetOptData()[:8])
		hashLastBlocktransaction := at.GetData().GetOptData()[8:40]
		if !bytes.Equal(hashLastBlocktransaction, lastBlock.GetBlockHash().GetBytes()) {
			ha, err := blocks.LoadHashOfBlock(heightTransaction - 1)
			if err != nil {
				return blocks.Block{}, err
			}
			return blocks.Block{}, fmt.Errorf("last block hash and nonce hash do not match", ha, " ", lastBlock.GetBlockHash().GetBytes())
		}
		if heightTransaction != heightLastBlocktransaction+1 {
			return blocks.Block{}, fmt.Errorf("last block height and nonce height do not match")
		}
		totalFee += at.GasUsage * at.GasPrice
	}

	reward := account.GetReward(lastBlock.GetBlockSupply())
	supply := lastBlock.GetBlockSupply() - totalFee + reward

	sendingTimeTransaction := nonceTx[0].GetParam().SendingTime
	ti := sendingTimeTransaction - lastBlock.GetBlockTimeStamp()
	bblock := lastBlock.GetBaseBlock()
	diff := blocks.AdjustDifficulty(bblock.BaseHeader.Difficulty, ti)
	sendingTimeMessage := common.GetByteInt64(nonceTx[0].GetParam().SendingTime)
	rootMerkleTrie := common.Hash{}
	rootMerkleTrie.Set(merkleTrie.GetRootHash())
	bh := blocks.BaseHeader{
		PreviousHash:     lastBlock.GetBlockHash(),
		Difficulty:       diff,
		Height:           heightTransaction,
		DelegatedAccount: common.GetDelegatedAccount(),
		OperatorAccount:  myWallet.Address,
		RootMerkleTree:   rootMerkleTrie,
		Signature:        common.Signature{},
		SignatureMessage: sendingTimeMessage,
	}
	sign, signatureBlockHeaderMessage, err := bh.Sign()
	if err != nil {
		return blocks.Block{}, err
	}
	bh.Signature = sign
	bh.SignatureMessage = signatureBlockHeaderMessage
	bhHash, err := bh.CalcHash()
	if err != nil {
		return blocks.Block{}, err
	}
	bb := blocks.BaseBlock{
		BaseHeader:       bh,
		BlockHeaderHash:  bhHash,
		BlockTimeStamp:   common.GetCurrentTimeStampInSecond(),
		RewardPercentage: common.GetRewardPercentage(),
		Supply:           supply,
	}

	if err != nil {
		return blocks.Block{}, err
	}
	bl := blocks.Block{
		BaseBlock:          bb,
		Chain:              transactionChain,
		TransactionsHashes: txs,
		BlockHash:          common.Hash{},
	}
	hash, err := bl.CalcBlockHash()
	if err != nil {
		return blocks.Block{}, err
	}
	bl.BlockHash = hash

	return bl, nil
}

func GenerateBlockMessage(bl blocks.Block) message.TransactionsMessage {

	bm := message.BaseMessage{
		Head:    []byte("bl"),
		ChainID: common.GetChainID(),
		Chain:   bl.GetChain(),
	}
	txm := [2]byte{}
	copy(txm[:], append([]byte("N"), bm.Chain))
	atm := message.TransactionsMessage{
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
	SendMutexNonce.Lock()
	SendChanNonce <- nb
	SendMutexNonce.Unlock()
}

func BroadcastBlock(bl blocks.Block) {
	atm := GenerateBlockMessage(bl)
	nb := atm.GetBytes()
	SendNonce("0.0.0.0", nb)
	SendNonce(tcpip.MyIP, nb)
}
