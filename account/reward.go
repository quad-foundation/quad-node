package account

import (
	"github.com/quad/quad-node/common"
	"math"
)

func getRatio(height int64) float64 {

	rci := float64(common.RewardChangeInterval)
	h := float64(height)

	return math.Pow(0.999, math.Floor(h/rci))
}

func getRemainingSupply(supply int64) int64 {
	return common.MaxTotalSupply - supply
}

func GetReward(supply int64) int64 {
	cr := common.RewardRatio * float64(getRemainingSupply(supply))
	cr = math.Round(cr)
	return int64(cr)
}
