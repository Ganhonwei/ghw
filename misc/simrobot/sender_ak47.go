package main

import (
	"goserver/gen/pb"
	"goserver/pkg/glog"
)

// hua

// sendAK47EntryRoom 进入房间
func (c *Robot) sendAK47EntryRoom() {
	glog.Debugf("enter roomid %s", c.roomid)
	c2s := new(pb.AK47CoinEnterRoomReq)
	c2s.Gameid = c.gameid
	c2s.Roomid = c.roomid
	c.Sender(c2s)
}

// 准备请求
func (c *Robot) sendAK47Ready2Req() {
	c2s := new(pb.AK47Ready2Req)
	c.Sender(c2s)
}

// 看牌请求
func (c *Robot) sendAK47CoinSeeReq() {
	c2s := new(pb.AK47CoinSeeReq)
	c.SendDefer(c2s)
	c.see = true
}

// 跟注请求
func (c *Robot) sendAK47CoinCallReq() {
	c2s := new(pb.AK47CoinCallReq)
	c.SendDefer(c2s)
}

// 加注请求
func (c *Robot) sendAK47CoinRaiseReq() {
	c2s := new(pb.AK47CoinRaiseReq)
	c.SendDefer(c2s)
}

// 弃牌请求
func (c *Robot) sendAK47CoinFoldReq() {
	c2s := new(pb.AK47CoinFoldReq)
	c.SendDefer(c2s)
	c.alive = false
}

// 比牌请求
func (c *Robot) sendAK47CoinBiReq() {
	c2s := new(pb.AK47CoinBiReq)
	c.SendDefer(c2s)
}

// 回复比牌请求
func (c *Robot) sendAK47CoinReplyBiReq(agress bool) {
	c2s := new(pb.AK47CoinReplyBiReq)
	c2s.Agree = agress
	c.SendDefer(c2s)
}

// 机器人离开
func (c *Robot) sendAK47LeaveReq() {
	c2s := new(pb.AK47LeaveReq)
	c.SendDefer(c2s)
}
