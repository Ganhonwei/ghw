package handler

import (
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/table"
)

// PackRBFreeUser 打包进入百人玩家基础数据
func PackRBFreeUser(p *data.User) *pb.RBFreeUser {
	return &pb.RBFreeUser{
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

// PackRBRoomBets 打包百人下注信息
func PackRBRoomBets(bets map[uint32]int64) (msg []*pb.RBRoomBets) {
	for k, v := range bets {
		msg2 := &pb.RBRoomBets{
			Seat: k,
			Bets: v,
		}
		msg = append(msg, msg2)
	}
	return
}

// PackRBFreeRoom 打包百人房间信息
func PackRBFreeRoom(d *data.DeskData) *pb.RBFreeRoom {
	room := &pb.RBFreeRoom{
		Roomid: d.Rid,                                  //牌局id
		Gtype:  d.Gtype,                                //game type
		Rtype:  d.Rtype,                                //room type
		Dtype:  d.Dtype,                                //desk type
		Rname:  d.Rname,                                //room name
		Count:  d.Count,                                //当前房间限制玩家数量
		Chip:   []int32{500, 2000, 5000, 10000, 50000}, //房间下注筹码
	}
	base := table.GetTables().RbRoomBaseTable.Get()
	if base != nil {
		room.Chip = base.ChipLimit
	}
	return room
}

// RBLeaveMsg 离开消息
func RBLeaveMsg(userid string, seat uint32) *pb.RBLeaveRsp {
	return &pb.RBLeaveRsp{
		Seat:   seat,
		Userid: userid,
	}
}

// PackRBCoinRoom 打包百人房间信息
func PackRBCoinRoom(d *data.DeskData) *pb.RBRoomData {
	return &pb.RBRoomData{
		Roomid:   d.Rid,     //牌局id
		Gtype:    d.Gtype,   //game type
		Rtype:    d.Rtype,   //room type
		Dtype:    d.Dtype,   //desk type
		Ltype:    d.Ltype,   //level type
		Rname:    d.Rname,   //room name
		Count:    d.Count,   //当前房间限制玩家数量
		Ante:     d.Ante,    //房间底分
		Round:    d.Round,   //
		Userid:   d.Cid,     //
		Expire:   d.Expire,  //
		Code:     d.Code,    //
		Minimum:  d.Minimum, //
		Maximum:  d.Maximum, //
		Pub:      d.Pub,
		Mode:     d.Mode,
		Multiple: d.Multiple,
	}
}

// PackRBCoinUser 打包进入百人玩家基础数据
func PackRBCoinUser(p *data.User) *pb.RBRoomUser {
	return &pb.RBRoomUser{
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

func TiggerRBStrategy(user *data.User, d *data.BaseStrategy, chargeType int) bool {
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
