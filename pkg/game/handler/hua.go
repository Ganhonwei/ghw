package handler

import (
	"goserver/gen/pb"
	"goserver/pkg/data"
)

// PackJHFreeUser 打包进入百人玩家基础数据
func PackJHFreeUser(p *data.User) *pb.JHFreeUser {
	return &pb.JHFreeUser{
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
func PackJHRoomBets(bets map[uint32]int64) (msg []*pb.JHRoomBets) {
	for k, v := range bets {
		msg2 := &pb.JHRoomBets{
			Seat: k,
			Bets: v,
		}
		msg = append(msg, msg2)
	}
	return
}

// PackJHFreeRoom 打包百人房间信息
func PackJHFreeRoom(d *data.DeskData) *pb.JHFreeRoom {
	return &pb.JHFreeRoom{
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
func JHLeaveMsg(userid string, seat uint32) *pb.JHLeaveRsp {
	return &pb.JHLeaveRsp{
		Seat:   seat,
		Userid: userid,
	}
}

// BeJHDealerMsg 上下庄消息
func BeJHDealerMsg(state int32, num int64, dealer,
	userid, name, photo string) *pb.JHFreeDealerRsp {
	return &pb.JHFreeDealerRsp{
		State:    state,
		Coin:     uint32(num),
		Userid:   userid,
		Dealer:   dealer,
		Nickname: name,
		Photo:    photo,
	}
}

// PackJHCoinRoom 打包百人房间信息
func PackJHCoinRoom(d *data.DeskData) *pb.JHRoomData {
	return &pb.JHRoomData{
		Roomid:   d.Rid,                    //牌局id
		Gtype:    d.Gtype,                  //game type
		Rtype:    d.Rtype,                  //room type
		Dtype:    d.Dtype,                  //desk type
		Ltype:    d.Ltype,                  //level type
		Rname:    d.Rname,                  //room name
		Count:    d.Count,                  //当前房间限制玩家数量
		Ante:     uint32(d.Game.TP.Bottom), //房间底分
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
func PackJHCoinUser(p *data.User) *pb.JHRoomUser {
	return &pb.JHRoomUser{
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
func PackJHCreateMsg(d *data.DeskData) *pb.JHCreateRoomRsp {
	msg := new(pb.JHCreateRoomRsp)
	msg.Data = PackJHCoinRoom(d)
	return msg
}

// PackJHRoomList 房间列表数据
func PackJHRoomList(arg *pb.GetRoomList,
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
