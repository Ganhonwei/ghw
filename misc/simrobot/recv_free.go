package main

import (
	"goserver/gen/pb"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
)

//.LHD

// ' 进入房间
func (r *Robot) recvLHFreeEnter(s2c *pb.LHFreeEnterRoomRsp) {
	var errcode = s2c.GetError()
	switch errcode {
	case pb.OK:
	default:
		glog.Infof("comein err -> %d", errcode)
		rbs.PutRobot(r, int32(pb.LHD))
		return
	}
	roominfo := s2c.GetRoominfo()
	r.gtype = roominfo.Gtype
	r.rtype = roominfo.Rtype
	r.dtype = roominfo.Dtype
	r.roomid = roominfo.Roomid
	r.lh.state = roominfo.State //状态
	// userinfo := s2c.GetUserinfo()
	// for _, v := range userinfo {
	// 	//只返回坐下玩家
	// 	if v.Userid == r.data.Userid {
	// 		glog.Debugf("comein user info -> %s", v.Userid)
	// 		break
	// 	}
	// }
	// r.gameStart(s2c.Roominfo.State)
}

// 人机策略
func (r *Robot) recvLHRobotStrategy(s2c *pb.LHRobotStrategyNtf) {
	r.lh.lhstrategy = s2c
	r.lh.betCount = 0
	// 下注轮数
	if len(s2c.BetTimes) <= 0 {
		r.lh.maxBetCount = 50
		return
	}
	r.lh.maxBetCount = int32(utils.RandMN(int(s2c.BetTimes[0]), int(s2c.BetTimes[1])))
}

// 状态变化
func (r *Robot) recvLHState(msg *pb.LHPushStateNtf) {
	r.lh.state = int32(msg.State)
	switch msg.State {
	case pb.STATE_READY:
		if r.lh.betCount >= r.lh.maxBetCount && !r.sim {
			// 离开
			msg := new(pb.LHLeaveReq)
			r.Sender(msg)
		}
	case pb.STATE_DEALING:
		r.data.LHNowChip = 0
	case pb.STATE_BET:
		// 下注
		r.lhdFreeBet()
		r.lh.betCount++
	}
}

// lhd下注响应
func (r *Robot) recvLHFreeBet(msg *pb.LHFreeBetRsp) {
	if msg.Error != pb.OK {
		// glog.Errorf("robot bet err:%v", msg.Error)
		return
	}
	// 下注
	r.lhdFreeBet()
}

// lhd离开
func (r *Robot) recvLHLeave(msg *pb.LHLeaveRsp) {
	if msg.Error == pb.OK {
		if msg.Userid == r.data.Userid {
			glog.Infof("lhd robot leave")
			rbs.PutRobot(r, int32(pb.LHD))
		}
	}
}

// lhd离开
func (r *Robot) recvLHLeaveNtf(msg *pb.LHLeaveNtf) {
	if msg.Userid == r.data.Userid {
		glog.Infof("lhd robot leave")
		rbs.PutRobot(r, int32(pb.LHD))
	}
}

// 7up

// ' 进入房间
func (r *Robot) recvUPFreeEnter(s2c *pb.UPFreeEnterRoomRsp) {
	var errcode = s2c.GetError()
	switch errcode {
	case pb.OK:
	default:
		glog.Infof("comein err -> %d", errcode)
		rbs.PutRobot(r, int32(pb.SEVEN))
		return
	}
	roominfo := s2c.GetRoominfo()
	r.gtype = roominfo.Gtype
	r.rtype = roominfo.Rtype
	r.dtype = roominfo.Dtype
	r.roomid = roominfo.Roomid
	r.lh.state = roominfo.State //状态
	// userinfo := s2c.GetUserinfo()
	// for _, v := range userinfo {
	// 	//只返回坐下玩家
	// 	if v.Userid == r.data.Userid {
	// 		glog.Debugf("comein user info -> %s", v.Userid)
	// 		break
	// 	}
	// }
	// r.gameStart(s2c.Roominfo.State)
}

// 人机策略
func (r *Robot) recvUPRobotStrategy(s2c *pb.UPRobotStrategyNtf) {
	r.up.lhstrategy = s2c
	r.up.betCount = 0
	// 下注轮数
	if len(s2c.BetTimes) <= 0 {
		r.up.maxBetCount = 50
		return
	}
	r.up.maxBetCount = int32(utils.RandMN(int(s2c.BetTimes[0]), int(s2c.BetTimes[1])))
}

