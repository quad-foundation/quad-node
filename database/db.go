package database

import (
	"github.com/chainpqc/chainpqc-node/common"
	"github.com/tecbot/gorocksdb"
	"log"
	"os"
)

var blockchainDB *gorocksdb.DB

func Init() error {
	homePath, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	homePath += "/.chainpqc/db/blockchain"

	// Create a new RocksDB options object
	opts := gorocksdb.NewDefaultOptions()
	opts.SetCreateIfMissing(true)
	// Open the database with the provided options
	blockchainDB, err = gorocksdb.OpenDb(opts, homePath)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
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

	// Create a new write options object
	wo := gorocksdb.NewDefaultWriteOptions()
	defer wo.Destroy()
	// Put a key-value pair into the database
	err = blockchainDB.Put(wo, k, wm)
	if err != nil {
		log.Fatalf("Error putting key-value pair: %v", err)
		return err
	}

	return nil
}

func LoadAllKeys(k []byte) [][]byte {
	prefix := k[:2]

	// Create a read options with a custom prefix extractor
	ro := gorocksdb.NewDefaultReadOptions()
	ro.SetIterateUpperBound(prefix)
	// Create an iterator with the read options
	iter := blockchainDB.NewIterator(ro)
	defer iter.Close()

	keys := [][]byte{}
	// Iterate over the keys with the specified prefix
	for iter.Seek(prefix); iter.ValidForPrefix(prefix); iter.Next() {
		keys = append(keys, iter.Key().Data())
	}
	return keys
}

func LoadAll(k []byte) ([]any, error) {

	prefix := k[:2]

	// Create a read options with a custom prefix extractor
	ro := gorocksdb.NewDefaultReadOptions()
	ro.SetIterateUpperBound(prefix)
	// Create an iterator with the read options
	iter := blockchainDB.NewIterator(ro)
	defer iter.Close()

	values := []interface{}{}
	// Iterate over the keys with the specified prefix
	for iter.Seek(prefix); iter.ValidForPrefix(prefix); iter.Next() {
		//keys = append(keys, iter.Key().Data())
		v := interface{}(nil)
		err := common.Unmarshal(iter.Value().Data(), string(prefix), v)
		if err != nil {
			return nil, err
		}
		values = append(values, v)
	}
	return values, nil
}

func Load(k []byte, v any) error {

	ro := gorocksdb.NewDefaultReadOptions()
	defer ro.Destroy()
	// Get the value for the given key
	value, err := blockchainDB.Get(ro, k)
	if err != nil {
		log.Fatalf("Error getting value for key: %v", err)
	}
	defer value.Free()

	err = common.Unmarshal(value.Data(), string(k[:2]), v)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func IsKey(key []byte) (bool, error) {
	ro := gorocksdb.NewDefaultReadOptions()
	defer ro.Destroy()
	value, err := blockchainDB.Get(ro, key)
	if err != nil {
		return false, err
	}
	defer value.Free()
	return value.Exists(), nil
}

func Delete(key []byte) error {
	wo := gorocksdb.NewDefaultWriteOptions()
	defer wo.Destroy()
	return blockchainDB.Delete(wo, key)
}

func LoadBytes(k []byte) ([]byte, error) {

	ro := gorocksdb.NewDefaultReadOptions()
	defer ro.Destroy()
	// Get the value for the given key
	value, err := blockchainDB.Get(ro, k)
	if err != nil {
		return nil, err
	}
	defer value.Free()

	return value.Data(), nil
}
