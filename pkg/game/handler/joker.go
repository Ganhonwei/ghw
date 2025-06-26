package handler

import (
	"goserver/gen/pb"
	"goserver/pkg/data"
)

// PackJOKERFreeUser 打包进入百人玩家基础数据
func PackJOKERFreeUser(p *data.User) *pb.JOKERFreeUser {
	return &pb.JOKERFreeUser{
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

// PackJOKERRoomBets 打包百人下注信息
func PackJOKERRoomBets(bets map[uint32]int64) (msg []*pb.JOKERRoomBets) {
	for k, v := range bets {
		msg2 := &pb.JOKERRoomBets{
			Seat: k,
			Bets: v,
		}
		msg = append(msg, msg2)
	}
	return
}

// PackJOKERFreeRoom 打包百人房间信息
func PackJOKERFreeRoom(d *data.DeskData) *pb.JOKERFreeRoom {
	return &pb.JOKERFreeRoom{
		Roomid: d.Rid,   //牌局id
		Gtype:  d.Gtype, //game type
		Rtype:  d.Rtype, //room type
		Dtype:  d.Dtype, //desk type
		Rname:  d.Rname, //room name
		Count:  d.Count, //当前房间限制玩家数量
		Ante:   d.Ante,  //房间底分
	}
}

// JOKERLeaveMsg 离开消息
func JOKERLeaveMsg(userid string, seat uint32) *pb.JOKERLeaveRsp {
	return &pb.JOKERLeaveRsp{
		Seat:   seat,
		Userid: userid,
	}
}

// BeJOKERDealerMsg 上下庄消息
func BeJOKERDealerMsg(state int32, num int64, dealer,
	userid, name, photo string) *pb.JOKERFreeDealerRsp {
	return &pb.JOKERFreeDealerRsp{
		State:    state,
		Coin:     uint32(num),
		Userid:   userid,
		Dealer:   dealer,
		Nickname: name,
		Photo:    photo,
	}
}

// PackJOKERCoinRoom 打包百人房间信息
func PackJOKERCoinRoom(d *data.DeskData) *pb.JOKERRoomData {
	return &pb.JOKERRoomData{
		Roomid:   d.Rid,                       //牌局id
		Gtype:    d.Gtype,                     //game type
		Rtype:    d.Rtype,                     //room type
		Dtype:    d.Dtype,                     //desk type
		Ltype:    d.Ltype,                     //level type
		Rname:    d.Rname,                     //room name
		Count:    d.Count,                     //当前房间限制玩家数量
		Ante:     uint32(d.Game.JOKER.Bottom), //房间底分
		Round:    d.Round,                     //
		Userid:   d.Cid,                       //
		Expire:   d.Expire,                    //
		Code:     d.Code,                      //
		Minimum:  d.Minimum,                   //
		Maximum:  d.Maximum,                   //
		Pub:      d.Pub,
		Mode:     d.Mode,
		Multiple: d.Multiple,
	}
}

// PackJOKERCoinUser 打包进入百人玩家基础数据
func PackJOKERCoinUser(p *data.User) *pb.JOKERRoomUser {
	return &pb.JOKERRoomUser{
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

// PackJOKERCreateMsg 创建房间消息
func PackJOKERCreateMsg(d *data.DeskData) *pb.JOKERCreateRoomRsp {
	msg := new(pb.JOKERCreateRoomRsp)
	msg.Data = PackJOKERCoinRoom(d)
	return msg
}

// PackJOKERRoomList 房间列表数据
func PackJOKERRoomList(arg *pb.GetRoomList,
	desks map[string]*data.DeskBase) *pb.JOKERRoomListRsp {
	msg := new(pb.JOKERRoomListRsp)
	for _, v := range desks {
		switch arg.Rtype {
		case int32(pb.ROOM_TYPE1): //私人
			if v.DeskData.Cid != arg.Userid && !v.DeskData.Pub {
				continue
			}
		}
		if v.DeskData.Gtype == arg.Gtype &&
			v.DeskData.Rtype == arg.Rtype {
			msg2 := PackJOKERCoinRoom(v.DeskData)
			msg2.Number = v.Number
			msg.List = append(msg.List, msg2)
		}
	}
	return msg
}