// 状态变化
func (r *Robot) recvUPState(msg *pb.UPPushStateNtf) {
	r.up.state = int32(msg.State)
	switch msg.State {
	case pb.STATE_READY:
		if r.up.betCount >= r.up.maxBetCount && !r.sim {
			// 离开
			msg := new(pb.UPLeaveReq)
			r.Sender(msg)
		}
	case pb.STATE_DEALING:
		r.data.LHNowChip = 0
	case pb.STATE_BET:
		// 下注
		r.upFreeBet()
		r.up.betCount++
	}
}

// 7up下注响应
func (r *Robot) recvUPFreeBet(msg *pb.UPFreeBetRsp) {
	if msg.Error != pb.OK {
		// glog.Errorf("robot bet err:%v", msg.Error)
		return
	}
	// 下注
	r.upFreeBet()
}

// 7UP离开
func (r *Robot) recvUPLeave(msg *pb.UPLeaveRsp) {
	if msg.Error == pb.OK {
		if msg.Userid == r.data.Userid {
			glog.Infof("7up robot leave")
			rbs.PutRobot(r, int32(pb.SEVEN))
		}
	}
}

// 7up离开
func (r *Robot) recvUPLeaveNtf(msg *pb.UPLeaveNtf) {
	if msg.Userid == r.data.Userid {
		glog.Infof("7up robot leave")
		rbs.PutRobot(r, int32(pb.SEVEN))
	}
}

// crash

// ' 进入房间
func (r *Robot) recvCRASHFreeEnter(s2c *pb.CRASHEnterRoomRsp) {
	var errcode = s2c.GetError()
	switch errcode {
	case pb.OK:
	default:
		glog.Infof("comein err -> %d", errcode)
		rbs.PutRobot(r, int32(pb.CRASH))
		return
	}
	roominfo := s2c.GetRoominfo()
	r.gtype = roominfo.Gtype
	r.rtype = roominfo.Rtype
	r.dtype = roominfo.Dtype
	r.roomid = roominfo.Roomid
	r.crash.state = roominfo.State //状态
}

// 人机策略
func (r *Robot) recvCRASHRobotStrategy(s2c *pb.CRASHRobotStrategyNtf) {
	r.crash.cashstrategy = s2c
	r.crash.betCount = 0
	// 下注轮数
	if len(s2c.BetTimes) <= 0 {
		r.crash.maxBetCount = 50
		return
	}
	r.crash.maxBetCount = int32(utils.RandMN(int(s2c.BetTimes[0]), int(s2c.BetTimes[1])))
}

// 状态变化
func (r *Robot) recvCRASHState(msg *pb.CRASHPushStateNtf) {
	r.crash.state = int32(msg.State)
	switch msg.State {
	case pb.STATE_READY:
		if r.crash.betCount >= r.crash.maxBetCount && !r.sim {
			// 离开
			msg := new(pb.CRASHLeaveReq)
			r.Sender(msg)
		}
		r.crash.back = false
	case pb.STATE_DEALING:
		r.data.LHNowChip = 0
	case pb.STATE_BET:
		// 下注
		if r.sim {
			r.crash.simBack = utils.RandInt32N(101) + 100
		}
		r.crashFreeBet()
		r.crash.betCount++
	}
}

// crash下注响应
func (r *Robot) recvCRASHFreeBet(msg *pb.CRASHBetRsp) {
	if msg.Error != pb.OK {
		// glog.Errorf("robot bet err:%v", msg.Error)
		return
	}
	// 下注
	r.crashFreeBet()
}

// crash离开
func (r *Robot) recvCRASHLeave(msg *pb.CRASHLeaveRsp) {
	if msg.Error == pb.OK {
		if msg.Userid == r.data.Userid {
			glog.Infof("crash robot leave")
			rbs.PutRobot(r, int32(pb.CRASH))
		}
	}
}

// crash离开
func (r *Robot) recvCRASHLeaveNtf(msg *pb.CRASHLeaveNtf) {
	if msg.Userid == r.data.Userid {
		glog.Infof("crash robot leave")
		rbs.PutRobot(r, int32(pb.CRASH))
	}
}

// crash撤离
func (r *Robot) recvCRASHMultiple(msg *pb.CRASHMultipleNtf) {
	if r.sim {
		if 101 <= msg.Multiple && !r.crash.back {
			// if r.crash.simBack <= msg.Multiple && !r.crash.back {
			msg := new(pb.CRASHBackReq)
			r.Sender(msg)
			r.crash.back = true
		}
		return
	}
	if r.crash.cashstrategy == nil {
		return
	}
	crash := r.crash.cashstrategy.RobotCrash
	if msg.Multiple > crash[0] && !r.crash.back {
		// 计算是否逃离
		if utils.RandWan(crash[1]) {
			msg := new(pb.CRASHBackReq)
			r.Sender(msg)
			r.crash.back = true
		}
	}
}

// crash撤离
func (r *Robot) recvCRASHBack(msg *pb.CRASHBackRsp) {

}
