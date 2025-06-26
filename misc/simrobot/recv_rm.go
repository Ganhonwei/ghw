package main

import (
	"goserver/gen/pb"
	"goserver/pkg/game/algo"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
)

// rm
// 进入房间响应
func (r *Robot) recvRMCoinEnterRoomRsp(s2c *pb.RMCoinEnterRoomRsp) {
	var errcode = s2c.GetError()
	switch errcode {
	case pb.OK:
	default:
		glog.Errorf("comein err -> %d", errcode)
		rbs.PutRobot(r, int32(pb.RUMMY))
		return
	}
	roominfo := s2c.GetRoominfo()
	r.gtype = roominfo.Gtype
	r.rtype = roominfo.Rtype
	r.dtype = roominfo.Dtype
	r.roomid = roominfo.Roomid
	userinfo := s2c.GetUserinfo()
	for _, v := range userinfo {
		//只返回坐下玩家
		if v.Userid == r.data.Userid {
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
func (r *Robot) recvRMPushStateNtf(msg *pb.RMPushStateNtf) {
	if msg.State == int32(pb.STATE_DEALING) {
		// r.see = false
		// r.alive = true
		// r.jhstrategy = nil
		// r.cards = []uint32{}
		// r.state = 0
		// r.wildCard = 0
		// r.qiCard = 0
		r.sendRMReady2Req()
	} else if msg.State == int32(pb.STATE_BET) { //下注
		r.EmojiGameStart(r.GetRandSeat())
	}
}

// 动作状态推送
func (r *Robot) recvRMPushActStateNtf(msg *pb.RMPushActStateNtf) {
	if r.alive && msg.Seat == r.seat { //自己操作
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

// 庄家推送
func (r *Robot) recvRMPushDealerNtf(msg *pb.RMPushDealerNtf) {
	r.alive = true
	r.cards = []uint32{}
	r.wildCard = msg.WildCard
	r.qiCard = msg.QiCard
}

// 玩家手牌推送
func (r *Robot) recvRMPushCardsNtf(msg *pb.RMPushCardsNtf) {
	r.cards = msg.Cards

	//摆牌
	var left_cards []uint32
	r.sortCards, left_cards = algo.SortCards(r.cards, r.wildCard)
	r.sortCards = append(r.sortCards, left_cards)
	r.sendRMSortReq(r.sortCards)
}

// 玩家出牌推送
func (r *Robot) recvRMDiscardNtf(msg *pb.RMDiscardNtf) {
	r.qiCard = msg.Card
	r.preSeat = msg.Seat

	//人机表情发送
	r.EmojiOtherChuCard(msg.Seat)
}

// 摸牌响应
func (r *Robot) recvRMDrawCardRsp(msg *pb.RMDrawCardRsp) {
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

// 出牌响应
func (r *Robot) recvRMDiscardRsp(msg *pb.RMDiscardRsp) {
	if msg.Error == pb.OK {
		r.cards = algo.RemoveCard(r.cards, msg.Card)
		//摆牌
		var left_cards []uint32
		r.sortCards, left_cards = algo.SortCards(r.cards, r.wildCard)
		r.sortCards = append(r.sortCards, left_cards)
		r.sendRMSortReq(r.sortCards)
	}
}

// finish 推送
func (r *Robot) recvRMFinishNtf(msg *pb.RMFinishNtf) {
	if msg.Huseat != r.seat {
		var left_cards []uint32
		r.sortCards, left_cards = algo.SortCards(r.cards, r.wildCard)
		r.sortCards = append(r.sortCards, left_cards)
		r.sendRMDeclareReq(r.sortCards)
	}
}

// 游戏结束推送
func (r *Robot) recvRMCoinGameoverNtf(msg *pb.RMCoinGameoverNtf) {
	if !r.sim && utils.RandWan(5000) {
		r.sendRMLeaveReq()
	}
}

// 等待太久
func (r *Robot) RMCoinWaitTooLongNtf(msg *pb.RMCoinWaitTooLongNtf) {
	r.EmojiWaitLong(msg.Seat)
}

// 声明牌响应
// func (r *Robot) recvRMDeclareRsp(msg *pb.RMDeclareRsp) {

// }

// 胡牌响应
// func (r *Robot) recvRMFinishRsp(msg *pb.RMFinishRsp) {

// }

// 离开房间响应
func (r *Robot) recvRMLeaveRsp(msg *pb.RMLeaveRsp) {
	if msg.Error == pb.OK {
		rbs.PutRobot(r, int32(pb.RUMMY))
	}
}
