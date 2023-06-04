package memDatabase

import (
	"errors"
	"fmt"
	"github.com/chainpqc/chainpqc-node/common"
	"github.com/chainpqc/chainpqc-node/wallet"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
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

	// in memery DB only
	memStorage := storage.NewMemStorage()
	blockchainDB, err = leveldb.Open(memStorage, nil)
	if err != nil {
		log.Fatal(err)
	}
	w := wallet.EmptyWallet().GetWallet()
	err = Store(append([]byte(common.PubKeyDBPrefix), w.Address.GetBytes()...),
		w.PublicKey.GetBytes())
	return nil
}

func CloseDB() {
	blockchainDB.Close()
}

func Store(k []byte, v []byte) error {

	//prefix := [2]byte{}
	//copy(prefix[:], k[:2])
	//wm, err := common.Marshal(v, prefix)
	//if err != nil {
	//	log.Println(err)
	//	return err
	//}

	// Put a key-value pair into the database
	err := blockchainDB.Put(k, v, nil)
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

func LoadAll(prefix []byte) ([][]byte, error) {
	iter := blockchainDB.NewIterator(util.BytesPrefix(prefix), nil)
	defer iter.Release()
	prefix2 := [2]byte{}
	copy(prefix2[:], prefix[:2])
	values := [][]byte{}
	for iter.Next() {
		//v := interface{}(nil)
		//err := common.Unmarshal(iter.Value(), prefix2, &v)
		//if err != nil {
		//	return nil, err
		//}
		values = append(values, iter.Value())
	}
	err := iter.Error()
	if err != nil {
		return nil, err
	}
	return values, nil
}

func Load(k []byte) ([]byte, error) {
	value, err := blockchainDB.Get(k, nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return []byte{}, fmt.Errorf("key not found")
		}
		log.Fatalf("Error getting value for key: %v", err)
	}
	//err = json.Unmarshal(value, v)
	//if err != nil {
	//	log.Println(err)
	//	return err
	//}
	return value, nil
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

//func LoadBytes(k []byte) ([]byte, error) {
//	value, err := blockchainDB.Get(k, nil)
//	if err != nil {
//		return nil, err
//	}
//	return value, nil
//}
//
//func LoadPubKey(addr common.Address) (pk common.PubKey, err error) {
//	val, err := LoadBytes(append([]byte(common.PubKeyDBPrefix), addr.GetBytes()...))
//	if err != nil {
//		return pk, err
//	}
//	err = pk.Init(val)
//	if err != nil {
//		return pk, err
//	}
//	return pk, nil
//}
