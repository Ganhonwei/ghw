package handler

import (
	"goserver/gen/pb"
	"goserver/pkg/data"
)

// PackJHFreeUser 打包进入百人玩家基础数据
func PackRMFreeUser(p *data.User) *pb.RMFreeUser {
	return &pb.RMFreeUser{
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

// PackJHRoomBets 打包百人下注信息
func PackRMRoomBets(bets map[uint32]int64) (msg []*pb.RMRoomBets) {
	for k, v := range bets {
		msg2 := &pb.RMRoomBets{
			Seat: k,
			Bets: v,
		}
		msg = append(msg, msg2)
	}
	return
}

// PackJHFreeRoom 打包百人房间信息
func PackRMFreeRoom(d *data.DeskData) *pb.RMFreeRoom {
	return &pb.RMFreeRoom{
		Roomid: d.Rid,   //牌局id
		Gtype:  d.Gtype, //game type
		Rtype:  d.Rtype, //room type
		Dtype:  d.Dtype, //desk type
		Rname:  d.Rname, //room name
		Count:  d.Count, //当前房间限制玩家数量
		Ante:   d.Ante,  //房间底分
	}
}

// JHLeaveMsg 离开消息
func RMLeaveMsg(userid string, seat uint32) *pb.RMLeaveRsp {
	return &pb.RMLeaveRsp{
		Seat:   seat,
		Userid: userid,
	}
}

// BeJHDealerMsg 上下庄消息
func BeRMDealerMsg(state int32, num int64, dealer,
	userid, name, photo string) *pb.RMFreeDealerRsp {
	return &pb.RMFreeDealerRsp{
		State:    state,
		Coin:     uint32(num),
		Userid:   userid,
		Dealer:   dealer,
		Nickname: name,
		Photo:    photo,
	}
}

// PackJHCoinRoom 打包百人房间信息
func PackRMCoinRoom(d *data.DeskData) *pb.RMRoomData {
	return &pb.RMRoomData{
		Roomid:   d.Rid,                    //牌局id
		Gtype:    d.Gtype,                  //game type
		Rtype:    d.Rtype,                  //room type
		Dtype:    d.Dtype,                  //desk type
		Ltype:    d.Ltype,                  //level type
		Rname:    d.Rname,                  //room name
		Count:    d.Count,                  //当前房间限制玩家数量
		Ante:     uint32(d.Game.RM.Bottom), //房间底分
		Round:    d.Round,                  //
		Userid:   d.Cid,                    //
		Expire:   d.Expire,                 //
		Code:     d.Code,                   //
		Minimum:  d.Minimum,                //
		Maximum:  d.Maximum,                //
		Pub:      d.Pub,
		Mode:     d.Mode,
		Multiple: d.Multiple,
	}
}

// PackJHCoinUser 打包进入百人玩家基础数据
func PackRMCoinUser(p *data.User) *pb.RMRoomUser {
	return &pb.RMRoomUser{
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
		Fraction: p.GetFraction(),
		VipLv:    int32(p.Vip.Lv),
	}
}

// PackJHCreateMsg 创建房间消息
func PackRMCreateMsg(d *data.DeskData) *pb.RMCreateRoomRsp {
	msg := new(pb.RMCreateRoomRsp)
	msg.Data = PackRMCoinRoom(d)
	return msg
}

// PackJHRoomList 房间列表数据
func PackRMRoomList(arg *pb.GetRoomList,
	desks map[string]*data.DeskBase) *pb.JHRoomListRsp {
	msg := new(pb.JHRoomListRsp)
	for _, v := range desks {
		switch arg.Rtype {
		case int32(pb.ROOM_TYPE1): //私人
			if v.DeskData.Cid != arg.Userid && !v.DeskData.Pub {
				continue
			}
		}
		if v.DeskData.Gtype == arg.Gtype &&
			v.DeskData.Rtype == arg.Rtype {
			msg2 := PackJHCoinRoom(v.DeskData)
			msg2.Number = v.Number
			msg.List = append(msg.List, msg2)
		}
	}
	return msg
}
