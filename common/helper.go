package common

import (
	"encoding/binary"
	"github.com/chainpqc/chainpqc-node/crypto/blake2b"
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

func GetHashFromBytes(b []byte) (Hash, error) {
	h := Hash{}
	hb, err := CalcHashToByte(b)
	if err != nil {
		return Hash{}, err
	}
	h, err = h.Init(hb)
	if err != nil {
		log.Println("Get Hash from bytes failed")
		return Hash{}, err
	}
	return h, nil
}
