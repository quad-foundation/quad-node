package nonceMsg

import (
	"fmt"
	"github.com/chainpqc/chainpqc-node/common"
	"github.com/chainpqc/chainpqc-node/message"
	"github.com/chainpqc/chainpqc-node/tcpip"
	"github.com/chainpqc/chainpqc-node/transactionType"
	transactionType6 "github.com/chainpqc/chainpqc-node/transactionType/contractChainTransaction"
	transactionType5 "github.com/chainpqc/chainpqc-node/transactionType/dexChainTransaction"
	transactionType2 "github.com/chainpqc/chainpqc-node/transactionType/mainChainTransaction"
	transactionType3 "github.com/chainpqc/chainpqc-node/transactionType/pubKeyChainTransaction"
	transactionType4 "github.com/chainpqc/chainpqc-node/transactionType/stakeChainTransaction"
	"github.com/chainpqc/chainpqc-node/wallet"
	"log"
	"sync"
	"time"
)

var sendChan chan []byte
var sendChanSelf chan []byte
var sendMutex sync.RWMutex

func InitNonceService() {
	sendMutex.Lock()
	sendChan = make(chan []byte)

	sendChanSelf = make(chan []byte)
	sendMutex.Unlock()
	startPublishingNonceMsg()
	time.Sleep(time.Second)
	go sendNonceMsgInLoop()
}

func generateNonceMsg(chain uint8) (message.AnyNonceMessage, error) {
	common.HeightMutex.RLock()
	h := common.GetHeight()
	common.HeightMutex.RUnlock()

	var nonceTransaction transactionType.AnyTransaction
	tp := transactionType.TxParam{
		ChainID:     common.GetChainID(),
		Sender:      wallet.EmptyWallet().GetWallet().Address,
		SendingTime: common.GetCurrentTimeStampInSecond(),
		Nonce:       0,
		Chain:       chain,
	}

	switch chain {
	case 0:
		nonceTransaction = transactionType2.MainChainTransaction{
			TxData:    transactionType2.MainChainTxData{},
			TxParam:   tp,
			Hash:      common.Hash{},
			Signature: common.Signature{},
			Height:    h + 1,
			GasPrice:  0,
			GasUsage:  0,
		}
	case 1:
		nonceTransaction = transactionType3.PubKeyChainTransaction{
			TxData:    transactionType3.PubKeyChainTxData{},
			TxParam:   tp,
			Hash:      common.Hash{},
			Signature: common.Signature{},
			Height:    h + 1,
			GasPrice:  0,
			GasUsage:  0,
		}
	case 2:
		nonceTransaction = transactionType4.StakeChainTransaction{
			TxData:    transactionType4.StakeChainTxData{},
			TxParam:   tp,
			Hash:      common.Hash{},
			Signature: common.Signature{},
			Height:    h + 1,
			GasPrice:  0,
			GasUsage:  0,
		}
	case 3:
		nonceTransaction = transactionType5.DexChainTransaction{
			TxData:    transactionType5.DexChainTxData{},
			TxParam:   tp,
			Hash:      common.Hash{},
			Signature: common.Signature{},
			Height:    h + 1,
			GasPrice:  0,
			GasUsage:  0,
		}
	case 4:
		nonceTransaction = transactionType6.ContractChainTransaction{
			TxData:    transactionType6.ContractChainTxData{},
			TxParam:   tp,
			Hash:      common.Hash{},
			Signature: common.Signature{},
			Height:    h + 1,
			GasPrice:  0,
			GasUsage:  0,
		}
	default:
		return message.AnyNonceMessage{}, fmt.Errorf("chain is not correct")
	}
	bm := message.BaseMessage{Head: []byte("nn"),
		ChainID: common.GetChainID(),
		Chain:   chain}

	bb, err := transactionType.SignTransactionAllToBytes(nonceTransaction)
	if err != nil {
		return message.AnyNonceMessage{}, fmt.Errorf("error signing transaction: %v", err)
	}
	hb := [2]byte{}
	copy(hb[:], "nn")
	n := message.AnyNonceMessage{
		BaseMessage: bm,
		NonceBytes:  map[[2]byte][][]byte{hb: {bb}},
	}
	//fmt.Printf("%v", n)
	return n, nil
}

