package handler

import (
	"goserver/gen/pb"
	"goserver/pkg/data"
)

// PackCPFreeRoom 打包百人房间信息
func PackCPFreeRoom(d *data.DeskData) *pb.LotteryRoom {
	return &pb.LotteryRoom{
		Roomid: d.Rid,                    //牌局id
		Gtype:  d.Gtype,                  //game type
		Rtype:  d.Rtype,                  //room type
		Dtype:  d.Dtype,                  //desk type
		Chip:   d.Game.LOTTERY.ChipLimit, //房间下注筹码
	}
}

func TiggerCPStrategy(user *data.User, d *data.BaseStrategy, chargeType int) bool {
	for i, v := range d.UserType {
		if v == 0 && user.RegistArea == i {
			return false
		}
	}

	// chargeType := GetChargeType(user)
	for i, v := range d.ChargeType {
		if v == 0 {
			continue
		}
		if chargeType == i {
			return true
		}
	}

	return false
}
