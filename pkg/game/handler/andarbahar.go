package handler

import (
	"goserver/gen/pb"
	"goserver/pkg/data"
)

// PackABFreeUser 打包进入百人玩家基础数据
func PackABFreeUser(p *data.User) *pb.ABFreeUser {
	return &pb.ABFreeUser{
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

// PackABCoinUser 打包进入百人玩家基础数据
func PackABCoinUser(p *data.User) *pb.ABRoomUser {
	return &pb.ABRoomUser{
		Userid:   p.GetUserid(),
		Nickname: p.GetNickname(),
		Phone:    p.GetPhone(),
		Sex:      p.GetSex(),
		Photo:    p.GetPhoto(),
		Coin:     p.GetCoin(),
		Diamond:  p.GetDiamond(),
		Lat:      p.Lat,
		Lng:      p.Lng,
		Address:  p.Address,
		Sign:     p.Sign,
		VipLv:    int32(p.Vip.Lv),
	}
}

// PackABFreeRoom 打包百人房间信息
func PackABFreeRoom(d *data.DeskData) *pb.ABFreeRoom {
	return &pb.ABFreeRoom{
		Roomid: d.Rid,               //牌局id
		Gtype:  d.Gtype,             //game type
		Rtype:  d.Rtype,             //room type
		Dtype:  d.Dtype,             //desk type
		Rname:  d.Rname,             //room name
		Count:  d.Count,             //当前房间限制玩家数量
		Chip:   d.Game.AB.ChipLimit, //房间下注筹码
	}
}

// PackABCreateMsg 创建房间消息
func PackABCreateMsg(d *data.DeskData) *pb.ABCreateRoomRsp {
	msg := new(pb.ABCreateRoomRsp)
	msg.Data = PackABPrivRoom(d)
	return msg
}

// PackABPrivRoom 打包私人房间信息
func PackABPrivRoom(d *data.DeskData) (room *pb.ABPrivRoom) {
	room = &pb.ABPrivRoom{
		Roomid: d.Rid,               //牌局id
		Gtype:  d.Gtype,             //game type
		Rtype:  d.Rtype,             //room type
		Dtype:  d.Dtype,             //desk type
		Rname:  d.Rname,             //room name
		Count:  d.Count,             //当前房间限制玩家数量
		Chip:   d.Game.AB.ChipLimit, //房间下注筹码
		Code:   d.Code,
		Round:  d.Round,
		Userid: d.Cid,
	}
	return
}

// PackABPrivUser 打包进入私人房玩家基础数据
func PackABPrivUser(p *data.User) *pb.ABPrivUser {
	return &pb.ABPrivUser{
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

func TiggerABStrategy(user *data.User, d *data.BaseStrategy, chargeType int) bool {
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