func sendNonceMsgInLoopSelf(chanRecv chan []byte) {

Q:
	for range time.Tick(time.Second) {
		chain := common.GetChainForHeight(common.GetHeight() + 1)
		sendNonceMsg(tcpip.MyIP, chain)
		select {
		case s := <-chanRecv:
			if len(s) == 4 && string(s) == "EXIT" {
				break Q
			}
		default:
		}
	}
}

func sendNonceMsg(ip string, chain uint8) {
	isync := common.IsSyncing.Load()
	if isync == true {
		return
	}
	n, err := generateNonceMsg(chain)
	if err != nil {
		log.Println(err)
		return
	}
	Send(ip, n.GetBytes())
}

func Send(addr string, nb []byte) {
	bip := []byte(addr)
	lip := common.GetByteInt16(int16(len(bip)))
	lip = append(lip, bip...)
	nb = append(lip, nb...)
	sendMutex.Lock()
	sendChan <- nb
	sendMutex.Unlock()
}

func sendNonceMsgInLoop() {
	for range time.Tick(time.Second * 10) {
		chain := common.GetChainForHeight(common.GetHeight() + 1)
		sendNonceMsg("0.0.0.0", chain)
	}
}

func startPublishingNonceMsg() {
	sendMutex.Lock()
	for i := 0; i < 5; i++ {
		go tcpip.StartNewListener(sendChan, tcpip.NonceTopic[i])
		go tcpip.StartNewListener(sendChanSelf, tcpip.SelfNonceTopic[i])
	}
	sendMutex.Unlock()
}

func StartSubscribingNonceMsg(ip string, chain uint8) {
	recvChan := make(chan []byte)

	go tcpip.StartNewConnection(ip, recvChan, tcpip.NonceTopic[chain])
	log.Println("Enter connection receiving loop (nonce msg)", ip)
Q:

	for {
		select {
		case s := <-recvChan:
			if len(s) == 4 && string(s) == "EXIT" {
				break Q
			}
			if len(s) > 2 {
				l := common.GetInt16FromByte(s[:2])
				if len(s) > 2+int(l) {
					ipr := string(s[2 : 2+l])

					OnMessage(ipr, s[2+l:])
				}
			}

		case <-tcpip.Quit:
			break Q
		default:
		}

	}
	log.Println("Exit connection receiving loop (nonce msg)", ip)
}

func StartSubscribingNonceMsgSelf(chain uint8) {
	recvChanSelf := make(chan []byte)
	recvChanExit := make(chan []byte)

	go tcpip.StartNewConnection(tcpip.MyIP, recvChanSelf, tcpip.SelfNonceTopic[chain])
	go sendNonceMsgInLoopSelf(recvChanExit)
	log.Println("Enter connection receiving loop (nonce msg self)")
Q:

	for {
		select {
		case s := <-recvChanSelf:
			if len(s) == 4 && string(s) == "EXIT" {
				recvChanExit <- s
				break Q
			}
			if len(s) > 2 {
				l := common.GetInt16FromByte(s[:2])
				if len(s) > 2+int(l) {
					ipr := string(s[2 : 2+l])

					OnMessage(ipr, s[2+l:])
				}
			}
		case <-tcpip.Quit:
			break Q
		default:
		}

	}
	log.Println("Exit connection receiving loop (nonce msg self)")
}

