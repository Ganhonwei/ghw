package main

import (
	"goserver/gen/pb"
	"goserver/pkg/glog"
)

// hua

// sendJOKEREntryRoom 进入房间
func (c *Robot) sendJOKEREntryRoom() {
	glog.Debugf("enter roomid %s", c.roomid)
	c2s := new(pb.JOKERCoinEnterRoomReq)
	c2s.Gameid = c.gameid
	c2s.Roomid = c.roomid
	c.Sender(c2s)
}

// 准备请求
func (c *Robot) sendJOKERReady2Req() {
	c2s := new(pb.JOKERReady2Req)
	c.Sender(c2s)
}

// 看牌请求
func (c *Robot) sendJOKERCoinSeeReq() {
	c2s := new(pb.JOKERCoinSeeReq)
	c.SendDefer(c2s)
	c.see = true
}

// 跟注请求
func (c *Robot) sendJOKERCoinCallReq() {
	c2s := new(pb.JOKERCoinCallReq)
	c.SendDefer(c2s)
}

// 加注请求
func (c *Robot) sendJOKERCoinRaiseReq() {
	c2s := new(pb.JOKERCoinRaiseReq)
	c.SendDefer(c2s)
}

// 弃牌请求
func (c *Robot) sendJOKERCoinFoldReq() {
	c2s := new(pb.JOKERCoinFoldReq)
	c.SendDefer(c2s)
	c.alive = false
}

// 比牌请求
func (c *Robot) sendJOKERCoinBiReq() {
	c2s := new(pb.JOKERCoinBiReq)
	c.SendDefer(c2s)
}

// 回复比牌请求
func (c *Robot) sendJOKERCoinReplyBiReq(agress bool) {
	c2s := new(pb.JOKERCoinReplyBiReq)
	c2s.Agree = agress
	c.SendDefer(c2s)
}

// 机器人离开
func (c *Robot) sendJOKERLeaveReq() {
	c2s := new(pb.JOKERLeaveReq)
	c.SendDefer(c2s)
}
