package statistics

import (
	"fmt"
	"github.com/quad/quad-node/common"
	memDatabase "github.com/quad/quad-node/database"
	"log"
	"sync"
)

type MainStats struct {
	Heights                 int64         `json:"heights"`
	HeightMax               int64         `json:"heightMax"`
	Chain                   uint8         `json:"chain"`
	TimeInterval            int64         `json:"timeInterval"`
	Transactions            map[uint8]int `json:"transactions"`
	TransactionsPending     map[uint8]int `json:"transactions_pending"`
	TransactionsSize        map[uint8]int `json:"transaction_size"`
	TransactionsPendingSize map[uint8]int `json:"transactions_pending_size"`
	Tps                     float32       `json:"tps"`
	Syncing                 bool          `json:"syncing"`
	Difficulty              int32         `json:"difficulty"`
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
			Chain:                   uint8(0),
			TimeInterval:            int64(0),
			Transactions:            map[uint8]int{0: 0, 1: 0, 2: 0, 3: 0, 4: 0},
			TransactionsSize:        map[uint8]int{0: 0, 1: 0, 2: 0, 3: 0, 4: 0},
			TransactionsPending:     map[uint8]int{0: 0, 1: 0, 2: 0, 3: 0, 4: 0},
			TransactionsPendingSize: map[uint8]int{0: 0, 1: 0, 2: 0, 3: 0, 4: 0},
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
