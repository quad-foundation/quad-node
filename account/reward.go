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

func GetReward(height int64) int64 {
	cr := float64(common.InitialReward) * getRatio(height)

	cr = math.Round(cr)
	return int64(cr)
}
