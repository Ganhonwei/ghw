package gate

import (
	"goserver/gen/pb"
	"goserver/pkg/glog"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// 进入房间
func (rs *RoleActor) LotteryEnterReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.LotteryEnterReq)
	glog.Debugf("LotteryEnterReq %#v", arg)
	rs.enterLottery(arg, ctx)
}

// 下注
func (rs *RoleActor) LotteryBetReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.LotteryBetReq)
	// glog.Debugf("LotteryBetReq %#v", arg)
	if rs.cpPid == nil {
		rsp := new(pb.LotteryBetRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	value := arg.GetValue()
	if value <= 0 {
		rsp := new(pb.LotteryBetRsp)
		rsp.Error = pb.OperateError
		rs.Send(rsp)
		return
	}
	if rs.Diamond < int64(arg.Value) {
		rsp := new(pb.LotteryBetRsp)
		rsp.Error = pb.NotEnoughCoin
		rs.Send(rsp)
		return
	}
	rs.cpPid.Request(arg, ctx.Self())
}

// 离开
func (rs *RoleActor) LotteryLeaveReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.LotteryLeaveReq)
	glog.Debugf("LotteryLeaveReq %#v", arg)
	if rs.cpPid == nil {
		rsp := new(pb.LotteryLeaveRsp)
		rsp.Error = pb.NotInRoom
		rs.Send(rsp)
		return
	}
	rs.cpPid.Request(arg, ctx.Self())
}

// 进房间
func (rs *RoleActor) enterLottery(arg *pb.LotteryEnterReq, ctx actor.Context) {
	if rs.cpPid != nil && rs.enterRoom(ctx) {
		// 直接进入
		return
	}
	//获取游戏服节点或者房间进程
	msg := new(pb.MatchDesk)
	//TODO 优化查找规则
	msg.Name = cfg.Section("game.lottery").Name()
	msg.Gtype = int32(pb.LOTTERY) //彩票
	msg.Dtype = rs.getFreeDtype()
	msg.Rtype = int32(pb.ROOM_TYPE2)
	rs.selectDesk(msg, ctx)
}

// 进入房间
func (rs *RoleActor) enterRoom(ctx actor.Context) bool {
	msg := new(pb.EnterDesk)
	msg.Gtype = int32(pb.LOTTERY) //彩票
	msg.Dtype = rs.getFreeDtype()
	msg.Rtype = int32(pb.ROOM_TYPE2)
	if !rs.enterDeskMsg(msg, ctx) {
		glog.Errorf("userid %s enter faild %s",
			rs.User.GetUserid(), rs.gamePid.String())
		return false
	}
	rs.cpPid.Request(msg, ctx.Self())
	return true
}
