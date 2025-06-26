package handler

import (
	"goserver/gen/pb"
	"goserver/pkg/data"
)

// PackLHFreeRoom 打包百人房间信息
func PackUPFreeRoom(d *data.DeskData) *pb.UPFreeRoom {
	return &pb.UPFreeRoom{
		Roomid: d.Rid,   //牌局id
		Gtype:  d.Gtype, //game type
		Rtype:  d.Rtype, //room type
		Dtype:  d.Dtype, //desk type
		Rname:  d.Rname, //room name
		Count:  d.Count, //当前房间限制玩家数量
		// Chip:   d.Game.UP.ChipLimit, //房间下注筹码
	}
}

// PackUPFreeUser 打包进入百人玩家基础数据
func PackUPFreeUser(p *data.User) *pb.UPFreeUser {
	return &pb.UPFreeUser{
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
