package main

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"time"
)

//' 登录

// sendRegist 发送注册请求
// func (c *Robot) sendRegist() {
// 	c2s := new(pb.RegistReq)
// 	c2s.Phone = c.data.Phone
// 	c2s.Nickname = c.data.Nickname
// 	c2s.Smscode = "8888"
// 	passwd := cfg.Section("robot").Key("passwd").Value()
// 	c2s.Password = utils.Md5(passwd)
// 	c.Sender(c2s)
// }

// sendLogin 发送登录请求
func (c *Robot) sendLogin() {
	c2s := new(pb.LoginReq)
	c2s.Phone = c.data.Phone
	passwd := cfg.Section("simrobot").Key("passwd").Value()
	c2s.Password = utils.Md5(passwd)
	c.Sender(c2s)
}

// sendUserData 获取玩家数据
func (c *Robot) sendUserData() {
	c2s := new(pb.UserDataReq)
	c2s.Userid = c.data.Userid
	c.Sender(c2s)
}

// SendPing 心跳
func (c *Robot) sendPing() {
	c2s := new(pb.PingReq)
	c2s.Ctime = 1 //uint32(utils.Timestamp())
	c.Sender(c2s)
}

// 重置数据
func (r *Robot) Reset() {
	r.adjust = false
	r.seats = []uint32{}
	r.emoji = data.EmojiConfig{}
}

// 修改金币
func (r *Robot) adjustCoin() {
	var coin, diamond int
	if r.sim {
		coin = 0
		diamond = 100000000
	} else {
		// if r.max == -1 {
		// 	r.max = r.min * 10
		// }
		if r.gtype == int32(pb.LHD) ||
			r.gtype == int32(pb.SEVEN) ||
			r.gtype == int32(pb.CRASH) {
			coin = utils.RandMN(int(r.min), int(r.max))
		} else {
			coin = int(r.min * 1000)
		}
		// coin = utils.RandMN(int(r.min), int(r.max))
		diamond = 0
	}
	// r.adjust = true
	r.modifyCurrency(int64(coin), int64(diamond))
	glog.Infof("adjustCoin -> %d,%d", r.min, r.max)
}

// 修改货币
func (c *Robot) modifyCurrency(coin, diamond int64) {
	msg4 := &pb.ModifyCurrency{
		Userid:  c.data.Userid,
		Type:    int32(pb.LOG_TYPE44),
		Diamond: diamond,
		Coin:    coin,
	}
	rolePid.Tell(msg4)
}

// addCurrency 添加货币
func (c *Robot) addCurrency() {
	msg4 := &pb.PayCurrency{
		Userid: c.data.Userid,
		Type:   int32(pb.LOG_TYPE44),
		Coin:   200000,
	}
	rolePid.Tell(msg4)
}

// SendDefer 延迟发送
func (c *Robot) SendDefer(msg interface{}) {
	utils.Sleep(utils.RandIntN(2) + 2) //随机
	c.Sender(msg)
}

// SendDefer 延迟发送
func (c *Robot) SendDefer2(msg interface{}, during int32) {
	// utils.Sleep(utils.RandIntN(2) + 2) //随机
	<-time.After(time.Duration(during) * time.Millisecond)
	c.Sender(msg)
}

func GetContentAndDelay(prob int, weight []int, delay []int) (content string, ret_delay int) {
	// if true {
	if utils.RandWan(int32(prob)) {
		var choices []utils.Choice
		for i, v := range weight {
			choices = append(choices, utils.Choice{Weight: v, Item: i})
		}
		_c, _ := utils.WeightedChoice(choices)
		i := _c.Item.(int)
		content = fmt.Sprintf("%d", i+1)
		ret_delay = utils.RandMN(delay[0], delay[1])
	}
	return
}

