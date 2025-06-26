package handler

import (
	"goserver/gen/pb"
	"goserver/pkg/data"
)

// PackLHFreeUser 打包进入百人玩家基础数据
func PackLHFreeUser(p *data.User) *pb.LHFreeUser {
	return &pb.LHFreeUser{
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

// PackLHRoomBets 打包百人下注信息
func PackLHRoomBets(bets map[uint32]int64) (msg []*pb.LHRoomBets) {
	for k, v := range bets {
		msg2 := &pb.LHRoomBets{
			Seat: k,
			Bets: v,
		}
		msg = append(msg, msg2)
	}
	return
}

// PackLHFreeRoom 打包百人房间信息
func PackLHFreeRoom(d *data.DeskData) *pb.LHFreeRoom {
	return &pb.LHFreeRoom{
		Roomid: d.Rid,                //牌局id
		Gtype:  d.Gtype,              //game type
		Rtype:  d.Rtype,              //room type
		Dtype:  d.Dtype,              //desk type
		Rname:  d.Rname,              //room name
		Count:  d.Count,              //当前房间限制玩家数量
		Chip:   d.Game.LHD.ChipLimit, //房间下注筹码
	}
}

// LHLeaveMsg 离开消息
func LHLeaveMsg(userid string, seat uint32) *pb.LHLeaveRsp {
	return &pb.LHLeaveRsp{
		Seat:   seat,
		Userid: userid,
	}
}

// LhdBeDealerMsg 上下庄消息
// func LhdBeDealerMsg(state int32, num, carry int64, dealer,
// 	userid, name, photo string) *pb.LHFreeDealerRsp {
// 	return &pb.LHFreeDealerRsp{
// 		State:    state,
// 		Coin:     uint32(num),
// 		Userid:   userid,
// 		Dealer:   dealer,
// 		Nickname: name,
// 		Photo:    photo,
// 		Carry:    uint32(carry),
// 	}
// }

// PackLHCoinRoom 打包百人房间信息
func PackLHCoinRoom(d *data.DeskData) *pb.LHRoomData {
	return &pb.LHRoomData{
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

// PackLHCoinUser 打包进入百人玩家基础数据
func PackLHCoinUser(p *data.User) *pb.LHRoomUser {
	return &pb.LHRoomUser{
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

func TiggerLHDStrategy(user *data.User, d *data.BaseStrategy) bool {
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
