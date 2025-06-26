package handler

import (
	"goserver/gen/pb"
	"goserver/pkg/data"
)

// PackPlaneFreeRoom 打包百人房间信息
func PackPlaneFreeRoom(d *data.DeskData) *pb.CRASHFreeRoom {
	return &pb.CRASHFreeRoom{
		Roomid: d.Rid,                  //牌局id
		Gtype:  d.Gtype,                //game type
		Rtype:  d.Rtype,                //room type
		Dtype:  d.Dtype,                //desk type
		Rname:  d.Rname,                //room name
		Count:  d.Count,                //当前房间限制玩家数量
		Chip:   d.Game.PLANE.ChipLimit, //房间下注筹码
	}
}

// PackPlaneFreeUser 打包进入百人玩家基础数据
func PackPlaneFreeUser(p *data.User) *pb.CRASHFreeUser {
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