//func (n *AnyNonceMsg) CreateBlockFromNonceMain(lastBlock block.BlockSideChain,
//	hashes []common.HHash, stateAccountsHash common.HHash,
//	stakeHashes []common.HHash) (block.BlockMainChain, error) {
//
//	height := common.GetInt64FromByte(n.Value["last_height_msg"])
//	ts := common.GetInt64FromByte(n.Value["timestamp_msg"])
//
//	sigAddr := common.Address{}
//	err := sigAddr.Init(n.Value["operator_address_msg"])
//	if err != nil {
//		log.Println("Can not create operator of nonce msg address", err)
//		return block.BlockMainChain{}, err
//	}
//	sigMsg := common.Signature{}
//	err = sigMsg.Init(n.Value["signature_msg"], sigAddr)
//	if err != nil {
//		log.Println("Can not obtain signature hash from nonceMsg")
//		return block.BlockMainChain{}, err
//	}
//
//	if bytes.Compare(n.Value["last_block_hash_msg"], lastBlock.BlockHash.GetByte()) != 0 {
//		log.Println("last block hash and nonce hash do not match")
//		return block.BlockMainChain{}, err
//	}
//
//	lbh := common.HHash{}
//	lbh, err = lbh.Init(n.Value["last_block_hash_msg"])
//	if err != nil {
//		log.Println("Can not obtain last block hash from nonceMsg")
//		return block.BlockMainChain{}, err
//	}
//
//	ti := ts - lastBlock.BaseBlock.BlockTimeStamp
//	diff := block.AdjustDifficulty(lastBlock.BaseBlock.Header.Difficulty, ti)
//
//	bh := block.Header{
//		PreviousHash:     lastBlock.BlockHash,
//		Difficulty:       diff,
//		Height:           height,
//		DelegatedAccount: common.DelegatedAccount,
//		OperatorAccount:  wallet.MainWallet.Address,
//		SignatureMsg:     sigMsg,
//		NonceMessage:     n.GetByte(),
//	}
//	hhbh := crypto.GetHHashFromByte(bh.GetByte()...)
//
//	bb := block.Block{
//		Header:            bh,
//		StateAccountsHash: stateAccountsHash,   // todo
//		OutputLogsHash:    common.EmptyHHash(), // todo
//		TransactionsHash:  common.EmptyHHash(),
//		BlockHeaderHash:   hhbh,
//		BlockTimeStamp:    ts,
//		RewardPercentage:  common.RewardPercentage,
//		PriceOracle:       []byte{},
//		RandOracle:        []byte{},
//		BridgeOracle:      []byte{},
//	}
//
//	txhSum, err := common.SumHHashes(hashes)
//	if err != nil {
//		log.Println(err)
//		return block.BlockMainChain{}, err
//	}
//	if txhSum.Len == common.HHashLength {
//		bb.TransactionsHash = txhSum
//	}
//
//	stakehSum, err := common.SumHHashes(stakeHashes)
//	if err != nil {
//		log.Println(err)
//		return block.BlockMainChain{}, err
//	}
//	if len(stakeHashes) == 0 {
//		stakehSum = common.EmptyHHash()
//	}
//
//	b := block.BlockMainChain{
//		BaseBlock:   bb,
//		Chain:       0,
//		StakingHash: stakehSum,
//		BlockHash:   common.EmptyHHash(),
//	}
//	hhb := b.CalcHHash()
//	b.BlockHash = hhb
//	return b, nil
//}
//
//func (n *AnyNonceMsg) CreateBlockFromNonceSide(lastBlock block.BlockMainChain,
//	hashes []common.HHash, stateAccountsHash common.HHash) (block.BlockSideChain, error) {
//
//	height := common.GetInt64FromByte(n.Value["last_height_msg"])
//	ts := common.GetInt64FromByte(n.Value["timestamp_msg"])
//
//	sigAddr := common.Address{}
//	err := sigAddr.Init(n.Value["operator_address_msg"])
//	if err != nil {
//		log.Println("Can not create operator of nonce msg address", err)
//		return block.BlockSideChain{}, err
//	}
//
//	sigMsg := common.Signature{}
//	err = sigMsg.Init(n.Value["signature_msg"], sigAddr)
//	if err != nil {
//		log.Println("Can not obtain signature hash from nonceMsg")
//		return block.BlockSideChain{}, err
//	}
//
//	if bytes.Compare(n.Value["last_block_hash_msg"], lastBlock.BlockHash.GetByte()) != 0 {
//		log.Println("last block hash and nonce hash do not match")
//		return block.BlockSideChain{}, err
//	}
//
//	lbh := common.HHash{}
//	lbh, err = lbh.Init(n.Value["last_block_hash_msg"])
//	if err != nil {
//		log.Println("Can not obtain last block hash from nonceMsg")
//		return block.BlockSideChain{}, err
//	}
//
//	ti := ts - lastBlock.BaseBlock.BlockTimeStamp
//	diff := block.AdjustDifficulty(lastBlock.BaseBlock.Header.Difficulty, ti)
//
//	bh := block.Header{
//		PreviousHash:     lastBlock.BlockHash,
//		Difficulty:       diff,
//		Height:           height,
//		DelegatedAccount: common.DelegatedAccount,
//		OperatorAccount:  wallet.MainWallet.Address,
//		SignatureMsg:     sigMsg,
//		NonceMessage:     n.GetByte(),
//	}
//	hhbh := crypto.GetHHashFromByte(bh.GetByte()...)
//
//	bb := block.Block{
//		Header:            bh,
//		StateAccountsHash: stateAccountsHash,
//		OutputLogsHash:    common.EmptyHHash(), // todo
//		TransactionsHash:  common.EmptyHHash(),
//		BlockHeaderHash:   hhbh,
//		BlockTimeStamp:    ts,
//		RewardPercentage:  common.RewardPercentage,
//		PriceOracle:       []byte{},
//		RandOracle:        []byte{},
//		BridgeOracle:      []byte{},
//	}
//
//	txhSum, err := common.SumHHashes(hashes)
//	if err != nil {
//		log.Println(err)
//		return block.BlockSideChain{}, err
//	}
//	if txhSum.Len == common.HHashLength {
//		bb.TransactionsHash = txhSum
//	}
//
//	b := block.BlockSideChain{
//		BaseBlock: bb,
//		BlockHash: common.EmptyHHash(),
//		Chain:     1,
//	}
//	hhb := b.CalcHHash()
//	b.BlockHash = hhb
//	return b, nil
//}
//
//func GetPriceThreshold() int64 {
//	// todo
//	return 0
//}
//
//func CreateBlockHashesFromBlock(bl block.AnyBlock, hashes []common.HHash, stakeHashes []common.HHash, stateAcount []byte) block.AnyBlockHashes {
//
//	var blh block.AnyBlockHashes
//	switch bl.GetChain() {
//	case 0:
//		b := block.BlockHashesMainChain{
//			BlockMainChain:    bl.(block.BlockMainChain),
//			TransactionHashes: []common.HHash{},
//			StakingHashes:     []common.HHash{},
//			OutputLogsHashes:  []common.HHash{},
//			StateAccounts:     stateAcount,
//		}
//		if len(hashes) > 0 {
//			b.TransactionHashes = hashes
//		}
//		if len(stakeHashes) > 0 {
//			b.StakingHashes = stakeHashes
//		}
//		blh = &b
//	case 1:
//		b := block.BlockHashesSideChain{
//			BlockSideChain:    bl.(block.BlockSideChain),
//			TransactionHashes: []common.HHash{},
//			OutputLogsHashes:  []common.HHash{},
//			StateAccounts:     stateAcount,
//		}
//		if len(hashes) > 0 {
//			b.TransactionHashes = hashes
//		}
//		blh = &b
//	}
//
//	return blh
//}
//
//func GenerateBlockMessageHashes(bl block.AnyBlockHashes) (AnyNonceMsg, []byte) {
//	bm := message2.BaseMessage{
//		Head:    "block",
//		ChainID: common.ChainID,
//	}
//
//	n := message2.Message{
//		BaseMessage: bm,
//		Value:       map[string][]byte{},
//	}
//
//	blm := bl.Marshal()
//	n.Value["block"] = blm
//
//	an := AnyNonceMsg(n)
//	nb, err := json.Marshal(an)
//	if err != nil {
//		log.Println("Can not marshal block message with error (main chain)", err)
//		return an, []byte{}
//	}
//	return an, nb
//}
//
//func SendNonce(ip string, nb []byte) {
//	bip := []byte(ip)
//	lip := common.GetByteInt16(int16(len(bip)))
//	lip = append(lip, bip...)
//	nb = append(lip, nb...)
//	sendMutex.Lock()
//	sendChan <- nb
//	sendMutex.Unlock()
//}
