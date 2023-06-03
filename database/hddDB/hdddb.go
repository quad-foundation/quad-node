package hddDatabase

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chainpqc/chainpqc-node/common"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"log"
	"os"
)

var blockchainDB *leveldb.DB

func Init() error {
	homePath, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	homePath += "/.chainpqc/db/blockchain"

	// Open or create a new LevelDB database
	blockchainDB, err = leveldb.OpenFile(homePath, nil)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func CloseDB() {
	blockchainDB.Close()
}

func Store(k []byte, v any) error {

	wm, err := common.Marshal(v, string(k[:2]))
	if err != nil {
		log.Println(err)
		return err
	}

	// Put a key-value pair into the database
	err = blockchainDB.Put(k, wm, nil)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func LoadAllKeys(prefix []byte) ([][]byte, error) {
	if len(prefix) == 0 {
		return nil, errors.New("prefix cannot be empty")
	}
	// Create a key range with the specified prefix
	keyRange := util.BytesPrefix(prefix)
	// Create an iterator with the key range
	iter := blockchainDB.NewIterator(keyRange, nil)
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

func LoadAll(prefix []byte) ([]interface{}, error) {
	iter := blockchainDB.NewIterator(util.BytesPrefix(prefix), nil)
	defer iter.Release()
	values := []interface{}{}
	for iter.Next() {
		v := interface{}(nil)
		err := common.Unmarshal(iter.Value(), string(prefix), &v)
		if err != nil {
			return nil, err
		}
		values = append(values, v)
	}
	err := iter.Error()
	if err != nil {
		return nil, err
	}
	return values, nil
}

func Load(k []byte, v interface{}) error {
	value, err := blockchainDB.Get(k, nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return fmt.Errorf("key not found")
		}
		log.Fatalf("Error getting value for key: %v", err)
	}
	err = json.Unmarshal(value, v)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func IsKey(key []byte) (bool, error) {
	_, err := blockchainDB.Get(key, nil)
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
func Delete(key []byte) error {
	return blockchainDB.Delete(key, nil)
}

func clearLevelDB() error {
	iter := blockchainDB.NewIterator(util.BytesPrefix([]byte("")), nil)
	defer iter.Release()
	batch := new(leveldb.Batch)
	for iter.Next() {
		key := iter.Key()
		batch.Delete(key)
	}
	err := blockchainDB.Write(batch, nil)
	if err != nil {
		return err
	}
	return nil
}

func LoadBytes(k []byte) ([]byte, error) {
	value, err := blockchainDB.Get(k, nil)
	if err != nil {
		return nil, err
	}
	return value, nil
}
