package robot

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/game/algo"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"goserver/pkg/zlog"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// 离开消息
func (a *RoleActor) RMLeaveRsp(ctx actor.Context) {
	ntf := ctx.Message().(*pb.RMLeaveRsp)
	if ntf.Userid == a.Userid {
		// 关闭节点
		glog.Debugf("RMLeaveRsp %v", ntf)
		stop := new(pb.ServeStop)
		a.pid.Tell(stop)
	}
}

// 离开消息
func (a *RoleActor) RMLeaveNtf(ctx actor.Context) {
	ntf := ctx.Message().(*pb.RMLeaveNtf)
	if ntf.Userid == a.Userid {
		// 关闭节点
		glog.Debugf("RMLeaveNtf %v", ntf)
		stop := new(pb.ServeStop)
		a.pid.Tell(stop)
	}
}

func (a *RoleActor) RMCoinEnterRoomRsp(ctx actor.Context) {
	res := ctx.Message().(*pb.RMCoinEnterRoomRsp)
	glog.Debugf("RMCoinEnterRoomRsp %v", res)
	if res.Error != pb.OK {
		glog.Infof("RMCoinEnterRoomRsp enter err:%s", res.Error.String())
		stop := new(pb.ServeStop)
		a.pid.Tell(stop)
	}
}

// 进入或匹配桌子
func (rs *RoleActor) enterRummyMatchDesk(arg *pb.RobotMsg) *pb.MatchDesk {
	//获取游戏服节点或者房间进程
	msg := new(pb.MatchDesk)
	//TODO 优化查找规则
	msg.Name = cfg.Section("game.rummy").Name()
	msg.Gtype = int32(pb.RUMMY) //rummy
	msg.Roomid = arg.Roomid
	msg.Gameid = arg.Gameid
	return msg
}

// 进入或匹配桌子
func (rs *RoleActor) enterRummy2MatchDesk(arg *pb.RobotMsg) *pb.MatchDesk {
	//获取游戏服节点或者房间进程
	msg := new(pb.MatchDesk)
	//TODO 优化查找规则
	msg.Name = cfg.Section("game.rummy_2").Name()
	msg.Gtype = int32(pb.RUMMY2) //rummy
	msg.Roomid = arg.Roomid
	msg.Gameid = arg.Gameid
	return msg
}

// 进入房间响应
func (r *RoleActor) recvRMCoinEnterRoomRsp(ctx actor.Context) {
	s2c := ctx.Message().(*pb.RMCoinEnterRoomRsp)
	var errcode = s2c.GetError()
	switch errcode {
	case pb.OK:
	default:
		// 关闭节点
		glog.Errorf("enter ak47 fail user:%s,err:%v ", r.Userid, errcode)
		r.closeRs()
		return
	}
	roominfo := s2c.GetRoominfo()
	r.gtype = roominfo.Gtype
	// r.rtype = roominfo.Rtype
	// r.dtype = roominfo.Dtype
	// r.roomid = roominfo.Roomid
	userinfo := s2c.GetUserinfo()
	for _, v := range userinfo {
		//只返回坐下玩家
		if v.Userid == r.Userid {
			glog.Debugf("comein user info -> %s", v.Userid)
			r.seat = v.Seat
			break
		} else { //记录其他玩家座位号
			r.seats = append(r.seats, v.Seat)
		}
	}
	//上桌发送表情
	r.EmojiDeskSend(r.GetRandSeat())
	// r.gameStart(s2c.Roominfo.State)
}

// 房间状态推送
func (r *RoleActor) recvRMPushStateNtf(ctx actor.Context) {
	msg := ctx.Message().(*pb.RMPushStateNtf)
	if msg.State == int32(pb.STATE_DEALING) {
		r.sendRMReady2Req()
	} else if msg.State == int32(pb.STATE_BET) { //下注
		r.EmojiGameStart(r.GetRandSeat())
	}
}

func (r *RoleActor) recvRMPushActStateNtf(ctx actor.Context) {
	switch r.gtype {
	case int32(pb.RUMMY2):
		r.recvRMPushActStateNtf2(ctx)
	default:
		r.recvRMPushActStateNtf1(ctx)
	}
}

