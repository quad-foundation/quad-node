package statistics

import (
	"fmt"
	"github.com/quad-foundation/quad-node/blocks"
	"github.com/quad-foundation/quad-node/common"
	memDatabase "github.com/quad-foundation/quad-node/database"
	"github.com/quad-foundation/quad-node/transactionsDefinition"
	"github.com/quad-foundation/quad-node/transactionsPool"
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
	PriceOracle             float32 `json:"priceOracle"`
	RandOracle              int64   `json:"randOracle"`
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
			PriceOracle:             float32(1),
			RandOracle:              int64(0),
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

func UpdateStatistics(newBlock blocks.Block, lastBlock blocks.Block) {
	if GmsMutex.Mutex.TryLock() {
		defer GmsMutex.Mutex.Unlock()
		stats, _ := LoadStats()
		stats.MainStats.Heights = common.GetHeight()
		stats.MainStats.HeightMax = common.GetHeightMax()
		stats.MainStats.Difficulty = newBlock.BaseBlock.BaseHeader.Difficulty
		stats.MainStats.PriceOracle = float32(newBlock.BaseBlock.PriceOracle) / 100000000.0
		stats.MainStats.RandOracle = newBlock.BaseBlock.RandOracle
		stats.MainStats.Syncing = common.IsSyncing.Load()
		stats.MainStats.TimeInterval = newBlock.BaseBlock.BlockTimeStamp - lastBlock.BaseBlock.BlockTimeStamp
		empt := transactionsDefinition.EmptyTransaction()

		hs, _ := newBlock.GetTransactionsHashes(newBlock.GetHeader().Height)
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
