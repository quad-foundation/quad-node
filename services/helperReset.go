package services

import (
	"github.com/quad-foundation/quad-node/account"
	"github.com/quad-foundation/quad-node/blocks"
	"github.com/quad-foundation/quad-node/common"
	"github.com/quad-foundation/quad-node/transactionsPool"
	"log"
)

func AdjustShiftInPastInReset(height int64) {
	h := common.GetHeight()
	if height-h <= 0 {
		common.ShiftToPastInReset = 1
		return
	}
	common.ShiftToPastInReset *= 2
	if common.ShiftToPastInReset > h {
		common.ShiftToPastInReset = h
	}
}

func RevertVMToBlockHeight(height int64) bool {
	blocks.StateMutex.Lock()
	defer blocks.StateMutex.Unlock()
	lastNum := 0
	for h, n := range blocks.State.HeightToSnapShotNum {
		if h > height {
			continue
		}
		if n > lastNum {
			lastNum = n
		}
	}
	_, ok := blocks.State.GetSnapShotNum(height)
	if ok {
		blocks.State.RevertToSnapshot(lastNum)
	}
	return true
}

func ResetAccountsAndBlocksSync(height int64) {
	if height < 0 {
		log.Println("try to reset from negative height")
		height = 0
	}

	err := account.LoadAccounts(height)
	if err != nil {
		return
	}
	err = account.LoadStakingAccounts(height)
	if err != nil {
		return
	}
	ha, err := account.LastHeightStoredInAccounts()
	if err != nil {
		log.Println(err)
	}
	hsa, err := account.LastHeightStoredInStakingAccounts()
	if err != nil {
		log.Println(err)
	}
	hb, err := blocks.LastHeightStoredInBlocks()
	if err != nil {
		log.Println(err)
	}
	hd, err := account.LastHeightStoredInDexAccounts()
	if err != nil {
		log.Println(err)
	}

	if RevertVMToBlockHeight(height) == false {
		log.Println("reverting VM to height ", height, " fails.")
	}

	for i := hb; i > height; i-- {
		err := blocks.RemoveBlockFromDB(i)
		if err != nil {
			log.Println(err)
		}
	}
	for i := ha; i > height; i-- {
		err := account.RemoveAccountsFromDB(i)
		if err != nil {
			log.Println(err)
		}
	}
	for i := hsa; i > height; i-- {
		err := account.RemoveStakingAccountsFromDB(i)
		if err != nil {
			log.Println(err)
		}
	}
	for i := hd; i > height; i-- {
		err := account.RemoveDexAccountsFromDB(i)
		if err != nil {
			log.Println(err)
		}
	}

	hm, err := transactionsPool.LastHeightStoredInMerleTrie()
	if err != nil {
		log.Println(err)
	}
	for i := hm; i > height; i-- {
		err := transactionsPool.RemoveMerkleTrieFromDB(i)
		if err != nil {
			log.Println(err)
		}
	}
	common.SetHeight(height)
	common.IsSyncing.Store(true)
}
