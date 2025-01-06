package voting

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/quad-foundation/quad-node/account"
	"github.com/quad-foundation/quad-node/common"
	"sync"
)

type Votes struct {
	Values []byte `json:"values"`
	Height int64  `json:"height"`
	Staked int64  `json:"staked"`
}

var (
	VotesEncryption1        = make(map[uint8]Votes)
	VotesEncryption2        = make(map[uint8]Votes)
	VotesEncryption1RWMutex = make(map[uint8]*sync.RWMutex)
	VotesEncryption2RWMutex = make(map[uint8]*sync.RWMutex)
)

func init() {
	for i := 0; i < 256; i++ {
		VotesEncryption1RWMutex[uint8(i)] = new(sync.RWMutex)
		VotesEncryption2RWMutex[uint8(i)] = new(sync.RWMutex)
	}
}

func SaveVotesEncryption1(value []byte, height int64, delegatedAccount common.Address, staked int64) error {
	id, err := common.GetIDFromDelegatedAccountAddress(delegatedAccount)
	if err != nil {
		return err
	}

	if id >= 256 {
		return fmt.Errorf("delegated account is invalid: %d", id)
	}
	VotesEncryption1RWMutex[uint8(id)].Lock()
	defer VotesEncryption1RWMutex[uint8(id)].Unlock()

	po, exists := VotesEncryption1[uint8(id)]
	if !exists || po.Height <= height {
		VotesEncryption1[uint8(id)] = Votes{
			Values: value,
			Height: height,
			Staked: staked,
		}
	} else {
		return errors.New("invalid height in voting, 1")
	}

	return nil
}

func SaveVotesEncryption2(value []byte, height int64, delegatedAccount common.Address, staked int64) error {
	id, err := common.GetIDFromDelegatedAccountAddress(delegatedAccount)
	if err != nil {
		return err
	}

	if id >= 256 {
		return fmt.Errorf("delegated account is invalid: %d", id)
	}
	VotesEncryption2RWMutex[uint8(id)].Lock()
	defer VotesEncryption2RWMutex[uint8(id)].Unlock()

	po, exists := VotesEncryption2[uint8(id)]
	if !exists || po.Height <= height {
		VotesEncryption2[uint8(id)] = Votes{
			Values: value,
			Height: height,
			Staked: staked,
		}
	} else {
		return errors.New("invalid height in voting, 2")
	}

	return nil
}

func GenerateEncryption1Data(height int64) ([]byte, [][]byte, int64) {
	valueData := make([]byte, 0)
	values := [][]byte{}
	staked := int64(0)
	for i, po := range VotesEncryption1 {
		VotesEncryption1RWMutex[i].RLock()
		if height <= po.Height+common.VotingHeightDistance && len(po.Values) > 0 {
			valueData = append(valueData, i)
			valueData = append(valueData, common.GetByteInt64(po.Height)...)
			valueData = append(valueData, common.BytesToLenAndBytes(po.Values[:])...)
			values = append(values, po.Values[:])
			staked += po.Staked
		}
		VotesEncryption1RWMutex[i].RUnlock()
	}
	return valueData, values, staked
}

func GenerateEncryption2Data(height int64) ([]byte, [][]byte, int64) {
	valueData := make([]byte, 0)
	values := [][]byte{}
	staked := int64(0)
	for i, po := range VotesEncryption2 {
		VotesEncryption2RWMutex[i].RLock()
		if height <= po.Height+common.VotingHeightDistance && len(po.Values) > 0 {
			valueData = append(valueData, i)
			valueData = append(valueData, common.GetByteInt64(po.Height)...)
			valueData = append(valueData, common.BytesToLenAndBytes(po.Values[:])...)
			values = append(values, po.Values[:])
			staked += po.Staked
		}
		VotesEncryption2RWMutex[i].RUnlock()
	}
	return valueData, values, staked
}

func ParseVotesData(votingData []byte) (map[uint8]Votes, [][]byte, int64, error) {
	parsedData := make(map[uint8]Votes)
	dataLen := len(votingData)
	values := [][]byte{}
	allStaked := int64(0)

	if dataLen%17 != 0 {
		return nil, nil, 0, fmt.Errorf("invalid priceData length: %d", dataLen)
	}
	var err error
	value := []byte{}
	b := votingData[:]
	for i := 0; i < dataLen; i += 17 {
		id := b[i]
		height := common.GetInt64FromByte(b[i+1 : i+9])
		value, b, err = common.BytesWithLenToBytes(b[i+9:])
		if err != nil {
			return nil, nil, 0, err
		}
		values = append(values, value)
		_, staked, _ := account.GetStakedInDelegatedAccount(int(id))
		allStaked += int64(staked)
		parsedData[id] = Votes{
			Values: value,
			Height: height,
			Staked: int64(staked),
		}
	}

	return parsedData, values, allStaked, nil
}

// removeOneDifferent removes one outlier byte slice, if all others are the same.
func removeOneDifferent(values [][]byte) [][]byte {
	if len(values) <= 1 {
		return values
	}

	byteCount := make(map[string]int)
	var commonPattern []byte
	maxCount := 0

	// Count occurrences of each byte slice
	for _, v := range values {
		vs := string(v) // Convert the byte slice to a string for easy comparison in a map
		byteCount[vs]++
		if byteCount[vs] > maxCount {
			maxCount = byteCount[vs]
			commonPattern = v
		}
	}

	// Remove one outlier byte slice if present
	var result [][]byte
	outlierRemoved := false
	for _, v := range values {
		if !bytes.Equal(v, commonPattern) && !outlierRemoved {
			// Skip adding this outlier byte slice to result once
			outlierRemoved = true
		} else {
			result = append(result, v)
		}
	}

	return result
}

// one has to think what happens when verification is not on current block than GetStakedInDelegatedAccount should depend on height
func VerifyEncryptionForPausing(height int64, totalStaked int64, version int) bool {
	values := [][]byte{}
	staked := int64(0)
	if version == 1 {
		_, values, staked = GenerateEncryption1Data(height)
	} else if version == 2 {
		_, values, staked = GenerateEncryption2Data(height)
	} else {
		return false
	}

	// 1/3 for pausing
	if staked <= totalStaked/3 {
		return false
	}

	// Remove max and min value
	if len(values) > 2 {
		values = removeOneDifferent(values)
	}

	if len(values) == 0 {
		return false
	}

	// Compare bytes if the same
	isSame := true
	first := values[0]
	for _, b := range values {
		if !bytes.Equal(first, b) {
			isSame = false
			break
		}
	}

	return isSame
}

// one has to think what happens when verification is not on current block than GetStakedInDelegatedAccount should depend on height
func VerifyEncryptionForInvalidation(height int64, totalStaked int64, version int) bool {
	values := [][]byte{}
	staked := int64(0)
	if version == 1 {
		_, values, staked = GenerateEncryption1Data(height)
	} else if version == 2 {
		_, values, staked = GenerateEncryption2Data(height)
	} else {
		return false
	}

	// 2/3 for invalidation
	if staked <= 2*totalStaked/3 {
		return false
	}

	// Remove max and min value
	if len(values) > 2 {
		values = removeOneDifferent(values)
	}

	if len(values) == 0 {
		return false
	}

	// Compare bytes if the same
	isSame := true
	first := values[0]
	for _, b := range values {
		if !bytes.Equal(first, b) {
			isSame = false
			break
		}
	}

	return isSame
}
