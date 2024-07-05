package common

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

var (
	Decimals                  uint8   = 8
	MaxTotalSupply            int64   = 230000000000000000
	InitSupply                int64   = 23000000000000000
	RewardRatio                       = 1e-7
	ValidationTag                     = "validationTag"
	DifficultyMultiplier      int32   = 10
	BlockTimeInterval         float32 = 5 // 5 sec.
	DifficultyChange          float32 = 5
	MaxGasUsage               int64   = 137000000 // circa 65k transactions in block
	MaxGasPrice               int64   = 100000
	MaxTransactionsPerBlock   int16   = 32000
	MaxTransactionInPool              = 1000000
	MaxPeersConnected         int     = 6
	NumberOfHashesInBucket    int64   = 32
	NumberOfBlocksInBucket    int64   = 20
	MinStakingForNode         int64   = 100000000000000
	MinStakingUser            int64   = 100000000000 // should be 100000000000
	MinDistributedAmount      int64   = 100000000
	DefaultWalletHomePath             = "~/.quad/db/wallet/"
	DefaultBlockchainHomePath         = "~/.quad/db/blockchain/"
)

// db prefixes
var (
	BlocksDBPrefix                     = [2]byte{'B', 'I'}
	StatDBPrefix                       = [2]byte{'M', 'S'}
	BlockHeaderDBPrefix                = [2]byte{'H', 'B'}
	WalletDBPrefix                     = [2]byte{'W', '0'}
	PubKeyDBPrefix                     = [2]byte{'P', 'K'}
	BlockByHeightDBPrefix              = [2]byte{'B', 'H'}
	TransactionsHashesByHeightDBPrefix = [2]byte{'R', 'H'}
	MerkleTreeDBPrefix                 = [2]byte{'M', 'M'}
	MerkleNodeDBPrefix                 = [2]byte{'N', 'N'}
	RootHashMerkleTreeDBPrefix         = [2]byte{'R', 'R'}
	TransactionDBPrefix                = [2]byte{'T', 'T'}
	//StakingDBPrefix                    = [2]byte{'S', 'S'}
	TransactionPoolHashesDBPrefix    = [2]byte{'D', '0'}
	TransactionToSendHashesDBPrefix  = [2]byte{'E', '0'}
	TransactionSyncingHashesDBPrefix = [2]byte{'S', '0'}
	AccountsDBPrefix                 = [2]byte{'A', 'C'}
	StakingAccountsDBPrefix          = [2]byte{'S', 'A'}
	OutputLogsHashesDBPrefix         = [2]byte{'O', '0'}
	OutputLogDBPrefix                = [2]byte{'Z', '0'}
	OutputAddressesHashesDBPrefix    = [2]byte{'C', '0'}
	TokenDetailsDBPrefix             = [2]byte{'T', 'D'}
	DexAccountsDBPrefix              = [2]byte{'D', 'A'}
)

var chainID = int16(23)
var delegatedAccount Address
var rewardPercentage float64
var ShiftToPastInReset int64

func GetChainID() int16 {
	return chainID
}

func SetChainID(chainid int16) {
	chainID = chainid
}

func GetDelegatedAccount() Address {
	return delegatedAccount
}
func GetRewardPercentage() float64 {
	return rewardPercentage
}
func init() {
	//log.SetOutput(io.Discard)
	ShiftToPastInReset = 1
	homePath, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	err = godotenv.Load(homePath + "/.quad/.env")
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}
	da, err := strconv.Atoi(os.Getenv("DELEGATED_ACCOUNT"))
	if err != nil {
		log.Fatal("Error getting DELEGATED_ACCOUNT")
	}
	delegatedAccount = GetDelegatedAccountAddress(int16(da))

	//DefaultPercentageReward int16 = 1000 // 0.001
	v, err := strconv.Atoi(os.Getenv("REWARD_PERCENTAGE"))
	if err != nil {
		log.Fatal("Error getting REWARD_PERCENTAGE")
	}
	rewardPercentage = float64(v) / 1000.0
	if rewardPercentage > 0.5 {
		log.Fatal("reward for operational account has to be less than 50%")
	}
}
