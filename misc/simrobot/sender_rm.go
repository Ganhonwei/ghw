package main

import (
	"goserver/gen/pb"
	"goserver/pkg/glog"
)

// 进入房间请求
func (c *Robot) sendRMEntryRoom() {
	glog.Debugf("enter roomid %s", c.roomid)
	c2s := new(pb.RMCoinEnterRoomReq)
	c2s.Gameid = c.gameid
	c2s.Roomid = c.roomid
	c.Sender(c2s)
}

// 准备请求
func (c *Robot) sendRMReady2Req() {
	c2s := new(pb.RMReady2Req)
	c.Sender(c2s)
}

// 摸牌请求
func (c *Robot) sendRMDrawCardReq(area uint32) {
	c2s := new(pb.RMDrawCardReq)
	c2s.Area = area
	c.SendDefer(c2s)
}

// 出牌请求
func (c *Robot) sendRMDiscardReq(card uint32) {
	c2s := new(pb.RMDiscardReq)
	c2s.Card = card
	c.SendDefer(c2s)
}

// 胡牌请求
func (c *Robot) sendRMFinishReq(card uint32, groups [][]uint32) {
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
func (c *Robot) sendRMDeclareReq(cards [][]uint32) {
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
func (c *Robot) sendRMLeaveReq() {
	c2s := new(pb.RMLeaveReq)
	c.SendDefer(c2s)
}

// sort请求
func (c *Robot) sendRMSortReq(cards [][]uint32) {
	c2s := new(pb.RMSortReq)
	for _, v := range cards {
		info := &pb.RMSortInfo{
			Cards: v,
		}
		c2s.Infos = append(c2s.Infos, info)
	}
	c.SendDefer(c2s)
}
