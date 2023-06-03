package hddDatabase

import (
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
