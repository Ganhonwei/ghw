package utils

import (
	"math"
	"strconv"

	"github.com/bwmarrin/snowflake"
)

const snowStartTime int64 = 1714533741822

var snowNode *snowflake.Node

func InitSnowflakeId(machineID int64) {
	var err error
	snowflake.Epoch = snowStartTime
	snowNode, err = snowflake.NewNode(machineID)
	if err != nil {
		panic(err)
	}
}

func GenSnowId() int64 {
	return snowNode.Generate().Int64()
}

func GenSnowIdS() string {
	return strconv.FormatInt(snowNode.Generate().Int64(), 10)
}

// 使用位操作的方式转换
func Int64ToFloat64Bits(i int64) float64 {
	return math.Float64frombits(uint64(i))
}

func Float64ToInt64Bits(f float64) int64 {
	return int64(math.Float64bits(f))
}
