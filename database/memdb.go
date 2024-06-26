package memDatabase

import (
	"errors"
	"fmt"
	"github.com/quad-foundation/quad-node/common"
	commoneth "github.com/quad-foundation/quad-node/common"
	"github.com/quad-foundation/quad-node/wallet"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"github.com/syndtr/goleveldb/leveldb/util"
	"log"
	"sync"
)

type BlockchainDB struct {
	ldb   *leveldb.DB
	mutex sync.RWMutex
}

func (db *BlockchainDB) GetLdb() *leveldb.DB {
	return db.ldb
}

func (db *BlockchainDB) InitPermanent() (*BlockchainDB, error) {
	var err error
	db.mutex.Lock()
	defer db.mutex.Unlock()
	// in memery db only
	memStorage := storage.NewMemStorage()
	db.ldb, err = leveldb.Open(memStorage, nil)
	if err != nil {
		log.Fatal(err)
	}
	w := wallet.GetActiveWallet()
	err = db.Put(append(common.PubKeyDBPrefix[:], w.Address.GetBytes()...),
		w.PublicKey.GetBytes())
	return db, nil
}

func (db *BlockchainDB) InitInMemory() (*BlockchainDB, error) {
	var err error
	// in memery db only
	memStorage := storage.NewMemStorage()
	db.ldb, err = leveldb.Open(memStorage, nil)
	if err != nil {
		log.Fatal(err)
	}
	return db, nil
}

func (db *BlockchainDB) GetNode(hash commoneth.Hash) ([]byte, error) {
	return db.Get(hash[:])
}

func (d *BlockchainDB) Close() {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.ldb.Close()
}

func (db *BlockchainDB) Put(k []byte, v []byte) error {
	if len(k) == 0 {
		return errors.New("key cannot be empty")
	}
	db.mutex.Lock()
	defer db.mutex.Unlock()
	// Put a key-value pair into the database
	err := db.ldb.Put(k, v, nil)
	if err != nil {
		return err
	}
	return nil
}

func (db *BlockchainDB) LoadAllKeys(prefix []byte) ([][]byte, error) {
	if len(prefix) == 0 {
		return nil, errors.New("prefix cannot be empty")
	}
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	// Create a key range with the specified prefix
	keyRange := util.BytesPrefix(prefix)
	// Create an iterator with the key range
	iter := db.ldb.NewIterator(keyRange, nil)
	defer iter.Release()
	keys := [][]byte{}
	// Iterate over the keys with the specified prefix
	for iter.Next() {
		key := make([]byte, len(iter.Key()))
		copy(key, iter.Key())
		keys = append(keys, key)
	}
	return keys, iter.Error()
}

func (db *BlockchainDB) LoadAll(prefix []byte) ([][]byte, error) {
	if len(prefix) == 0 {
		return nil, errors.New("prefix cannot be empty")
	}
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	iter := db.ldb.NewIterator(util.BytesPrefix(prefix), nil)
	defer iter.Release()
	prefix2 := [2]byte{}
	copy(prefix2[:], prefix[:2])
	values := [][]byte{}
	for iter.Next() {
		values = append(values, iter.Value())
	}
	err := iter.Error()
	if err != nil {
		return nil, err
	}
	return values, nil
}

func (db *BlockchainDB) Get(k []byte) ([]byte, error) {
	if len(k) == 0 {
		return nil, errors.New("key cannot be empty")
	}
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	value, err := db.ldb.Get(k, nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return []byte{}, fmt.Errorf("key not found %s", k)
		}
		return []byte{}, fmt.Errorf("Error getting value for key: %v, key %s", err, k)
	}
	return value, nil
}

func (db *BlockchainDB) IsKey(key []byte) (bool, error) {
	if len(key) == 0 {
		return false, errors.New("key cannot be empty")
	}
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	_, err := db.ldb.Get(key, nil)
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return false, nil
		}
		// Optionally print the stack trace if in debug mode
		// if debugMode {
		//     debug.PrintStack()
		// }
		return false, err
	}
	return true, nil
}

func (db *BlockchainDB) Delete(key []byte) error {
	if len(key) == 0 {
		return errors.New("key cannot be empty")
	}
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	return db.ldb.Delete(key, nil)
}
