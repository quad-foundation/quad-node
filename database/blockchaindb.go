package memDatabase

import (
	"github.com/quad-foundation/quad-node/common"
	"github.com/tecbot/gorocksdb"
	"sync"
)

var MainDB *BlockchainDB

type AnyBlockchainDB interface {
	Close()
	Put(k []byte, v []byte) error
	LoadAllKeys(prefix []byte) ([][]byte, error)
	LoadAll(prefix []byte) ([][]byte, error)
	Get(k []byte) ([]byte, error)
	IsKey(key []byte) (bool, error)
	Delete(key []byte) error
	GetLdb() *gorocksdb.DB
	GetNode(common.Hash) ([]byte, error)
}

func Init() {
	db := &BlockchainDB{}
	pdb, err := db.InitInMemory() // should be changed to permanent
	if err != nil {
		return
	}
	MainDB = pdb
}

func CloseDB() {
	MainDB.mutex.Lock()
	defer MainDB.mutex.Unlock()
	(*MainDB).Close()
}

type InMemoryDBReader struct {
	db AnyBlockchainDB
}

func (r *InMemoryDBReader) Node(owner common.Hash, path []byte, hash common.Hash) ([]byte, error) {
	key := append(owner.Bytes(), path...)
	key = append(key, hash.Bytes()...)
	value, err := r.db.Get(key)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (r *InMemoryDBReader) Get(key []byte) ([]byte, error) {
	value, err := r.db.Get(key)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func GetDBPermanentInstance() AnyBlockchainDB {
	return &BlockchainDB{
		db:    (*MainDB).GetLdb(),
		mutex: sync.RWMutex{},
	}
}

func NewInMemoryDB() AnyBlockchainDB {
	db := BlockchainDB{}
	db.mutex = sync.RWMutex{}
	memory, err := db.InitInMemory()
	if err != nil {
		return nil
	}
	return &BlockchainDB{
		db:    memory.GetLdb(),
		mutex: sync.RWMutex{},
	}
}
