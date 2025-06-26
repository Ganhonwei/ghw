package up7

import (
	"goserver/pkg/utils"
)

// 自然概率开奖
func (t *Desk) NatureWinner() {
	// 指定房间
	// if t.Dtype != int32(pb.DESK_TYPE_NORMAL) {
	// 	return
	// }

	// t.DeskFree.LHCards = make([]uint32, 0)

	// max := utils.RandInt32N(11) + 2
	// var point int32
	// if max <= 6 {
	// 	point = utils.RandInt32N(max-1) + 1
	// } else {
	// 	point = utils.RandInt32N(13-max) + max - 6
	// }
	// t.LHCards = []uint32{uint32(point), uint32(max - point)}

	dice1 := utils.RandInt32N(6) + 1
	dice2 := utils.RandInt32N(6) + 1
	t.LHCards = []uint32{uint32(dice1), uint32(dice2)}

	// 中奖上限
}
