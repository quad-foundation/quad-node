package syncServices

import (
	"bytes"
	"github.com/quad/quad-node/blocks"
	"github.com/quad/quad-node/common"
	"github.com/quad/quad-node/message"
	"github.com/quad/quad-node/transactionsPool"
	"log"
)

func OnMessage(addr string, m []byte) {

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
		if amsg.GetChain() != 255 {
			panic("not adequate chain for hi")
		}
		txn := amsg.(message.TransactionsMessage).GetTransactionsBytes()
		h := common.GetHeight()
		lastOtherHeight := common.GetInt64FromByte(txn[[2]byte{'L', 'H'}][0])
		lastOtherBlockHashBytes := txn[[2]byte{'L', 'B'}][0]
		if lastOtherHeight == h {
			//lastBlock, err := blocks.LoadBlock(h)
			//if err != nil {
			//	panic(err)
			//}
			lastBlockHashBytes, err := blocks.LoadHashOfBlock(h)
			if err != nil {
				panic(err)
			}
			if bytes.Compare(lastOtherBlockHashBytes, lastBlockHashBytes) != 0 {
				SendGetHeaders(addr, lastOtherHeight)
			}
			return
		} else if lastOtherHeight < h {
			//bHeight := h - common.NumberOfHashesInBucket
			//if bHeight <= 0 {
			//	bHeight = 0
			//}
			//SendHeaders(addr, bHeight, h)
			return
		}
		// when others have longer chain
		SendGetHeaders(addr, lastOtherHeight)
		return
	case "sh":
		if amsg.GetChain() != 255 {
			panic("not adequate chain for hi")
		}
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
						Chain:              0,
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
		h := common.GetHeight()
		if indices[len(indices)-1] <= h {
			log.Println("shorter other chain")
			return
		}
		if indices[0] > h {
			log.Println("too far blocks of other")
			return
		}
		// check blocks
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
			if index > h {
				hashOfMyBlockBytes, err := blocks.LoadHashOfBlock(index)
				if err != nil {
					panic("cannot load block hash")
				}
				if bytes.Compare(block.BlockHash.GetBytes(), hashOfMyBlockBytes) == 0 {
					lastGoodBlock = index
					continue
				}
			}
			if i > 0 {
				oldBlock = blcks[i-1]
			} else {
				oldBlock, err = blocks.LoadBlock(index - 1)
				if err != nil {
					panic("cannot load block")
				}
			}

			if header.Height != index {
				panic("not relevant height vs index")
			}
			if !common.CheckHeight(block.Chain, index) {
				panic("chain improper related to height")
			}
			merkleTrie, err := blocks.CheckBaseBlock(block, oldBlock)
			defer merkleTrie.Destroy()
			if err != nil {
				panic(err)
			}
			merkleTries[index] = merkleTrie
		}
		for i := 0; i < len(blcks); i++ {
			block := blcks[i]
			index := indices[i]
			if block.GetHeader().Height <= lastGoodBlock {
				continue
			}
			oldBlock := blocks.Block{}
			if i > 0 {
				oldBlock = blcks[i-1]
			} else {
				oldBlock, err = blocks.LoadBlock(index - 1)
				if err != nil {
					panic("cannot load block")
				}
			}
			err := blocks.CheckBlockAndTransferFunds(block, oldBlock, merkleTries[index])
			if err != nil {
				return
			}
			// storing blocks
			err = block.StoreBlock()
			if err != nil {
				log.Println(err)
				panic("storing block failed")
			}
			common.SetHeight(block.GetHeader().Height)
			log.Println("New Block success -------------------------------------", block.GetHeader().Height, "-------chain", block.Chain)
		}

	case "gh":
		if amsg.GetChain() != 255 {
			panic("not adequate chain for hi")
		}
		txn := amsg.(message.TransactionsMessage).GetTransactionsBytes()

		bHeight := common.GetInt64FromByte(txn[[2]byte{'B', 'H'}][0])
		eHeight := common.GetInt64FromByte(txn[[2]byte{'E', 'H'}][0])
		SendHeaders(addr, bHeight, eHeight)
	default:
	}
}
