package services

import (
	"github.com/quad/quad-node/account"
	"github.com/quad/quad-node/blocks"
	"github.com/quad/quad-node/common"
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

func ResetAccountsAndBlocksSync(height int64) {
	if height < 0 {
		log.Println("try to reset from negative height")
		height = 0
	}

	err := account.LoadAccounts(height)
	if err != nil {
		return
	}
	ha, err := account.LastHeightStoredInAccounts()
	if err != nil {
		log.Println(err)
	}
	hb, err := blocks.LastHeightStoredInBlocks()
	if err != nil {
		log.Println(err)
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
	common.SetHeight(height)
	common.IsSyncing.Store(true)
}