// 动作状态推送
func (r *RoleActor) recvRMPushActStateNtf1(ctx actor.Context) {
	msg := ctx.Message().(*pb.RMPushActStateNtf)
	if r.alive && msg.Seat == r.seat { //自己操作
		//判断是否drop
		sort := algo.SortCards4(r.cards, r.wildCard)
		score := algo.CalcScore(sort, r.wildCard)
		if !r.core && msg.Turn == 0 && score >= 60 && utils.RandWan(1000) {
			r.sendRMDropReq(sort)
			return
		} else if !r.core && msg.Turn == 1 && score >= 80 && utils.RandWan(500) {
			r.sendRMDropReq(sort)
			return
		}

		//摸牌
		if algo.IsLaizi(r.qiCard, r.wildCard) { //癞子必摸
			r.sendRMDrawCardReq(2)
		} else if algo.NeedCard(r.cards, r.qiCard, r.wildCard) { //判断弃牌是否需要
			r.sendRMDrawCardReq(2)
			r.EmojiMoPreCard(r.preSeat)
		} else { //摸牌库
			r.sendRMDrawCardReq(1)
		}
	}
}

func (r *RoleActor) recvRMPushActStateNtf2(ctx actor.Context) {
	msg := ctx.Message().(*pb.RMPushActStateNtf)
	if r.alive && msg.Seat == r.seat { //自己操作
		zlog.Infof("%s rm robot start draw:", r.Userid)
		// 人机弃牌
		if msg.Rdrop {
			zlog.Infof("%s rm robot drop card", r.Userid)
			r.sendRMDropReq([][]uint32{r.cards})
			return
		}
		// 摸排逻辑
		var area uint32
		latestCard := r.qiCards[len(r.qiCards)-1]
		if algo.IsCardWild(latestCard, r.wildCard) || algo.IsCardJoker(latestCard) {
			fmt.Println("首张翻拍是癞子牌，吃牌")
			area = 2
		} else {
			now := time.Now().Unix()
			area = uint32(algo.GetTouchType(r.cards, r.qiCards, r.wildCard))
			delay := time.Now().Unix() - now
			if delay > 10 {
				zlog.Warningf("GetTouchType timeout: %s, %ds, cards=%s, qiCards=%s, wildCard=%d", r.Userid, delay, algo.CardsString(r.cards), algo.CardsString(r.qiCards), r.wildCard)
			}
		}
		zlog.Infof("%s rm robot end draw: area=%d", r.Userid, area)
		r.sendRMDrawCardReq(area)
		r.EmojiMoPreCard(r.preSeat)
	}
}

// 摸牌
func (r *RoleActor) recvRMDrawCardNtf(ctx actor.Context) {
	switch r.gtype {
	case int32(pb.RUMMY2):
		r.recvRMDrawCardNtf2(ctx)
	default:
		r.recvRMDrawCardNtf1(ctx)
	}
}

func (r *RoleActor) recvRMDrawCardNtf1(ctx actor.Context) {

}

func (r *RoleActor) recvRMDrawCardNtf2(ctx actor.Context) {
	msg := ctx.Message().(*pb.RMDrawCardNtf)
	if msg.Area == 2 { // 摸弃牌堆
		length := len(r.qiCards)
		if length > 0 {
			r.qiCards = r.qiCards[:length-1]
		}
	}
}

// 庄家推送
func (r *RoleActor) recvRMPushDealerNtf(ctx actor.Context) {
	switch r.gtype {
	case int32(pb.RUMMY2):
		r.recvRMPushDealerNtf2(ctx)
	default:
		r.recvRMPushDealerNtf1(ctx)
	}
}

// 庄家推送
func (r *RoleActor) recvRMPushDealerNtf1(ctx actor.Context) {
	msg := ctx.Message().(*pb.RMPushDealerNtf)
	r.alive = true
	r.cards = []uint32{}
	r.wildCard = msg.WildCard
	r.qiCard = msg.QiCard
	r.core = false

	if msg.CoreRobot == r.seat {
		r.core = true
	}
}

