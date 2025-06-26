package main

import (
	"goserver/gen/pb"
	"goserver/pkg/glog"
)

// hua

// sendJHEntryRoom 进入房间
func (c *Robot) sendJHEntryRoom() {
	glog.Debugf("enter roomid %s", c.roomid)
	c2s := new(pb.JHCoinEnterRoomReq)
	c2s.Gameid = c.gameid
	c2s.Roomid = c.roomid
	c.Sender(c2s)
}

// 准备请求
func (c *Robot) sendJHReady2Req() {
	c2s := new(pb.JHReady2Req)
	c.Sender(c2s)
}

// 看牌请求
func (c *Robot) sendJHCoinSeeReq() {
	c2s := new(pb.JHCoinSeeReq)
	c.SendDefer(c2s)
	c.see = true
}

// 跟注请求
func (c *Robot) sendJHCoinCallReq() {
	c2s := new(pb.JHCoinCallReq)
	c.SendDefer(c2s)
}

// 加注请求
func (c *Robot) sendJHCoinRaiseReq() {
	c2s := new(pb.JHCoinRaiseReq)
	c.SendDefer(c2s)
}

// 弃牌请求
func (c *Robot) sendJHCoinFoldReq() {
	c2s := new(pb.JHCoinFoldReq)
	c.SendDefer(c2s)
	c.alive = false
}

// 比牌请求
func (c *Robot) sendJHCoinBiReq() {
	c2s := new(pb.JHCoinBiReq)
	c.SendDefer(c2s)
}

// 回复比牌请求
func (c *Robot) sendJHCoinReplyBiReq(agress bool) {
	c2s := new(pb.JHCoinReplyBiReq)
	c2s.Agree = agress
	c.SendDefer(c2s)
}

// 机器人离开
func (c *Robot) sendJHLeaveReq() {
	c2s := new(pb.JHLeaveReq)
	c.SendDefer(c2s)
}