func GetContentAndDelay2(prob []int, weight []int, delay []int, win, lose uint32) (to uint32, content string, ret_delay int) {
	i := utils.RandSliceIndex(prob)
	switch i {
	case 0: //不发
		return
	case 1: //发赢
		to = win
		face_idx := utils.RandSliceIndex(weight)
		content = fmt.Sprintf("%d", face_idx+1)
		ret_delay = utils.RandMN(delay[0], delay[1])
	case 2: //发输
		to = lose
		face_idx := utils.RandSliceIndex(weight)
		content = fmt.Sprintf("%d", face_idx+1)
		ret_delay = utils.RandMN(delay[0], delay[1])
	}
	return
}

// 上桌发送
func (c *Robot) EmojiDeskSend(to uint32) {
	if !c.emoji.IsValid() {
		return
	}
	content, delay := GetContentAndDelay(c.emoji.DeskSendProb, c.emoji.DeskSendWeight, c.emoji.DeskSendDelay)
	c.SendEmoji(to, content, delay)
}

// 游戏开始
func (c *Robot) EmojiGameStart(to uint32) {
	if !c.emoji.IsValid() {
		return
	}

	content, delay := GetContentAndDelay(c.emoji.GameStartProb, c.emoji.GameStartWeight, c.emoji.GameStartDelay)
	c.SendEmoji(to, content, delay)
}

func (c *Robot) EmojiWaitLong(to uint32) {
	if !c.emoji.IsValid() {
		return
	}
	content, delay := GetContentAndDelay(c.emoji.WaitLongProb, c.emoji.WaitLongWeight, c.emoji.WaitLongDelay)
	c.SendEmoji(to, content, delay)
}

// 回复表情
func (c *Robot) EmojiReply(to uint32) {
	if !c.emoji.IsValid() {
		return
	}
	content, delay := GetContentAndDelay(c.emoji.ReplyProb, c.emoji.ReplyWeight, c.emoji.ReplyDelay)
	c.SendEmoji(to, content, delay)
}

// 其他玩家看牌后
func (c *Robot) EmojiOtherSee(to uint32) {
	if !c.emoji.IsValid() {
		return
	}
	content, delay := GetContentAndDelay(c.emoji.OtherSeeProb, c.emoji.OtherSeeWeight, c.emoji.OtherSeeDelay)
	c.SendEmoji(to, content, delay)
}

// 其他玩家加注
func (c *Robot) EmojiOtherRaise(to uint32) {
	if !c.emoji.IsValid() {
		return
	}
	content, delay := GetContentAndDelay(c.emoji.OtherRaiseProb, c.emoji.OtherRaiseWeight, c.emoji.OtherRaiseDelay)
	c.SendEmoji(to, content, delay)
}

// 其他玩家比牌
func (c *Robot) EmojiOtherBi(win, lose uint32) {
	if !c.emoji.IsValid() {
		return
	}
	to, content, delay := GetContentAndDelay2(c.emoji.OtherBiProb, c.emoji.OtherBiWeight, c.emoji.OtherBiDelay, win, lose)
	c.SendEmoji(to, content, delay)
}

// 自己比牌赢
func (c *Robot) EmojiSelfBiWin(lose uint32) {
	if !c.emoji.IsValid() {
		return
	}
	content, delay := GetContentAndDelay(c.emoji.SelfBiWinProb, c.emoji.SelfBiWinWeight, c.emoji.SelfBiWinDelay)
	c.SendEmoji(lose, content, delay)
}

// 摸上家牌
func (c *Robot) EmojiMoPreCard(seat uint32) {
	if !c.emoji.IsValid() {
		return
	}
	content, delay := GetContentAndDelay(c.emoji.MoPreCardProb, c.emoji.MoPreCardWeight, c.emoji.MoPreCardDelay)
	c.SendEmoji(seat, content, delay)
}

// 其他玩家出牌
func (c *Robot) EmojiOtherChuCard(seat uint32) {
	if !c.emoji.IsValid() {
		return
	}
	content, delay := GetContentAndDelay(c.emoji.OtherChuCardProb, c.emoji.OtherChuCardWeight, c.emoji.OtherChuCardDelay)
	c.SendEmoji(seat, content, delay)
}