// 庄家推送
func (r *RoleActor) recvRMPushDealerNtf2(ctx actor.Context) {
	msg := ctx.Message().(*pb.RMPushDealerNtf)
	r.alive = true
	r.cards = []uint32{}
	r.wildCard = msg.WildCard
	r.qiCard = msg.QiCard
	r.qiCards = []uint32{} // reset
	r.qiCards = append(r.qiCards, msg.QiCard)
	r.faceUpCard = []uint32{} // reset
	r.faceUpCard = append(r.faceUpCard, msg.QiCard)
	r.core = false

	if msg.CoreRobot == r.seat {
		r.core = true
	}
}

// 玩家手牌推送
func (r *RoleActor) recvRMPushCardsNtf(ctx actor.Context) {
	switch r.gtype {
	case int32(pb.RUMMY2):
		r.recvRMPushCardsNtf2(ctx)
	default:
		r.recvRMPushCardsNtf1(ctx)
	}
}

func (r *RoleActor) recvRMPushCardsNtf1(ctx actor.Context) {
	msg := ctx.Message().(*pb.RMPushCardsNtf)
	r.cards = msg.Cards

	//摆牌
	var left_cards []uint32
	r.sortCards, left_cards = algo.SortCards(r.cards, r.wildCard)
	r.sortCards = append(r.sortCards, left_cards)
	r.sendRMSortReq(r.sortCards)
}

func (r *RoleActor) recvRMPushCardsNtf2(ctx actor.Context) {
	msg := ctx.Message().(*pb.RMPushCardsNtf)
	r.cards = msg.Cards

	//摆牌
	zlog.Infof("%s rm robot RMPushCardsNtf %#v:", r.Userid, msg)
	r.sortCards = algo.GroupTheCards(r.cards, r.wildCard)
	zlog.Infof("%s rm robot sort card end %#v:", r.Userid, algo.CardGroupsString(r.sortCards))
	r.sendRMSortReq(r.sortCards)
}

// 玩家出牌推送
func (r *RoleActor) recvRMDiscardNtf(ctx actor.Context) {
	switch r.gtype {
	case int32(pb.RUMMY2):
		r.recvRMDiscardNtf2(ctx)
	default:
		r.recvRMDiscardNtf1(ctx)
	}
}

// 玩家出牌推送
func (r *RoleActor) recvRMDiscardNtf1(ctx actor.Context) {
	msg := ctx.Message().(*pb.RMDiscardNtf)
	r.qiCard = msg.Card
	r.preSeat = msg.Seat

	//人机表情发送
	r.EmojiOtherChuCard(msg.Seat)
}

// 玩家出牌推送
func (r *RoleActor) recvRMDiscardNtf2(ctx actor.Context) {
	msg := ctx.Message().(*pb.RMDiscardNtf)
	r.qiCard = msg.Card
	r.preSeat = msg.Seat
	r.qiCards = append(r.qiCards, msg.Card)
	r.faceUpCard = append(r.faceUpCard, msg.Card)

	//人机表情发送
	r.EmojiOtherChuCard(msg.Seat)
}

// 摸牌响应
func (r *RoleActor) recvRMDrawCardRsp(ctx actor.Context) {
	switch r.gtype {
	case int32(pb.RUMMY2):
		r.recvRMDrawCardRsp2(ctx)
	default:
		r.recvRMDrawCardRsp1(ctx)
	}
}
func (r *RoleActor) recvRMDrawCardRsp1(ctx actor.Context) {
	msg := ctx.Message().(*pb.RMDrawCardRsp)
	if msg.Error == pb.OK {
		r.cards = append(r.cards, msg.Card)
		finish, discard_card := algo.GetOneCardToDiscard(r.cards, r.wildCard)
		if finish {
			r.cards = algo.RemoveCard(r.cards, discard_card)
			r.sortCards, _ = algo.SortCards(r.cards, r.wildCard)
			r.sendRMFinishReq(discard_card, r.sortCards)
		} else {
			r.sendRMDiscardReq(discard_card)
		}
	} else {
		glog.Errorf("recvRMDrawCardRsp error %#v", msg)
	}
}

