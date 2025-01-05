package common

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

// Define constants or derive values for offsets and size
const (
	sigNameLength         int = 20 // Example fixed length for SigName
	pubKeyLengthBytes     int = 4  // int32
	privateKeyLengthBytes int = 4  // int32
	signatureLengthBytes  int = 4  // int32
	isValidByte           int = 1
	isPausedByte          int = 1
	totalLength           int = sigNameLength + pubKeyLengthBytes + privateKeyLengthBytes + signatureLengthBytes + isValidByte + isPausedByte
)

// Config holds the configurable parameters for your application.
type ConfigEnc struct {
	PubKeyLength     int    `json:"pubKeyLength"`
	PrivateKeyLength int    `json:"privateKeyLength"`
	SignatureLength  int    `json:"signatureLength"`
	SigName          string `json:"SigName"`
	IsValid          bool   `json:"isValid"`
	IsPaused         bool   `json:"isPaused"`
}

// NewConfig creates a Config with default values.
func NewConfigEnc1() *ConfigEnc {
	return &ConfigEnc{
		PubKeyLength:     897,
		PrivateKeyLength: 1281,
		SignatureLength:  666,
		SigName:          "Falcon-512",
		IsValid:          true,
		IsPaused:         false,
	}
}

// NewConfig creates a Config with default values.
func NewConfigEnc2() *ConfigEnc {
	return &ConfigEnc{
		PubKeyLength:     264608,                   // Default value for public key length
		PrivateKeyLength: 64,                       // Default value for private key length
		SignatureLength:  164,                      // Default value for signature length
		SigName:          "Rainbow-III-Compressed", // Default signature scheme name
		IsValid:          true,
		IsPaused:         false,
	}
}

// CreateEncryptionScheme
func CreateEncryptionScheme(sigName string, pubKeyLength int, privateKeyLength int, signatureLength int, isValid bool, isPaused bool) ConfigEnc {
	// Encryption scheme
	scheme := ConfigEnc{
		SigName:          sigName,
		PubKeyLength:     pubKeyLength,
		PrivateKeyLength: privateKeyLength,
		SignatureLength:  signatureLength,
		IsValid:          isValid,
		IsPaused:         isPaused,
	}

	return scheme
}

// ConvertBytesToStruct converts byte slice input to encryption scheme parameters.
func GenerateParamsEncryptionSchemesFromBytes(bb []byte) (sigName string, pubKeyLength int, privateKeyLength int, signatureLength int, isValid bool, isPaused bool, err error) {
	// Check if byte slice length is valid
	if len(bb) < totalLength {
		return "", 0, 0, 0, false, false, errors.New("invalid byte slice length")
	}

	// Initialize a reader
	reader := bytes.NewReader(bb)

	// Decode SigName (as UTF-8 string from fixed-length byte slice)
	sigNameBytes := make([]byte, sigNameLength)
	if _, err = reader.Read(sigNameBytes); err != nil {
		return "", 0, 0, 0, false, false, fmt.Errorf("failed to read SigName: %w", err)
	}
	sigName = string(bytes.Trim(sigNameBytes, "\x00")) // Remove trailing NULL bytes

	// Decode pubKeyLength (as int32)
	var pubKeyLength32 int32
	if err = binary.Read(reader, binary.LittleEndian, &pubKeyLength32); err != nil {
		return "", 0, 0, 0, false, false, fmt.Errorf("failed to read pubKeyLength: %w", err)
	}
	pubKeyLength = int(pubKeyLength32)

	// Decode privateKeyLength (as int32)
	var privateKeyLength32 int32
	if err = binary.Read(reader, binary.LittleEndian, &privateKeyLength32); err != nil {
		return "", 0, 0, 0, false, false, fmt.Errorf("failed to read privateKeyLength: %w", err)
	}
	privateKeyLength = int(privateKeyLength32)

	// Decode signatureLength (as int32)
	var signatureLength32 int32
	if err = binary.Read(reader, binary.LittleEndian, &signatureLength32); err != nil {
		return "", 0, 0, 0, false, false, fmt.Errorf("failed to read signatureLength: %w", err)
	}
	signatureLength = int(signatureLength32)

	// Decode isValid (as boolean)
	var isValidByte byte
	if err = binary.Read(reader, binary.LittleEndian, &isValidByte); err != nil {
		return "", 0, 0, 0, false, false, fmt.Errorf("failed to read isValid: %w", err)
	}
	isValid = isValidByte != 0

	// Decode isPaused (as boolean)
	var isPausedByte byte
	if err = binary.Read(reader, binary.LittleEndian, &isPausedByte); err != nil {
		return "", 0, 0, 0, false, false, fmt.Errorf("failed to read isPaused: %w", err)
	}
	isPaused = isPausedByte != 0

	return sigName, pubKeyLength, privateKeyLength, signatureLength, isValid, isPaused, nil
}

// GenerateBytesFromParams converts encryption scheme parameters to a byte slice.
func GenerateBytesFromParams(sigName string, pubKeyLength, privateKeyLength, signatureLength int, isValid, isPaused bool) ([]byte, error) {
	buf := new(bytes.Buffer)

	// Ensure SigName fits fixed length
	paddedSigName := make([]byte, sigNameLength)
	copy(paddedSigName, sigName)

	// Encode SigName
	if _, err := buf.Write(paddedSigName); err != nil {
		return nil, fmt.Errorf("failed to write SigName: %w", err)
	}

	// Encode pubKeyLength (as int32)
	if err := binary.Write(buf, binary.LittleEndian, int32(pubKeyLength)); err != nil {
		return nil, fmt.Errorf("failed to write pubKeyLength: %w", err)
	}

	// Encode privateKeyLength (as int32)
	if err := binary.Write(buf, binary.LittleEndian, int32(privateKeyLength)); err != nil {
		return nil, fmt.Errorf("failed to write privateKeyLength: %w", err)
	}

	// Encode signatureLength (as int32)
	if err := binary.Write(buf, binary.LittleEndian, int32(signatureLength)); err != nil {
		return nil, fmt.Errorf("failed to write signatureLength: %w", err)
	}

	// Encode isValid (as byte)
	var isValidByte byte
	if isValid {
		isValidByte = 1
	}
	if err := binary.Write(buf, binary.LittleEndian, isValidByte); err != nil {
		return nil, fmt.Errorf("failed to write isValid: %w", err)
	}

	// Encode isPaused (as byte)
	var isPausedByte byte
	if isPaused {
		isPausedByte = 1
	}
	if err := binary.Write(buf, binary.LittleEndian, isPausedByte); err != nil {
		return nil, fmt.Errorf("failed to write isPaused: %w", err)
	}

	return buf.Bytes(), nil
}
