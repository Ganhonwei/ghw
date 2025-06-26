package handler

import (
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/config"
	"math/rand"
	"time"
)

// PackCrashFreeRoom 打包百人房间信息
func _PackCrashFreeRoom(d *data.DeskData) *pb.CRASHFreeRoom {
	return &pb.CRASHFreeRoom{
		Roomid:   d.Rid,                        //牌局id
		Gtype:    d.Gtype,                      //game type
		Rtype:    d.Rtype,                      //room type
		Dtype:    d.Dtype,                      //desk type
		Rname:    d.Rname,                      //room name
		Count:    d.Count,                      //当前房间限制玩家数量
		Chip:     d.Game.CRASH.ChipLimit,       //房间下注筹码
		ChipSeat: int32(d.Game.CRASH.ChipSeat), //筹码位置
	}
}

// PackCrashFreeUser 打包进入百人玩家基础数据
func PackCrashFreeUser(p *data.User) *pb.CRASHFreeUser {
	return &pb.CRASHFreeUser{
		Userid:   p.GetUserid(),
		Nickname: p.GetNickname(),
		Phone:    p.GetPhone(),
		Sex:      p.GetSex(),
		Photo:    p.GetPhoto(),
		Coin:     p.GetCoin(),
		Diamond:  p.GetDiamond(),
		VipLv:    int32(p.Vip.Lv),
	}
}

func TiggerCrashStrategy(user *data.User, d *data.BaseStrategy) bool {
	bean := config.GetCrashStrategy()
	if bean.Id == 0 {
		return false
	}

	for i, v := range d.UserType {
		if v == 0 && user.RegistArea == i {
			return false
		}
	}

	chargeType := GetChargeType(user)
	for i, v := range d.ChargeType {
		if v == 0 {
			continue
		}
		// if i < 2 && chargeType == data.NoneCharge {
		// 	if i == 0 && user.State == data.NoveiceState { // 新手
		// 		return true
		// 	}
		// 	if i == 1 && user.State == data.ExceptionState { // 平民
		// 		return true
		// 	}
		// }
		if chargeType == i {
			return true
		}
	}

	return false
}

func CreateFakeHistory(count int) []int32 {
	list := make([]int32, 0, count)
	for i := range count {
		r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(i)))
		multiple := CrashBoomMultiple(r)
		list = append(list, multiple-1)
	}
	return list
}

// Deprecated
func _CreateFakeHistory(backRate float64) []int32 {
	list := make([]int32, 0)
	// 生成假录单
	for i := 0; i < 10; i++ {
		boom := false
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for j := 100; j < 5000; j++ {
			if CrashIsBoom(int32(j), backRate, 0, r) {
				boom = true
				list = append(list, int32(j))
				break
			}
		}
		if !boom {
			list = append(list, int32(5000))
		}
	}
	return list
}
