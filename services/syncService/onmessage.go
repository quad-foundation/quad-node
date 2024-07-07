package syncServices

import (
	"bytes"
	"github.com/quad-foundation/quad-node/account"
	"github.com/quad-foundation/quad-node/blocks"
	"github.com/quad-foundation/quad-node/common"
	"github.com/quad-foundation/quad-node/message"
	"github.com/quad-foundation/quad-node/services"
	"github.com/quad-foundation/quad-node/services/transactionServices"
	"github.com/quad-foundation/quad-node/statistics"
	"github.com/quad-foundation/quad-node/tcpip"
	"github.com/quad-foundation/quad-node/transactionsPool"
	"log"
)

func OnMessage(addr string, m []byte) {
	h := common.GetHeight()
	if tcpip.IsIPBanned(addr, h, tcpip.SyncTopic) {
		return
	}
	//log.Println("New message nonce from:", addr)
	msg := message.TransactionsMessage{}

	defer func() {
		if r := recover(); r != nil {
			//debug.PrintStack()
			log.Println("recover (nonce Msg)", r)
		}

	}()

	amsg, err := msg.GetFromBytes(m)
	if err != nil {
		panic(err)
	}

	isValid := message.CheckMessage(amsg)
	if isValid == false {
		log.Println("message is invalid")
		panic("message is invalid")
	}

	switch string(amsg.GetHead()) {
	case "hi": // getheader

		txn := amsg.(message.TransactionsMessage).GetTransactionsBytes()

		if tcpip.GetPeersCount() < common.MaxPeersConnected {
			peers := tcpip.GetIPsfrombytes(txn[[2]byte{'P', 'P'}])
			peersConnected := tcpip.GetPeersConnected()
			for _, ip := range peers {
				if _, ok := peersConnected[ip]; ok {
					continue
				}
				if !tcpip.IsIPBanned(ip, h, tcpip.NonceTopic) {
					tcpip.AddNewPeer(ip, tcpip.NonceTopic)
				}
				if !tcpip.IsIPBanned(ip, h, tcpip.SyncTopic) {
					tcpip.AddNewPeer(ip, tcpip.SyncTopic)
				}
				if !tcpip.IsIPBanned(ip, h, tcpip.TransactionTopic) {
					tcpip.AddNewPeer(ip, tcpip.TransactionTopic)
				}
				if tcpip.GetPeersCount() > common.MaxPeersConnected {
					break
				}
			}
		}
		lastOtherHeight := common.GetInt64FromByte(txn[[2]byte{'L', 'H'}][0])
		common.SetHeightMax(lastOtherHeight)
		lastOtherBlockHashBytes := txn[[2]byte{'L', 'B'}][0]
		if lastOtherHeight == h {
			services.AdjustShiftInPastInReset(lastOtherHeight)
			lastBlockHashBytes, err := blocks.LoadHashOfBlock(h)
			if err != nil {
				panic(err)
			}
			if bytes.Compare(lastOtherBlockHashBytes, lastBlockHashBytes) != 0 {
				SendGetHeaders(addr, lastOtherHeight)
			}
			common.IsSyncing.Store(false)
			return
		} else if lastOtherHeight < h {
			services.AdjustShiftInPastInReset(lastOtherHeight)
			common.IsSyncing.Store(false)
			return
		}
		// when others have longer chain
		SendGetHeaders(addr, lastOtherHeight)
		return
	case "sh":

		txn := amsg.(message.TransactionsMessage).GetTransactionsBytes()
		blcks := []blocks.Block{}
		indices := []int64{}
		for k, tx := range txn {
			for _, t := range tx {
				if k == [2]byte{'I', 'H'} {
					index := common.GetInt64FromByte(t)
					indices = append(indices, index)
				} else if k == [2]byte{'H', 'V'} {
					block := blocks.Block{
						BaseBlock:          blocks.BaseBlock{},
						TransactionsHashes: nil,
						BlockHash:          common.Hash{},
					}
					block, err := block.GetFromBytes(t)
					if err != nil {
						panic("cannot unmarshal header")
					}
					blcks = append(blcks, block)
				}
			}
		}
		hmax := common.GetHeightMax()
		if indices[len(indices)-1] <= h {
			log.Println("shorter other chain")
			return
		}
		if indices[0] > h {
			log.Println("too far blocks of other")
			return
		}
		// check blocks
		was := false
		lastGoodBlock := indices[0]
		merkleTries := map[int64]*transactionsPool.MerkleTree{}
		for i := 0; i < len(blcks); i++ {
			header := blcks[i].GetHeader()
			index := indices[i]
			if index <= 0 {
				continue
			}
			block := blcks[i]
			oldBlock := blocks.Block{}
			if index <= h {
				hashOfMyBlockBytes, err := blocks.LoadHashOfBlock(index)
				if err != nil {
					panic("cannot load block hash")
				}
				if bytes.Compare(block.BlockHash.GetBytes(), hashOfMyBlockBytes) == 0 {
					lastGoodBlock = index
					continue
				}
			}
			if was == true {
				oldBlock = blcks[i-1]
			} else {
				oldBlock, err = blocks.LoadBlock(index - 1)
				if err != nil {
					panic("cannot load block")
				}
				was = true
			}

			if header.Height != index {
				common.ShiftToPastMutex.RLock()
				services.ResetAccountsAndBlocksSync(index - common.ShiftToPastInReset)
				common.ShiftToPastMutex.RUnlock()
				services.AdjustShiftInPastInReset(hmax)
				panic("not relevant height vs index")
			}

			merkleTrie, err := blocks.CheckBaseBlock(block, oldBlock)
			defer merkleTrie.Destroy()
			if err != nil {
				common.ShiftToPastMutex.RLock()
				services.ResetAccountsAndBlocksSync(index - common.ShiftToPastInReset)
				common.ShiftToPastMutex.RUnlock()
				services.AdjustShiftInPastInReset(hmax)
				panic(err)
			}
			merkleTries[index] = merkleTrie
			hashesMissing := blocks.IsAllTransactions(block)
			if len(hashesMissing) > 0 {
				transactionServices.SendGT(addr, hashesMissing)
			}
		}
		common.IsSyncing.Store(true)
		was = false
		for i := 0; i < len(blcks); i++ {
			block := blcks[i]
			index := indices[i]
			if block.GetHeader().Height <= lastGoodBlock {
				continue
			}
			oldBlock := blocks.Block{}
			if was == true {
				oldBlock = blcks[i-1]
			} else {
				oldBlock, err = blocks.LoadBlock(index - 1)
				if err != nil {
					panic("cannot load block")
				}
				was = true
			}
			err := blocks.CheckBlockAndTransferFunds(&block, oldBlock, merkleTries[index])
			if err != nil {
				log.Println(err)
				hashesMissing := blocks.IsAllTransactions(block)
				if len(hashesMissing) > 0 {
					transactionServices.SendGT(addr, hashesMissing)
				}
				services.RevertVMToBlockHeight(oldBlock.GetHeader().Height)
				return
			}
			// storing blocks
			err = block.StoreBlock()
			if err != nil {
				log.Println(err)
				panic("storing block failed")
			}
			common.SetHeight(block.GetHeader().Height)
			log.Println("New Block success -------------------------------------", block.GetHeader().Height)
			err = account.StoreAccounts(block.GetHeader().Height)
			if err != nil {
				log.Println(err)
			}
			err = account.StoreStakingAccounts(block.GetHeader().Height)
			if err != nil {
				log.Println(err)
			}
			statistics.UpdateStatistics(block, oldBlock)
		}

	case "gh":

		txn := amsg.(message.TransactionsMessage).GetTransactionsBytes()

		bHeight := common.GetInt64FromByte(txn[[2]byte{'B', 'H'}][0])
		eHeight := common.GetInt64FromByte(txn[[2]byte{'E', 'H'}][0])
		SendHeaders(addr, bHeight, eHeight)
	default:
	}
}
