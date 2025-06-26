package handler

import (
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/utils"
	"math"
)

// PackAK47FreeUser 打包进入百人玩家基础数据
func PackAK47FreeUser(p *data.User) *pb.AK47FreeUser {
	return &pb.AK47FreeUser{
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

// PackAK47RoomBets 打包百人下注信息
func PackAK47RoomBets(bets map[uint32]int64) (msg []*pb.AK47RoomBets) {
	for k, v := range bets {
		msg2 := &pb.AK47RoomBets{
			Seat: k,
			Bets: v,
		}
		msg = append(msg, msg2)
	}
	return
}

// PackAK47FreeRoom 打包百人房间信息
func PackAK47FreeRoom(d *data.DeskData) *pb.AK47FreeRoom {
	return &pb.AK47FreeRoom{
		Roomid: d.Rid,   //牌局id
		Gtype:  d.Gtype, //game type
		Rtype:  d.Rtype, //room type
		Dtype:  d.Dtype, //desk type
		Rname:  d.Rname, //room name
		Count:  d.Count, //当前房间限制玩家数量
		Ante:   d.Ante,  //房间底分
	}
}

// AK47LeaveMsg 离开消息
func AK47LeaveMsg(userid string, seat uint32) *pb.AK47LeaveRsp {
	return &pb.AK47LeaveRsp{
		Seat:   seat,
		Userid: userid,
	}
}

// BeAK47DealerMsg 上下庄消息
func BeAK47DealerMsg(state int32, num int64, dealer,
	userid, name, photo string) *pb.AK47FreeDealerRsp {
	return &pb.AK47FreeDealerRsp{
		State:    state,
		Coin:     uint32(num),
		Userid:   userid,
		Dealer:   dealer,
		Nickname: name,
		Photo:    photo,
	}
}

// PackAK47CoinRoom 打包百人房间信息
func PackAK47CoinRoom(d *data.DeskData) *pb.AK47RoomData {
	return &pb.AK47RoomData{
		Roomid:   d.Rid,                      //牌局id
		Gtype:    d.Gtype,                    //game type
		Rtype:    d.Rtype,                    //room type
		Dtype:    d.Dtype,                    //desk type
		Ltype:    d.Ltype,                    //level type
		Rname:    d.Rname,                    //room name
		Count:    d.Count,                    //当前房间限制玩家数量
		Ante:     uint32(d.Game.AK47.Bottom), //房间底分
		Round:    d.Round,                    //
		Userid:   d.Cid,                      //
		Expire:   d.Expire,                   //
		Code:     d.Code,                     //
		Minimum:  d.Minimum,                  //
		Maximum:  d.Maximum,                  //
		Pub:      d.Pub,
		Mode:     d.Mode,
		Multiple: d.Multiple,
	}
}

// PackAK47CoinUser 打包进入百人玩家基础数据
func PackAK47CoinUser(p *data.User) *pb.AK47RoomUser {
	return &pb.AK47RoomUser{
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

// PackAK47CreateMsg 创建房间消息
func PackAK47CreateMsg(d *data.DeskData) *pb.AK47CreateRoomRsp {
	msg := new(pb.AK47CreateRoomRsp)
	msg.Data = PackAK47CoinRoom(d)
	return msg
}

// PackAK47RoomList 房间列表数据
func PackAK47RoomList(arg *pb.GetRoomList,
	desks map[string]*data.DeskBase) *pb.AK47RoomListRsp {
	msg := new(pb.AK47RoomListRsp)
	for _, v := range desks {
		switch arg.Rtype {
		case int32(pb.ROOM_TYPE1): //私人
			if v.DeskData.Cid != arg.Userid && !v.DeskData.Pub {
				continue
			}
		}
		if v.DeskData.Gtype == arg.Gtype &&
			v.DeskData.Rtype == arg.Rtype {
			msg2 := PackAK47CoinRoom(v.DeskData)
			msg2.Number = v.Number
			msg.List = append(msg.List, msg2)
		}
	}
	return msg
}

// 获取万能牌数量
func GetWildCardNum(ct data.AK47GameCardType, robot bool, maxNum int) int {
	// 先随机多少张万能牌
	var choices []utils.Choice
	if !robot {
		choices = append(choices, utils.Choice{Weight: ct.PlayerWildCardWeight[0], Item: 0})
		choices = append(choices, utils.Choice{Weight: ct.PlayerWildCardWeight[1], Item: 1})
		choices = append(choices, utils.Choice{Weight: ct.PlayerWildCardWeight[2], Item: 2})
		choices = append(choices, utils.Choice{Weight: ct.PlayerWildCardWeight[3], Item: 3})
	} else {
		choices = append(choices, utils.Choice{Weight: ct.RobotWildCardWeight[0], Item: 0})
		choices = append(choices, utils.Choice{Weight: ct.RobotWildCardWeight[1], Item: 1})
		choices = append(choices, utils.Choice{Weight: ct.RobotWildCardWeight[2], Item: 2})
		choices = append(choices, utils.Choice{Weight: ct.RobotWildCardWeight[3], Item: 3})
	}

	c, _ := utils.WeightedChoice(choices)
	num := c.Item.(int)
	return int(math.Min(float64(num), float64(maxNum)))
}