// rummy 双人摸牌响应
func (r *RoleActor) recvRMDrawCardRsp2(ctx actor.Context) {
	msg := ctx.Message().(*pb.RMDrawCardRsp)
	if msg.Error != pb.OK {
		glog.Errorf("recvRMDrawCardRsp error %#v", msg)
		return
	}

	// 人机弃牌
	if msg.Rdrop {
		r.sendRMDropReq([][]uint32{r.cards})
		return
	}

	// todo 摸牌次数大于10000判断是否弃牌
	// if p.touchedCount > 10000 {
	//   algo.ShouldDrop()
	// }

	zlog.Infof("%s rm robot out card start", r.Userid)
	r.cards = append(r.cards, msg.Card)
	now := time.Now().Unix()
	discard_card, card_group, finish := algo.LetsOutCard(r.cards, r.faceUpCard, r.wildCard)
	delay := time.Now().Unix() - now
	if delay > 10 {
		zlog.Warningf("%s rm robot LetsOutCard timeout: %s, %ds, %s", r.Userid, delay, algo.CardsString(r.cards))
	}
	zlog.Infof("%s rm robot out card finish: finish=%v, out=%s, group=%s", r.Userid, finish, algo.CardsString([]uint32{discard_card}), card_group.String())
	// finish, discard_card := algo.GetOneCardToDiscard(r.cards, r.wildCard)
	if finish {
		r.cards = algo.RemoveCard(r.cards, discard_card)
		r.sortCards = card_group.BaseGroups
		if len(card_group.LeftCards) > 0 {
			r.sortCards = append(r.sortCards, card_group.LeftCards)
		}
		// r.sortCards, _ = algo.SortCards(r.cards, r.wildCard)
		r.sendRMFinishReq(discard_card, r.sortCards)
	} else {
		r.sendRMDiscardReq(discard_card)
	}
}

// 出牌响应
func (r *RoleActor) recvRMDiscardRsp(ctx actor.Context) {
	switch r.gtype {
	case int32(pb.RUMMY2):
		r.recvRMDiscardRsp2(ctx)
	default:
		r.recvRMDiscardRsp1(ctx)
	}
}

func (r *RoleActor) recvRMDiscardRsp1(ctx actor.Context) {
	msg := ctx.Message().(*pb.RMDiscardRsp)
	if msg.Error == pb.OK {
		r.cards = algo.RemoveCard(r.cards, msg.Card)
		//摆牌
		var left_cards []uint32
		r.sortCards, left_cards = algo.SortCards(r.cards, r.wildCard)
		r.sortCards = append(r.sortCards, left_cards)
		r.sendRMSortReq(r.sortCards)
	}
}

func (r *RoleActor) recvRMDiscardRsp2(ctx actor.Context) {
	msg := ctx.Message().(*pb.RMDiscardRsp)
	if msg.Error == pb.OK {
		r.cards = algo.RemoveCard(r.cards, msg.Card)
		//摆牌
		r.sortCards = algo.GroupTheCards(r.cards, r.wildCard)
		// var left_cards []uint32
		// r.sortCards, left_cards = algo.SortCards(r.cards, r.wildCard)
		// r.sortCards = append(r.sortCards, left_cards)
		r.sendRMSortReq(r.sortCards)
	} else {
		zlog.Errorf("%s discord error: %#v", r.Userid, msg)
	}
}

// finish 推送
func (r *RoleActor) recvRMFinishNtf(ctx actor.Context) {
	switch r.gtype {
	case int32(pb.RUMMY2):
		r.recvRMFinishNtf2(ctx)
	default:
		r.recvRMFinishNtf1(ctx)
	}
}

func (r *RoleActor) recvRMFinishNtf1(ctx actor.Context) {
	msg := ctx.Message().(*pb.RMFinishNtf)
	if msg.Huseat != r.seat {
		var left_cards []uint32
		r.sortCards, left_cards = algo.SortCards(r.cards, r.wildCard)
		r.sortCards = append(r.sortCards, left_cards)
		r.sendRMDeclareReq(r.sortCards)
	}
}

