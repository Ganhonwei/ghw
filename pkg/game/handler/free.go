package handler

import (
	"goserver/pkg/utils"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// PackNNFreeUser 打包进入百人玩家基础数据
// func PackNNFreeUser(p *data.User) *pb.NNFreeUser {
// 	return &pb.NNFreeUser{
// 		Userid:   p.GetUserid(),
// 		Nickname: p.GetNickname(),
// 		Phone:    p.GetPhone(),
// 		Sex:      p.GetSex(),
// 		Photo:    p.GetPhoto(),
// 		Coin:     p.GetCoin(),
// 		Diamond:  p.GetDiamond(),
// 	}
// }

// // PackNNRoomBets 打包百人下注信息
// func PackNNRoomBets(bets map[uint32]int64) (msg []*pb.NNRoomBets) {
// 	for k, v := range bets {
// 		msg2 := &pb.NNRoomBets{
// 			Seat: k,
// 			Bets: v,
// 		}
// 		msg = append(msg, msg2)
// 	}
// 	return
// }

// // PackNNFreeRoom 打包百人房间信息
// func PackNNFreeRoom(d *data.DeskData) *pb.NNFreeRoom {
// 	return &pb.NNFreeRoom{
// 		Roomid: d.Rid,   //牌局id
// 		Gtype:  d.Gtype, //game type
// 		Rtype:  d.Rtype, //room type
// 		Dtype:  d.Dtype, //desk type
// 		Rname:  d.Rname, //room name
// 		Count:  d.Count, //当前房间限制玩家数量
// 		Ante:   d.Ante,  //房间底分
// 	}
// }

// // NNLeaveMsg 离开消息
// func NNLeaveMsg(userid string, seat uint32) *pb.SNNLeave {
// 	return &pb.SNNLeave{
// 		Seat:   seat,
// 		Userid: userid,
// 	}
// }

// BeDealerMsg 上下庄消息
// func BeDealerMsg(state int32, num, carry int64, dealer,
// 	userid, name, photo string) *pb.SNNFreeDealer {
// 	return &pb.SNNFreeDealer{
// 		State:    state,
// 		Coin:     uint32(num),
// 		Userid:   userid,
// 		Dealer:   dealer,
// 		Nickname: name,
// 		Photo:    photo,
// 		Carry:    uint32(carry),
// 	}
// }

// // PackNNCoinRoom 打包百人房间信息
// func PackNNCoinRoom(d *data.DeskData) *pb.NNRoomData {
// 	return &pb.NNRoomData{
// 		Roomid:   d.Rid,     //牌局id
// 		Gtype:    d.Gtype,   //game type
// 		Rtype:    d.Rtype,   //room type
// 		Dtype:    d.Dtype,   //desk type
// 		Ltype:    d.Ltype,   //level type
// 		Rname:    d.Rname,   //room name
// 		Count:    d.Count,   //当前房间限制玩家数量
// 		Ante:     d.Ante,    //房间底分
// 		Round:    d.Round,   //
// 		Userid:   d.Cid,     //
// 		Expire:   d.Expire,  //
// 		Code:     d.Code,    //
// 		Minimum:  d.Minimum, //
// 		Maximum:  d.Maximum, //
// 		Pub:      d.Pub,
// 		Mode:     d.Mode,
// 		Multiple: d.Multiple,
// 	}
// }

// // PackNNCoinUser 打包进入百人玩家基础数据
// func PackNNCoinUser(p *data.User) *pb.NNRoomUser {
// 	return &pb.NNRoomUser{
// 		Userid:   p.GetUserid(),
// 		Nickname: p.GetNickname(),
// 		Phone:    p.GetPhone(),
// 		Sex:      p.GetSex(),
// 		Photo:    p.GetPhoto(),
// 		Coin:     p.GetCoin(),
// 		Diamond:  p.GetDiamond(),
// 		Lat:      p.Lat,
// 		Lng:      p.Lng,
// 		Address:  p.Address,
// 		Sign:     p.Sign,
// 	}
// }

// // PackNNCreateMsg 创建房间消息
// func PackNNCreateMsg(d *data.DeskData) *pb.SNNCreateRoom {
// 	msg := new(pb.SNNCreateRoom)
// 	msg.Data = PackNNCoinRoom(d)
// 	return msg
// }

// // PackNNRoomList 房间列表数据
// func PackNNRoomList(arg *pb.GetRoomList,
// 	desks map[string]*data.DeskBase) *pb.SNNRoomList {
// 	msg := new(pb.SNNRoomList)
// 	for _, v := range desks {
// 		switch arg.Rtype {
// 		case int32(pb.ROOM_TYPE1): //私人
// 			if v.DeskData.Cid != arg.Userid && !v.DeskData.Pub {
// 				continue
// 			}
// 		}
// 		if v.DeskData.Gtype == arg.Gtype &&
// 			v.DeskData.Rtype == arg.Rtype {
// 			msg2 := PackNNCoinRoom(v.DeskData)
// 			msg2.Number = v.Number
// 			msg.List = append(msg.List, msg2)
// 		}
// 	}
// 	return msg
// }

const (
	p0 int64 = 1000000000000000000 // math.Pow10(18)
	p1 int64 = 1000000000000000    // math.Pow10(15)
)

// 生成一个爆炸倍数
func CrashBoomMultiple(r *rand.Rand) (multiple int32) {
	multiple = 101
	for {
		var a, b int32
		if multiple == 101 {
			a, b = 97, 101
		} else {
			a, b = multiple-1, multiple
		}
		p := decimal.NewFromInt32(a * 1000).
			Div(decimal.NewFromInt32(b)).
			Mul(decimal.NewFromInt(p1)).
			IntPart()
		rate := p0 - p + 1
		// fmt.Println(multiple, p, rate)

		if r.Int63n(p0) < rate {
			break
		}
		multiple += 1
	}
	return
}

// 生成一个爆炸倍数
func _crashBoomMultiple(r *rand.Rand) (multiple int32) {
	multiple = 101
	m := math.Pow10(18)
	var rate int64
	for {
		if multiple == 101 {
			rate = int64(m - (97.0 / 101.0 * m))
		} else {
			rate = int64(m) - int64(float64(multiple-1)/float64(multiple)*m)
		}
		if r.Int63n(int64(m)) < rate {
			break
		}
		multiple += 1
	}
	return
}

// 根据倍数计算爆炸时间
func CrashBoomDuration(multiple int32) (ts time.Duration) {
	mt := float64(multiple) / 100
	r := math.Log(mt)
	sec := r / 0.0821
	ts = time.Duration(math.Floor((sec * 1000 * 1000))) * time.Microsecond
	return
}

// 根据时间差计算当前飞行倍数 multiple=E^(0.0821*duration)
func CrashTimeMultiple(takeoff, now time.Time) (multiple int32) {
	cs := now.Sub(takeoff).Microseconds()
	n := (float64(cs*821) / 10000) / (1000 * 1000)
	multiple = int32(math.Ceil(math.Pow(math.E, n) * 100))
	return
}

func CrashIsBoom(mulpitle int32, p, q float64, r *rand.Rand) bool {
	if mulpitle < 101 {
		return false
	}
	// b := float64(mulpitle) / 100
	// y := (100*b*b - 99*b - 1) / (b - 1)
	y := float64(mulpitle) + 100*p - 99
	m := math.Pow10(18)
	// glog.Warningf("mulpitle:%.2f,y:%.18f", b, 1/(y-1))
	// 爆炸概率
	boompro := int64((1/(y-1) + q) * m)
	result := r.Int63n(int64(m)) + 1
	return result <= boompro
}

func CrashInstantBang(mulpitle int32, pro int32) bool {
	if mulpitle > 100 {
		return false
	}

	return utils.RandWan(pro)
}

func AnalysisFreeStrategyParam(param string, val float64) bool {
	if param == "" {
		return true
	}

	strs := strings.Split(param, ",")
	if len(strs) < 2 {
		return true
	}

	endIndex := strings.Index(param, ",")
	left := param[1:endIndex]                 // 左边的值
	right := param[endIndex+1 : len(param)-1] // 右边的值

	head := param[0]
	foot := param[len(param)-1]

	if left == "" && right != "" {
		if string(foot) == ")" {
			return val < utils.Float64(right)
		} else {
			return val <= utils.Float64(right)
		}
	} else if left != "" && right == "" {
		if string(head) == "(" {
			return val > utils.Float64(left)
		} else {
			return val >= utils.Float64(left)
		}
	} else if left != "" && right != "" {
		if string(head) == "(" && val <= utils.Float64(left) {
			return false
		} else if string(head) == "[" && val < utils.Float64(left) {
			return false
		}
		if string(foot) == ")" && val >= utils.Float64(right) {
			return false
		} else if string(foot) == "]" && val > utils.Float64(right) {
			return false
		}
	}

	return true
}
