package common

import (
	"encoding/binary"
	"fmt"
	"github.com/quad/quad-node/crypto/blake2b"
	"log"
	"time"
)

func GetCurrentTimeStampInSecond() int64 {

	return time.Now().UTC().Unix()
}

func GetDelegatedAccountAddress(id int16) Address {
	a := Address{}
	b := make([]byte, 2)
	ba := make([]byte, a.GetLength()-2)
	binary.BigEndian.PutUint16(b, uint16(id))
	b = append(b, ba...)
	err := a.Init(b)
	if err != nil {
		panic(err)
	}
	return a
}

func CheckDelegatedAccountAddress(daddr Address) bool {

	n := GetInt16FromByte(daddr.GetBytes())
	return n > 0 && n < 256
}

func NumericalDelegatedAccountAddress(daddr Address) int16 {
	if CheckDelegatedAccountAddress(daddr) {
		n := GetInt16FromByte(daddr.GetBytes())
		return n
	}
	return 0
}

func Timer() func() float64 {
	start := time.Now()
	return func() float64 {
		return time.Since(start).Seconds()
	}
}

func CalcHashToByte(b []byte) ([]byte, error) {
	hashBlake2b, err := blake2b.New256(nil)
	if err != nil {
		return nil, err
	}
	hashBlake2b.Write(b)
	return hashBlake2b.Sum(nil), nil
}
func GetSignatureFromBytes(b []byte, address Address) (Signature, error) {
	s := Signature{}
	var err error
	err = s.Init(b, address)
	if err != nil {
		log.Println("Get Hash from bytes failed")
		return Signature{}, err
	}
	return s, nil
}

func GetHashFromBytes(b []byte) Hash {
	h := EmptyHash()
	(&h).Set(b)
	return h
}

func CalcHashFromBytes(b []byte) (Hash, error) {
	hb, err := CalcHashToByte(b)
	if err != nil {
		return Hash{}, err
	}
	h := GetHashFromBytes(hb)
	return h, nil
}

func ExtractKeys(m map[[2]byte][]byte) []string {
	keys := make([]string, 0, len(m))
	kb := make([]byte, 2)
	for k := range m {
		copy(kb, k[:])
		keys = append(keys, string(kb))
	}
	return keys
}

func ContainsKey(keys []string, searchKey string) bool {
	for _, key := range keys {
		if key == searchKey {
			return true
		}
	}
	return false
}

func ContainsKeyOfList(keys [][2]byte, searchKey [2]byte) bool {
	for _, key := range keys {
		if key == searchKey {
			return true
		}
	}
	return false
}

func ExtractKeysOfList(m map[[2]byte][][]byte) [][2]byte {
	keys := [][2]byte{}
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func IsInKeys2Byte(m map[[2]byte][]byte, searchKey string) bool {
	keys := ExtractKeys(m)
	return ContainsKey(keys, searchKey)
}

func IsInKeysOfList(m map[[2]byte][][]byte, searchKey [2]byte) bool {
	keys := ExtractKeysOfList(m)
	return ContainsKeyOfList(keys, searchKey)
}

func BytesToLenAndBytes(b []byte) []byte {
	lb := int32(len(b))
	bret := make([]byte, 4)
	binary.BigEndian.PutUint32(bret, uint32(lb))
	bret = append(bret, b...)
	return bret
}
func BytesWithLenToBytes(b []byte) ([]byte, []byte, error) {
	if len(b) < 4 {
		return nil, nil, fmt.Errorf("input byte slice is too short")
	}
	lb := int(binary.BigEndian.Uint32(b[:4]))
	if lb > len(b)-4 {
		return nil, nil, fmt.Errorf("length value in byte slice is incorrect")
	}
	return b[4 : 4+lb], b[4+lb:], nil
}