// finish rm双人推送
func (r *RoleActor) recvRMFinishNtf2(ctx actor.Context) {
	msg := ctx.Message().(*pb.RMFinishNtf)
	if msg.Huseat != r.seat {
		zlog.Infof("%s rm robot RMFinishNtf start", r.Userid)
		r.sortCards = algo.GroupTheCards(r.cards, r.wildCard)
		zlog.Infof("%s rm robot RMFinishNtf end, %s", r.Userid, algo.CardGroupsString(r.sortCards))
		r.sendRMDeclareReq(r.sortCards)
	}
}

// 游戏结束推送
func (r *RoleActor) recvRMCoinGameoverNtf(ctx actor.Context) {
	if utils.RandWan(5000) {
		r.sendRMLeaveReq()
	}
}

// 等待太久
func (r *RoleActor) RMCoinWaitTooLongNtf(ctx actor.Context) {
	msg := ctx.Message().(*pb.RMCoinWaitTooLongNtf)
	r.EmojiWaitLong(msg.Seat)
}

// 声明牌响应
// func (r *Robot) recvRMDeclareRsp(msg *pb.RMDeclareRsp) {

// }

// 胡牌响应
// func (r *Robot) recvRMFinishRsp(msg *pb.RMFinishRsp) {

// }

// 离开房间响应
func (r *RoleActor) recvRMLeaveRsp(ctx actor.Context) {
	ntf := ctx.Message().(*pb.RMLeaveRsp)
	glog.Debugf("RMLeaveRsp %v", ntf)
	if ntf.Userid == r.Userid {
		// 关闭节点
		stop := new(pb.ServeStop)
		r.pid.Tell(stop)
	}
}

// 离开房间响应
func (r *RoleActor) recvRMLeaveNtf(ctx actor.Context) {
	ntf := ctx.Message().(*pb.RMLeaveNtf)
	glog.Debugf("RMLeaveNtf %v", ntf)
	if ntf.Userid == r.Userid {
		// 关闭节点
		stop := new(pb.ServeStop)
		r.pid.Tell(stop)
	}
}

// 准备请求
func (c *RoleActor) sendRMReady2Req() {
	c2s := new(pb.RMReady2Req)
	c.Sender(c2s)
}

// drop请求
func (c *RoleActor) sendRMDropReq(cards [][]uint32) {
	c.alive = false
	c2s := new(pb.RMDropReq)
	for _, v := range cards {
		info := &pb.RMSortCard{
			Cards: v,
		}
		c2s.Sort = append(c2s.Sort, info)
	}
	c.SendDefer(c2s)
}

// 摸牌请求
func (c *RoleActor) sendRMDrawCardReq(area uint32) {
	c2s := new(pb.RMDrawCardReq)
	c2s.Area = area
	c.SendDefer(c2s)
}

// 出牌请求
func (c *RoleActor) sendRMDiscardReq(card uint32) {
	c2s := new(pb.RMDiscardReq)
	c2s.Card = card
	c.SendDefer(c2s)
}

// 胡牌请求
func (c *RoleActor) sendRMFinishReq(card uint32, groups [][]uint32) {
	c2s := new(pb.RMFinishReq)
	c2s.Card = card
	for _, group := range groups {
		info := &pb.RMFinishInfo{
			Cards: group,
		}
		c2s.Infos = append(c2s.Infos, info)
	}
	c.SendDefer(c2s)
}

// 声明牌请求
func (c *RoleActor) sendRMDeclareReq(cards [][]uint32) {
	c2s := new(pb.RMDeclareReq)
	for _, v := range cards {
		info := &pb.RMDeclareInfo{
			Cards: v,
		}
		c2s.Infos = append(c2s.Infos, info)
	}
	c.SendDefer(c2s)
}

// 离开请求
func (c *RoleActor) sendRMLeaveReq() {
	c2s := new(pb.RMLeaveReq)
	c.SendDefer(c2s)
}

// sort请求
func (c *RoleActor) sendRMSortReq(cards [][]uint32) {
	c2s := new(pb.RMSortReq)
	for _, v := range cards {
		info := &pb.RMSortInfo{
			Cards: v,
		}
		c2s.Infos = append(c2s.Infos, info)
	}
	c.SendDefer(c2s)
}
