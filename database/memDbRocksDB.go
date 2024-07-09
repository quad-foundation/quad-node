package memDatabase

import (
	"errors"
	"fmt"
	"github.com/quad-foundation/quad-node/common"
	commoneth "github.com/quad-foundation/quad-node/common"
	"github.com/quad-foundation/quad-node/wallet"
	"github.com/tecbot/gorocksdb"
	"log"
	"sync"
)

type BlockchainDB struct {
	db    *gorocksdb.DB
	mutex sync.RWMutex
}

func (db *BlockchainDB) GetLdb() *gorocksdb.DB {
	return db.db
}

func (db *BlockchainDB) InitPermanent() (*BlockchainDB, error) {
	var err error
	db.mutex.Lock()
	defer db.mutex.Unlock()
	opts := gorocksdb.NewDefaultOptions()
	opts.SetCreateIfMissing(true)
	db.db, err = gorocksdb.OpenDb(opts, common.DefaultBlockchainHomePath)
	if err != nil {
		log.Fatal(err)
	}
	w := wallet.GetActiveWallet()
	err = db.Put(append(common.PubKeyDBPrefix[:], w.Address.GetBytes()...), w.PublicKey.GetBytes())
	return db, nil
}
func (db *BlockchainDB) InitInMemory() (*BlockchainDB, error) {
	var err error
	db.mutex.Lock()
	defer db.mutex.Unlock()
	opts := gorocksdb.NewDefaultOptions()
	opts.SetCreateIfMissing(true)
	opts.SetEnv(gorocksdb.NewMemEnv())
	db.db, err = gorocksdb.OpenDb(opts, "quad")
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
	d.db.Close()
}
func (db *BlockchainDB) Put(k []byte, v []byte) error {
	if len(k) == 0 {
		return errors.New("key cannot be empty")
	}
	db.mutex.Lock()
	defer db.mutex.Unlock()
	wo := gorocksdb.NewDefaultWriteOptions()
	defer wo.Destroy()
	err := db.db.Put(wo, k, v)
	return err
}

func (db *BlockchainDB) LoadAllKeys(prefix []byte) ([][]byte, error) {
	if len(prefix) == 0 {
		return nil, errors.New("prefix cannot be empty")
	}
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	ro := gorocksdb.NewDefaultReadOptions()
	defer ro.Destroy()
	iter := db.db.NewIterator(ro)
	defer iter.Close()
	keys := [][]byte{}
	for iter.Seek(prefix); iter.ValidForPrefix(prefix); iter.Next() {
		key := make([]byte, len(iter.Key().Data()))
		copy(key, iter.Key().Data())
		keys = append(keys, key)
	}
	return keys, iter.Err()
}
func (db *BlockchainDB) LoadAll(prefix []byte) ([][]byte, error) {
	if len(prefix) == 0 {
		return nil, errors.New("prefix cannot be empty")
	}
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	ro := gorocksdb.NewDefaultReadOptions()
	defer ro.Destroy()
	iter := db.db.NewIterator(ro)
	defer iter.Close()
	values := [][]byte{}
	for iter.Seek(prefix); iter.ValidForPrefix(prefix); iter.Next() {
		values = append(values, iter.Value().Data())
	}
	return values, iter.Err()
}
func (db *BlockchainDB) Get(k []byte) ([]byte, error) {
	if len(k) == 0 {
		return nil, errors.New("key cannot be empty")
	}
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	ro := gorocksdb.NewDefaultReadOptions()
	defer ro.Destroy()
	value, err := db.db.Get(ro, k)
	if err != nil {
		return []byte{}, fmt.Errorf("Error getting value for key: %v, key %s", err, k)
	}
	return value.Data(), nil
}

func (db *BlockchainDB) IsKey(key []byte) (bool, error) {
	if len(key) == 0 {
		return false, errors.New("key cannot be empty")
	}
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	ro := gorocksdb.NewDefaultReadOptions()
	defer ro.Destroy()
	value, err := db.db.Get(ro, key)
	if err != nil {
		return false, err
	}
	defer value.Free()
	return value.Exists(), nil
}

func (db *BlockchainDB) Delete(key []byte) error {
	if len(key) == 0 {
		return errors.New("key cannot be empty")
	}
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	wo := gorocksdb.NewDefaultWriteOptions()
	defer wo.Destroy()
	return db.db.Delete(wo, key)
}
