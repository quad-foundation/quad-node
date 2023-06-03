package hddDatabase

import (
	"bytes"
	"testing"
)

func TestInit(t *testing.T) {
	err := Init()
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	defer CloseDB()
	if blockchainDB == nil {
		t.Fatal("blockchainDB not initialized")
	}
}
func TestStoreAndLoad(t *testing.T) {
	err := Init()
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	defer CloseDB()
	key := []byte("01testKey")
	value := "testValue"
	err = Store(key, value)
	if err != nil {
		t.Fatalf("Store() failed: %v", err)
	}
	var loadedValue string
	err = Load(key, &loadedValue)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if loadedValue != value {
		t.Fatalf("Loaded value does not match stored value: expected %v, got %v", value, loadedValue)
	}
}
func TestIsKey(t *testing.T) {
	err := Init()
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	defer CloseDB()
	key := []byte("01testKey")
	value := "testValue"
	err = Store(key, value)
	if err != nil {
		t.Fatalf("Store() failed: %v", err)
	}
	exists, err := IsKey(key)
	if err != nil {
		t.Fatalf("IsKey() failed: %v", err)
	}
	if !exists {
		t.Fatal("IsKey() returned false for an existing key")
	}
	nonExistentKey := []byte("01nonExistentKey")
	exists, err = IsKey(nonExistentKey)
	if err != nil {
		t.Fatalf("IsKey() failed: %v", err)
	}
	if exists {
		t.Fatal("IsKey() returned true for a non-existent key")
	}
}
func TestDelete(t *testing.T) {
	err := Init()
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	defer CloseDB()
	key := []byte("01testKey")
	value := "testValue"
	err = Store(key, value)
	if err != nil {
		t.Fatalf("Store() failed: %v", err)
	}
	err = Delete(key)
	if err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}
	exists, err := IsKey(key)
	if err != nil {
		t.Fatalf("IsKey() failed: %v", err)
	}
	if exists {
		t.Fatal("Key still exists after Delete()")
	}
}

func TestLoadAllKeys(t *testing.T) {
	err := Init()
	if err != nil {
		t.Fatal(err)
	}
	err = clearLevelDB()
	if err != nil {
		t.Fatal(err)
	}
	defer CloseDB()
	// Test data
	key1 := []byte("k1")
	value1 := "value1"
	key2 := []byte("k2234")
	value2 := "value2"
	// Store test data
	err = Store(key1, value1)
	if err != nil {
		t.Fatal(err)
	}
	err = Store(key2, value2)
	if err != nil {
		t.Fatal(err)
	}
	// Call LoadAllKeys with the appropriate prefix
	keys, err := LoadAllKeys([]byte("k"))
	if err != nil {
		t.Fatal(err)
	}
	// Check if the keys are correct
	if len(keys) != 2 {
		t.Fatalf("Expected 2 keys, got %d", len(keys))
	}
	if !bytes.Equal(keys[0], key1) || !bytes.Equal(keys[1], key2) {
		t.Fatalf("Keys do not match expected values")
	}
}
func TestLoadAll(t *testing.T) {
	err := Init()
	if err != nil {
		t.Fatal(err)
	}
	err = clearLevelDB()
	if err != nil {
		t.Fatal(err)
	}
	defer CloseDB()
	// Test data
	key1 := []byte("k1")
	value1 := "value1"
	key2 := []byte("k2")
	value2 := "value2"
	// Store test data
	err = Store(key1, value1)
	if err != nil {
		t.Fatal(err)
	}
	err = Store(key2, value2)
	if err != nil {
		t.Fatal(err)
	}
	// Call LoadAll with the appropriate prefix
	values, err := LoadAll([]byte("k"))
	if err != nil {
		t.Fatal(err)
	}
	// Check if the values are correct
	if len(values) != 2 {
		t.Fatalf("Expected 2 values, got %d", len(values))
	}
	if values[0] != value1 || values[1] != value2 {
		t.Fatalf("Values do not match expected values")
	}
}