// 发送表情
func (c *Robot) SendEmoji(to uint32, content string, delay int) {
	if to == 0 || content == "" || delay == 0 {
		return
	}

	//判断cd
	if !c.CheckEmojiCD() {
		return
	}

	c2s := new(pb.ChatTextReq)
	c2s.To = to
	c2s.Content = content
	c.SendDefer2(c2s, int32(delay))
}

// sendNNLeave 离开
// func (c *Robot) sendNNLeave() {
// 	c2s := new(pb.CNNLeave)
// 	c.Sender(c2s)
// }

// sendNNReady 准备
// func (c *Robot) sendNNReady() {
// 	c2s := new(pb.CNNReady)
// 	c2s.Ready = true
// 	c.SendDefer(c2s)
// }

// sendNNDealer 抢庄
// func (c *Robot) sendNNDealer() {
// 	c2s := new(pb.CNNDealer)
// 	if utils.RandIntN(100) > 50 {
// 		c2s.Dealer = true
// 		c2s.Num = uint32(utils.RandIntN(2) + 1)
// 	}
// 	c.SendDefer(c2s)
// }

// sendNNiu 提交
// func (c *Robot) sendNNiu() {
// 	c2s := new(pb.CNNiu)
// 	c.SendDefer(c2s)
// }

// sendNNStandup 玩家离坐
// func (c *Robot) sendNNStandup() {
// 	c.sendNNLeave()
// 	utils.Sleep(2)
// 	c.Close() //下线
// }

// sendNNEntryRoom 进入房间
// func (c *Robot) sendNNEntryRoom() {
// 	switch c.rtype {
// 	case int32(pb.ROOM_TYPE0):
// 		glog.Debugf("enter roomid %s", c.roomid)
// 		c2s := new(pb.CNNCoinEnterRoom)
// 		c2s.Id = c.roomid
// 		c.SendDefer(c2s)
// 	case int32(pb.ROOM_TYPE1):
// 		glog.Debugf("enter room code %s", c.code)
// 		c2s := new(pb.CNNEnterRoom)
// 		c2s.Code = c.code
// 		c.SendDefer(c2s)
// 	case int32(pb.ROOM_TYPE2):
// 		c2s := new(pb.CNNFreeEnterRoom)
// 		c.SendDefer(c2s)
// 	}
// }

// sendNNBet 玩家下注
// func (c *Robot) sendNNBet() {
// 	switch c.rtype {
// 	case int32(pb.ROOM_TYPE0):
// 		c2s := new(pb.CNNBet)
// 		c2s.Seatbet = c.seat
// 		c2s.Value = uint32(utils.RandIntN(10))
// 		c.SendDefer(c2s)
// 	case int32(pb.ROOM_TYPE1):
// 		c2s := new(pb.CNNBet)
// 		c2s.Seatbet = c.seat
// 		c2s.Value = 1
// 		c.SendDefer(c2s)
// 	case int32(pb.ROOM_TYPE2):
// 		//随机下注次数
// 		c.bits = uint32(utils.RandInt32N(20) + 1)
// 		c.bitNum = uint32(utils.RandInt32N(7) * 5000)
// 		c.sendNNFreeBet() //下注
// 	}
// }

// sendNNBet 玩家下注
// func (c *Robot) sendNNFreeBet() {
// 	//不同游戏位置不同
// 	var seats = []uint32{2, 3, 4, 5, 6, 7, 8, 9}
// 	var bets = []uint32{100, 500, 1000, 5000, 10000}
// 	var coin uint32 = uint32(c.data.Coin) / 4
// 	var max int
// 	for i := 4; i >= 0; i-- {
// 		if coin >= bets[i] {
// 			max = i
// 			break
// 		}
// 	}
// 	var v int
// 	switch max {
// 	case 0:
// 		v = max
// 	default:
// 		v = utils.RandIntN(max) //随机
// 	}
// 	var k = utils.RandIntN(len(seats)) //随机
// 	c2s := &pb.CNNFreeBet{
// 		Value: bets[v],
// 		Seat:  seats[k],
// 	}
// 	c.SendDefer(c2s)
// }

