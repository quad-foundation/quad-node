package oracles

import (
	"errors"
	"fmt"
	"github.com/quad-foundation/quad-node/account"
	"github.com/quad-foundation/quad-node/common"
	"sort"
	"sync"
)

type PriceOracle struct {
	Price  int64 `json:"price"`
	Height int64 `json:"height"`
	Staked int64 `json:"staked"`
}

type RandOracle struct {
	Rand   int64 `json:"rand"`
	Height int64 `json:"height"`
	Staked int64 `json:"staked"`
}

var (
	PriceOracles        = make(map[uint8]PriceOracle)
	PriceOraclesRWMutex = make(map[uint8]*sync.RWMutex)

	RandOracles        = make(map[uint8]RandOracle)
	RandOraclesRWMutex = make(map[uint8]*sync.RWMutex)
)

func init() {
	for i := 0; i < 256; i++ {
		PriceOraclesRWMutex[uint8(i)] = new(sync.RWMutex)
		RandOraclesRWMutex[uint8(i)] = new(sync.RWMutex)
	}
}

func SavePriceOracle(price int64, height int64, delegatedAccount common.Address, staked int64) error {
	id, err := common.GetIDFromDelegatedAccountAddress(delegatedAccount)
	if err != nil {
		return err
	}

	if id >= 256 {
		return fmt.Errorf("delegated account is invalid: %d", id)
	}
	PriceOraclesRWMutex[uint8(id)].Lock()
	defer PriceOraclesRWMutex[uint8(id)].Unlock()

	po, exists := PriceOracles[uint8(id)]
	if !exists || po.Height <= height {
		PriceOracles[uint8(id)] = PriceOracle{
			Price:  price,
			Height: height,
			Staked: staked,
		}
	} else {
		return errors.New("invalid height in price oracle")
	}

	return nil
}

func SaveRandOracle(rand int64, height int64, delegatedAccount common.Address, staked int64) error {
	id, err := common.GetIDFromDelegatedAccountAddress(delegatedAccount)
	if err != nil {
		return err
	}

	if id >= 256 {
		return fmt.Errorf("delegated account is invalid: %d", id)
	}
	RandOraclesRWMutex[uint8(id)].Lock()
	defer RandOraclesRWMutex[uint8(id)].Unlock()

	po, exists := RandOracles[uint8(id)]
	if !exists || po.Height <= height {
		RandOracles[uint8(id)] = RandOracle{
			Rand:   rand,
			Height: height,
			Staked: staked,
		}
	} else {
		return errors.New("invalid height in rand oracle")
	}

	return nil
}

func GeneratePriceData(height int64) ([]byte, []int64, int64) {
	priceData := make([]byte, 0)
	prices := []int64{}
	staked := int64(0)
	for i, po := range PriceOracles {
		PriceOraclesRWMutex[i].RLock()
		if height <= po.Height+common.OraclesHeightDistance && po.Price > 0 {
			priceData = append(priceData, i)
			priceData = append(priceData, common.GetByteInt64(po.Height)...)
			priceData = append(priceData, common.GetByteInt64(po.Price)...)
			prices = append(prices, po.Price)
			staked += po.Staked
		}
		PriceOraclesRWMutex[i].RUnlock()
	}
	return priceData, prices, staked
}

func ParsePriceData(priceData []byte) (map[uint8]PriceOracle, []int64, int64, error) {
	parsedData := make(map[uint8]PriceOracle)
	dataLen := len(priceData)
	prices := []int64{}
	allStaked := int64(0)

	if dataLen%17 != 0 {
		return nil, nil, 0, fmt.Errorf("invalid priceData length: %d", dataLen)
	}

	for i := 0; i < dataLen; i += 17 {
		id := priceData[i]
		height := common.GetInt64FromByte(priceData[i+1 : i+9])
		price := common.GetInt64FromByte(priceData[i+9 : i+17])
		prices = append(prices, price)
		_, staked, _ := account.GetStakedInDelegatedAccount(int(id))
		allStaked += int64(staked)
		parsedData[id] = PriceOracle{
			Price:  price,
			Height: height,
			Staked: int64(staked),
		}
	}

	return parsedData, prices, allStaked, nil
}

func ParseRandData(randData []byte) (map[uint8]RandOracle, []byte, int64, error) {
	parsedData := make(map[uint8]RandOracle)
	dataLen := len(randData)
	rands := make([]byte, 0)
	allStaked := int64(0)

	if dataLen%17 != 0 {
		return nil, nil, 0, fmt.Errorf("invalid randData length: %d", dataLen)
	}

	for i := 0; i < dataLen; i += 17 {
		id := randData[i]
		height := common.GetInt64FromByte(randData[i+1 : i+9])
		rand := common.GetInt64FromByte(randData[i+9 : i+17])
		rands = append(rands, randData[i+9:i+17]...)
		_, staked, _ := account.GetStakedInDelegatedAccount(int(id))
		allStaked += int64(staked)
		parsedData[id] = RandOracle{
			Rand:   rand,
			Height: height,
			Staked: int64(staked),
		}
	}

	return parsedData, rands, allStaked, nil
}

func GenerateRandData(height int64) ([]byte, []byte, int64) {
	randData := make([]byte, 0)
	rands := make([]byte, 0)
	staked := int64(0)
	for i, po := range RandOracles {
		RandOraclesRWMutex[i].RLock()
		if height <= po.Height+common.OraclesHeightDistance && po.Rand > 0 {
			randData = append(randData, i)
			randData = append(randData, common.GetByteInt64(po.Height)...)
			randData = append(randData, common.GetByteInt64(po.Rand)...)
			rands = append(rands, common.GetByteInt64(po.Rand)...)
			staked += po.Staked
		}
		RandOraclesRWMutex[i].RUnlock()
	}
	return randData, rands, staked
}

func CalculateRandOracle(height int64, totalStaked int64) (int64, []byte, error) {
	var rand int64
	randData, rands, staked := GenerateRandData(height)

	if staked <= 2*totalStaked/3 {
		return 0, randData, errors.New("in rand, there is not enough staked value for 2/3")
	}

	if len(rands) == 0 {
		return 0, randData, errors.New("not enough rands propositions")
	}

	// Calculate hash from all rand numbers propositions
	bytes, err := common.CalcHashFromBytes(rands)
	if err != nil {
		return 0, nil, err
	}
	rand = common.GetInt64FromByte(bytes[24:])
	return rand, randData, nil
}

func VerifyRandOracle(height int64, totalStaked int64, randBlock int64, randData []byte) bool {
	_, rands, staked, err := ParseRandData(randData)
	if err != nil {
		return false
	}

	if staked <= 2*totalStaked/3 {
		return false
	}

	if len(rands) == 0 {
		return false
	}

	// Calculate hash from all rand numbers propositions
	bytes, err := common.CalcHashFromBytes(rands)
	if err != nil {
		return false
	}
	rand := common.GetInt64FromByte(bytes[24:])
	return rand == randBlock
}

// one has to think what happens when verification is not on current block than GetStakedInDelegatedAccount should depend on height
func VerifyPriceOracle(height int64, totalStaked int64, priceBlock int64, priceData []byte) bool {
	_, prices, staked, err := ParsePriceData(priceData)
	if err != nil {
		return false
	}

	if staked <= 2*totalStaked/3 {
		return false
	}

	// Remove max and min value
	if len(prices) > 2 {
		prices = removeMinMax(prices)
	}

	if len(prices) == 0 {
		return false
	}

	// Calculate median price
	price := Median(prices)

	return price == priceBlock
}

func CalculatePriceOracle(height int64, totalStaked int64) (int64, []byte, error) {
	var price int64
	priceData, prices, staked := GeneratePriceData(height)

	if staked <= 2*totalStaked/3 {
		return 0, priceData, errors.New("in price, there is not enough staked value for 2/3")
	}

	// Remove max and min value
	if len(prices) > 2 {
		prices = removeMinMax(prices)
	}

	if len(prices) == 0 {
		return 0, priceData, errors.New("not enough prices propositions after removing min and max")
	}

	// Calculate median price
	price = Median(prices)

	return price, priceData, nil
}

func removeMinMax(prices []int64) []int64 {
	if len(prices) <= 2 {
		return []int64{}
	}

	// Sort
	sort.Slice(prices, func(i, j int) bool {
		return prices[i] < prices[j]
	})

	// Remove first, min, and last, max.
	return prices[1 : len(prices)-1]
}

func Median(prices []int64) int64 {
	if len(prices) == 0 {
		return 0
	}

	// Sort prices
	sort.Slice(prices, func(i, j int) bool {
		return prices[i] < prices[j]
	})

	mid := len(prices) / 2
	if len(prices)%2 == 0 {
		// If number of elements are even
		return (prices[mid-1] + prices[mid]) / 2
	}
	// When odd number of elements, so return in the middle one
	return prices[mid]
}
