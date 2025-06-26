package robot

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"sync"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
)

var seed int32 = 1
var UidMutex sync.Mutex

// 召唤机器人
func (a *RobotActor) RobotMsg(ctx actor.Context) {
	arg := ctx.Message().(*pb.RobotMsg)
	glog.Debugf("RobotMsg %v", arg)
	if arg.LimitTime <= 0 {
		a.callRobot(arg)
		return
	}
	now := utils.BsonNow().UnixMilli()
	limitTime := arg.LimitTime - now
	if limitTime < 100 {
		// 小于100ms就不召唤了
		return
	}
	for i := 0; i < int(arg.Num); i++ {
		call := utils.RandInt64N(limitTime-100) + 100
		d := RobotData{
			Rtype:  arg.Rtype,
			Ltype:  arg.Ltype,
			Gameid: arg.Gameid,
			Roomid: arg.Roomid,
			EnvBet: arg.EnvBet,
			Min:    arg.Min,
			Max:    arg.Max,
			Gtype:  arg.Gtype,
			Emoji:  arg.Emoji,
		}
		a.CallQeuee = append(a.CallQeuee, CallRobot{Msg: d, CallTime: now + call})
	}
}

func (a *RobotActor) RobotLeave(ctx actor.Context) {
	arg := ctx.Message().(*pb.RobotLeave)
	glog.Debugf("RobotLeave %v", arg)
	if r, ok := a.rules[arg.Userid]; ok {
		glog.Infof("recycle robot,id:%s", arg.Userid)
		if len(a.idleRoles) < 2<<7 {
			a.idleRoles = append(a.idleRoles, r.Pid)
		} else {
			// 空闲队列满了,关掉actor
			r.Pid.Tell(new(pb.ServeClose))
		}
	}
	delete(a.rules, arg.Userid)
}

// 召唤机器人
func (a *RobotActor) callRobot(arg *pb.RobotMsg) {
	userid := fmt.Sprintf("%d", generateOrderId())
	var pid *actor.PID
	if len(a.idleRoles) > 0 {
		// 获取一个空闲的actor
		pid = a.idleRoles[0]
		a.idleRoles = a.idleRoles[1:]
	} else {
		// 生成一个uid
		pid = a.spawnRole(userid)
	}
	a.rules[userid] = RobotPid{Roomid: arg.Roomid, Gtype: arg.Gtype, Pid: pid}
	arg.Userid = userid
	pid.Tell(arg)
}

// 新玩家
func (a *RobotActor) spawnRole(userid string) *actor.PID {
	newRole := NewRole()
	newRole.dbmsPid = a.dbmsPid
	newRole.roomPid = a.roomPid
	rolePid := newRole.initRs()
	newRole.pid = rolePid
	msg1 := &pb.ServeStart{
		Message: userid,
	}
	rolePid.Tell(msg1)
	return rolePid
}

// 生成订单号 时间32位|序列号16
func generateOrderId() int64 {
	defer UidMutex.Unlock()
	UidMutex.Lock()
	now := time.Now().UnixMilli()
	seed += 1
	if seed >= (1<<31 - 1) {
		seed = 1
	}
	return (int64(1001&0xFFFF) << 43) | ((now << 17) | int64(seed&0xFFFF))
}

// 随机获取n条机器人基本信息
func (a *RobotActor) RobotBaseGet(ctx actor.Context) {
	arg := ctx.Message().(*pb.RobotBaseGet)
	glog.Debugf("RobotBaseGet %v", arg)
	rsp := new(pb.RobotBaseGeted)
	rsp.VipRobots = make(map[int32]*pb.RobotBases)

	for vip, count := range arg.VipCount {
		rsp.VipRobots[vip] = &pb.RobotBases{}
		for i := 0; i < int(count); i++ {
			name, photo, sex := GetUserBase(int(vip), true)
			rsp.VipRobots[vip].Robots = append(rsp.VipRobots[vip].Robots, &pb.RobotBase{
				Id:       fmt.Sprintf("%d", generateOrderId()),
				Nickname: name,
				Photo:    photo,
				Sex:      sex,
				VipLv:    vip,
			})
		}
	}
	ctx.Respond(rsp)
}
