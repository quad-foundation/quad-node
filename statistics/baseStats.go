package statistics

import (
	"fmt"
	"github.com/quad/quad-node/blocks"
	"github.com/quad/quad-node/common"
	memDatabase "github.com/quad/quad-node/database"
	"github.com/quad/quad-node/transactionsDefinition"
	"github.com/quad/quad-node/transactionsPool"
	"log"
	"sync"
)

type MainStats struct {
	Heights                 int64   `json:"heights"`
	HeightMax               int64   `json:"heightMax"`
	TimeInterval            int64   `json:"timeInterval"`
	Transactions            int     `json:"transactions"`
	TransactionsPending     int     `json:"transactions_pending"`
	TransactionsSize        int     `json:"transaction_size"`
	TransactionsPendingSize int     `json:"transactions_pending_size"`
	Tps                     float32 `json:"tps"`
	Syncing                 bool    `json:"syncing"`
	Difficulty              int32   `json:"difficulty"`
	db                      memDatabase.AnyBlockchainDB
}

type GlobalMainStats struct {
	MainStats *MainStats
	Mutex     sync.Mutex
}

var globalMainStats *GlobalMainStats

var GmsMutex *GlobalMainStats

func InitGlobalMainStats() {
	GmsMutex = &GlobalMainStats{
		MainStats: nil,
		Mutex:     sync.Mutex{},
	}

	db := memDatabase.NewInMemoryDB()
	globalMainStats = &GlobalMainStats{
		MainStats: &MainStats{
			Heights:                 int64(0),
			HeightMax:               int64(0),
			TimeInterval:            int64(0),
			Transactions:            0,
			TransactionsSize:        0,
			TransactionsPending:     0,
			TransactionsPendingSize: 0,
			Tps:                     float32(0),
			Syncing:                 true,
			Difficulty:              int32(0),
			db:                      db,
		},
		Mutex: sync.Mutex{},
	}
}

func DestroyGlobalMainStats() {
	globalMainStats.Mutex.Lock()
	defer globalMainStats.Mutex.Unlock()
	globalMainStats.MainStats.Destroy()
}
func (ms *MainStats) Destroy() {
	ms.db = nil
}

func (ms *MainStats) SaveStats() error {
	msb, err := common.Marshal(*ms, common.StatDBPrefix)
	if err != nil {
		return err
	}
	err = ms.db.Put(common.StatDBPrefix[:], msb)
	if err != nil {
		return err
	}
	return nil
}

func LoadStats() (*GlobalMainStats, error) {
	if globalMainStats.Mutex.TryLock() {
		defer globalMainStats.Mutex.Unlock()
		if exist, _ := globalMainStats.MainStats.db.IsKey(common.StatDBPrefix[:]); !exist {
			err := globalMainStats.MainStats.SaveStats()
			if err != nil {
				log.Println("Can not initialize stats", err)
				return nil, err
			}
			return globalMainStats, nil
		}
		msb, err := globalMainStats.MainStats.db.Get(common.StatDBPrefix[:])
		if err != nil {
			return nil, err
		}
		err = common.Unmarshal(msb, common.StatDBPrefix, globalMainStats)
		if err != nil {
			return nil, err
		}
		globalMainStats.MainStats.Syncing = common.IsSyncing.Load()
		return globalMainStats, nil
	}
	return nil, fmt.Errorf("try Lock fails")
}

func UpdateStatistics(newBlock blocks.Block, merkleTrie *transactionsPool.MerkleTree, lastBlock blocks.Block) {
	if GmsMutex.Mutex.TryLock() {
		defer GmsMutex.Mutex.Unlock()
		stats, _ := LoadStats()
		stats.MainStats.Heights = common.GetHeight()
		stats.MainStats.HeightMax = common.GetHeightMax()
		stats.MainStats.Difficulty = newBlock.BaseBlock.BaseHeader.Difficulty
		stats.MainStats.Syncing = common.IsSyncing.Load()
		stats.MainStats.TimeInterval = newBlock.BaseBlock.BlockTimeStamp - lastBlock.BaseBlock.BlockTimeStamp
		empt := transactionsDefinition.EmptyTransaction()

		hs, _ := newBlock.GetTransactionsHashes(merkleTrie, newBlock.GetHeader().Height)
		stats.MainStats.Transactions = len(hs)
		stats.MainStats.TransactionsSize = len(hs) * len(empt.GetBytes())
		ntxs := len(hs)
		stats.MainStats.Tps = float32(ntxs) / float32(stats.MainStats.TimeInterval)

		nt := transactionsPool.PoolsTx.NumberOfTransactions()
		stats.MainStats.TransactionsPending = nt
		stats.MainStats.TransactionsPendingSize = nt * len(empt.GetBytes())

		err := stats.MainStats.SaveStats()
		if err != nil {
			log.Println(err)
		}
	}
}
