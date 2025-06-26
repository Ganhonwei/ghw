package handler

import (
	"fmt"
	"goserver/pkg/utils"
)

// 随机一个下标 数组内部结构: id+num+weight
func RandomIndex(target [][]int32) (int32, error) {
	var weights int32 = 0
	for _, v := range target {
		if len(v) < 3 {
			continue
		}
		weights += v[2]
	}
	w := utils.RandInt32N(weights) + 1
	var min int32 = 0
	var max int32 = 0
	for i, v := range target {
		if len(v) < 3 {
			continue
		}
		max += v[2]
		if max >= w && w > min {
			return int32(i), nil
		}
		min += v[2]
	}
	return 0, fmt.Errorf("not found index")
}