//.

//' ebg

// sendEBLeave 离开
// func (c *Robot) sendEBLeave() {
// 	c2s := new(pb.CEBLeave)
// 	c.Sender(c2s)
// }

// sendEBReady 准备
// func (c *Robot) sendEBReady() {
// 	c2s := new(pb.CEBReady)
// 	c2s.Ready = true
// 	c.SendDefer(c2s)
// }

// sendEBDealer 抢庄
// func (c *Robot) sendEBDealer() {
// 	c2s := new(pb.CEBDealer)
// 	if utils.RandIntN(100) > 50 {
// 		c2s.Dealer = true
// 		c2s.Num = uint32(utils.RandIntN(2) + 1)
// 	}
// 	c.SendDefer(c2s)
// }

// sendEBiu 提交
// func (c *Robot) sendEBiu() {
// 	c2s := new(pb.CEBiu)
// 	c.SendDefer(c2s)
// }

// sendEBStandup 玩家离坐
// func (c *Robot) sendEBStandup() {
// 	c.sendEBLeave()
// 	utils.Sleep(2)
// 	c.Close() //下线
// }

// sendEBEntryRoom 进入房间
// func (c *Robot) sendEBEntryRoom() {
// 	switch c.rtype {
// 	case int32(pb.ROOM_TYPE0):
// 		glog.Debugf("enter roomid %s", c.roomid)
// 		c2s := new(pb.CEBCoinEnterRoom)
// 		c2s.Id = c.roomid
// 		c.SendDefer(c2s)
// 	case int32(pb.ROOM_TYPE1):
// 		glog.Debugf("enter room code %s", c.code)
// 		c2s := new(pb.CEBEnterRoom)
// 		c2s.Code = c.code
// 		c.SendDefer(c2s)
// 	case int32(pb.ROOM_TYPE2):
// 		c2s := new(pb.CEBFreeEnterRoom)
// 		c.SendDefer(c2s)
// 	}
// }

// sendEBBet 玩家下注
// func (c *Robot) sendEBBet() {
// 	switch c.rtype {
// 	case int32(pb.ROOM_TYPE0):
// 		c2s := new(pb.CEBBet)
// 		c2s.Seatbet = c.seat
// 		c2s.Value = uint32(utils.RandIntN(10))
// 		c.SendDefer(c2s)
// 	case int32(pb.ROOM_TYPE1):
// 		c2s := new(pb.CEBBet)
// 		c2s.Seatbet = c.seat
// 		c2s.Value = 1
// 		c.SendDefer(c2s)
// 	case int32(pb.ROOM_TYPE2):
// 		//随机下注次数
// 		c.bits = uint32(utils.RandInt32N(20) + 1)
// 		c.bitNum = uint32(utils.RandInt32N(7) * 5000)
// 		c.sendEBFreeBet() //下注
// 	}
// }

// sendEBFreeBet 玩家下注
// func (c *Robot) sendEBFreeBet() {
// 	//不同游戏位置不同
// 	var seats = []uint32{2, 3, 4, 5, 6, 7, 8, 9}
// 	var bets = []uint32{100, 500, 1000, 5000, 10000}
// 	var coin uint32 = uint32(c.data.Coin) / 4
// 	var max int
// 	for i := 4; i >= 0; i-- {
// 		if coin >= bets[i] {
// 			max = i
// 			break
// 		}
// 	}
// 	var v int
// 	switch max {
// 	case 0:
// 		v = max
// 	default:
// 		v = utils.RandIntN(max) //随机
// 	}
// 	var k = utils.RandIntN(len(seats)) //随机
// 	c2s := &pb.CEBFreeBet{
// 		Value: bets[v],
// 		Seat:  seats[k],
// 	}
// 	c.SendDefer(c2s)
// }

//.

// vim: set foldmethod=marker foldmarker=//',//.:
