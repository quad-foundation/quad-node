package common

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

var (
	ValidChains                         = []uint8{0, 1, 2, 3, 4}
	Decimals                    uint8   = 8
	MaxTotalSupply              int64   = 230000000000000000
	RewardChangeInterval        int64   = 8640 * 3
	InitialReward               int64   = 100000000 // 3993055556
	ValidationTag                       = "validationTag"
	DifficultyMultiplier        int32   = 10
	BlockTimeInterval           float32 = 5 // 5 sec.
	DifficultyChange            float32 = 5
	MaxGasUsage                 int64   = 137000000 // circa 65k transactions in block
	MaxGasPrice                 int64   = 100000
	MaxTransactionsPerBlock     int32   = 100000
	MaxTransactionsSidePerBlock int32   = 100
	ConfirmationsNumber         int64   = 6
	NumberOfHashesInBucket      int64   = 32
	NumberOfBlocksInBucket      int64   = 20
	MinStakingForNode           int64   = 100000000000000
	MinStakingUser              int64   = 100000000000 // should be 100000000000
	MinDistributedAmount        int64   = 100000000
)

// DB prefixes
var (
	BlocksDBPrefix                          = "B0"
	BlockHashesDBPrefix                     = "K0"
	BlockHeaderDBPrefix                     = "H0"
	PubKeyDBPrefix                          = "PK"
	StakeTransactionDBPrefix                = "ST"
	StakingPoolHashesDBPrefix               = "SP"
	StakingPendingHashesDBPrefix            = "SE"
	StakingSyncingHashesDBPrefix            = "SD"
	BlockStakingHashesDBPrefix              = "SB"
	BlockByHeightDBPrefix                   = "I0"
	RootHashByHeightDBPrefix                = "R0"
	BlockTransactionHashesDBPrefix          = "L0"
	OutputLogsHashesDBPrefix                = "O0"
	OutputLogDBPrefix                       = "Z0"
	OutputAddressesHashesDBPrefix           = "C0"
	StateAccountsHashesDBPrefix             = "A0"
	StateAccountsHashesForNextBlockDBPrefix = "A1"
	AccountLogsByHeightDBPrefix             = "AL"
	TransactionDBPrefix                     = [2]string{"T0", "T1"}
	TransactionPoolHashesDBPrefix           = [2]string{"P0", "P1"}
	TransactionPendingHashesDBPrefix        = [2]string{"E0", "E1"}
	TransactionSyncingHashesDBPrefix        = [2]string{"D0", "D1"}
)

var chainID = int16(23)
var delegatedAccount Address
var rewardPercentage int16
var genesisAccounts []Address
var genesisAccountsStake []Address

func GetChainID() int16 {
	return chainID
}

func GetDelegatedAccount() Address {
	return delegatedAccount
}
func GetRewardPercentage() int16 {
	return rewardPercentage
}
func init() {
	//log.SetOutput(io.Discard)

	homePath, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	err = godotenv.Load(homePath + "/.chainpqc/.env")
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}
	da, err := strconv.Atoi(os.Getenv("DELEGATED_ACCOUNT"))
	if err != nil {
		log.Fatal("Error getting DELEGATED_ACCOUNT")
	}
	delegatedAccount = GetDelegatedAccountAddress(int16(da))

	//DefaultPercentageReward int16 = 1000 // 1%
	v, err := strconv.Atoi(os.Getenv("REWARD_PERCENTAGE"))
	if err != nil {
		log.Fatal("Error getting REWARD_PERCENTAGE")
	}
	rewardPercentage = int16(v)
}
