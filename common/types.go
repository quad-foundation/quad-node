package common

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/quad-foundation/quad-node/crypto/blake2b"
	"log"
)

const (
	AddressLength    int    = 20
	PubKeyLength     int    = 897  //264608
	PrivateKeyLength int    = 1281 //64
	HashLength       int    = 32
	ShortHashLength  int    = 8
	SignatureLength  int    = 666          // 164
	sigName          string = "Falcon-512" //"Rainbow-III-Compressed" //
)

func GetSigName() string {
	return sigName
}

func (a PubKey) GetLength() int {
	return PubKeyLength
}

func (p PrivKey) GetLength() int {
	return PrivateKeyLength
}

func (s Signature) GetLength() int {
	return len(s.ByteValue)
}

func (a Address) GetLength() int {
	return AddressLength
}

func (a Hash) GetLength() int {
	return HashLength
}

func (a ShortHash) GetLength() int {
	return ShortHashLength
}

type Address struct {
	ByteValue [AddressLength]byte `json:"byte_value"`
}

func (a *Address) Init(b []byte) error {
	if len(b) != a.GetLength() {
		return fmt.Errorf("error Address initialization with wrong length, should be %v", a.GetLength())
	}
	copy(a.ByteValue[:], b[:])
	return nil
}

func BytesToAddress(b []byte) (Address, error) {
	var a Address
	err := a.Init(b)
	if err != nil {
		log.Println("Cannot init Address")
		return a, err
	}
	return a, nil
}

func PubKeyToAddress(p PubKey) (Address, error) {
	hashBlake2b, err := blake2b.New160(nil)
	if err != nil {
		return Address{}, err
	}
	hashBlake2b.Write(p.GetBytes())
	return BytesToAddress(hashBlake2b.Sum(nil))
}

func (a *Address) GetBytes() []byte {
	return a.ByteValue[:]
}

func (a *Address) GetHex() string {
	return hex.EncodeToString(a.GetBytes())
}

type PubKey struct {
	ByteValue []byte  `json:"byte_value"`
	Address   Address `json:"address"`
}

func (pk *PubKey) Init(b []byte) error {
	if len(b) != pk.GetLength() {
		return fmt.Errorf("error Pubkey initialization with wrong length, should be %v", pk.GetLength())
	}
	pk.ByteValue = b[:]
	addr, err := PubKeyToAddress(*pk)
	if err != nil {
		return err
	}
	pk.Address = addr
	return nil
}

func (pk PubKey) GetBytes() []byte {
	return pk.ByteValue[:]
}

func (pk PubKey) GetHex() string {
	return hex.EncodeToString(pk.GetBytes())
}

func (pk PubKey) GetAddress() Address {
	return pk.Address
}

type PrivKey struct {
	ByteValue []byte  `json:"byte_value"`
	Address   Address `json:"address"`
}

func (pk *PrivKey) Init(b []byte, address Address) error {
	if len(b) != pk.GetLength() {
		return fmt.Errorf("error Private key initialization with wrong length, should be %v", pk.GetLength())
	}
	pk.ByteValue = b[:]
	pk.Address = address
	return nil
}

func (pk PrivKey) GetBytes() []byte {
	return pk.ByteValue[:]
}

func (pk PrivKey) GetHex() string {
	return hex.EncodeToString(pk.GetBytes())
}

func (pk PrivKey) GetAddress() Address {
	return pk.Address
}

type Signature struct {
	ByteValue []byte  `json:"byte_value"`
	Address   Address `json:"address"`
}

func (s *Signature) Init(b []byte, address Address) error {
	if len(b) > SignatureLength {
		return fmt.Errorf("error Signature initialization with wrong length, should be %v", s.GetLength())
	}
	s.ByteValue = b[:]
	s.Address = address
	return nil
}

func (s Signature) GetBytes() []byte {
	return s.ByteValue
}

func (s Signature) GetHex() string {
	return hex.EncodeToString(s.GetBytes())
}

func (s Signature) GetAddress() Address {
	return s.Address
}

type Hash [HashLength]byte
type ShortHash [ShortHashLength]byte

func (h *Hash) Set(b []byte) {
	copy(h[:], b[:])
}

func (h Hash) GetBytes() []byte {
	return h[:]
}

func (h Hash) GetHex() string {
	return hex.EncodeToString(h.GetBytes())
}
func (h *ShortHash) Set(b []byte) {
	copy(h[:], b[:])
}
func (h ShortHash) GetBytes() []byte {
	return h[:]
}

func (h ShortHash) GetHex() string {
	return hex.EncodeToString(h.GetBytes())
}

// GetByteInt32 converts an int32 value to a byte slice.
func GetByteInt32(u int32) []byte {
	tmp := make([]byte, 4)
	binary.LittleEndian.PutUint32(tmp, uint32(u))
	return tmp
}

// GetByteInt16 converts an int16 value to a byte slice.
func GetByteInt16(u int16) []byte {
	tmp := make([]byte, 2)
	binary.LittleEndian.PutUint16(tmp, uint16(u))
	return tmp
}

// GetByteInt64 converts an int64 value to a byte slice.
func GetByteInt64(u int64) []byte {
	tmp := make([]byte, 8)
	binary.LittleEndian.PutUint64(tmp, uint64(u))
	return tmp
}

// GetInt64FromByte converts a byte slice to an int64 value.
func GetInt64FromByte(bs []byte) int64 {
	return int64(binary.LittleEndian.Uint64(bs))
}

// GetInt32FromByte converts a byte slice to an int32 value.
func GetInt32FromByte(bs []byte) int32 {
	return int32(binary.LittleEndian.Uint32(bs))
}

// GetInt16FromByte converts a byte slice to an int16 value.
func GetInt16FromByte(bs []byte) int16 {
	return int16(binary.LittleEndian.Uint16(bs))
}

func EmptyHash() Hash {
	tmp := make([]byte, 32)
	h := Hash{}
	(&h).Set(tmp)
	return h
}

func EmptyAddress() Address {
	a := Address{}
	tmp := make([]byte, a.GetLength())
	err := a.Init(tmp)
	if err != nil {
		return Address{}
	}
	return a
}

func EmptySignature() Signature {
	s := Signature{}
	tmp := make([]byte, s.GetLength())
	s.Init(tmp, EmptyAddress())
	return s
}
