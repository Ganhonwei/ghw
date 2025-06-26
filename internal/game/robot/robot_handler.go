package robot

import (
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/config"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"sort"
	"strings"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func (a *RoleActor) ServeStart(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.ServeStart)
	glog.Debugf("rs ServeStart %s, %#v", ctx.Self().String(), arg)
	go a.ticker()
}

func (a *RoleActor) handlerStop(ctx actor.Context) {
	glog.Debugf("handlerStop %s", a.Userid)
	a.gamePid = nil
	a.gameId = ""
	a.roomId = ""
	nodePid.Tell(&pb.RobotLeave{Userid: a.Userid})
	if strings.HasPrefix(a.Photo, "http") {
		sex := "man"
		if a.Sex == 2 {
			sex = "woman"
		}

		index := strings.LastIndex(a.Photo, "/")
		str := a.Photo[index+1 : len(a.Photo)-4]
		id := utils.Uint64(str)
		// 归还头像
		returnHead(sex, uint32(id))
		glog.Debugf("%s return head %d", a.Userid, id)
	}
}

func (a *RoleActor) ServeClose(ctx actor.Context) {
	// 离开桌子
	a.leaveDesk()
	select {
	case <-a.stopCh:
		return
	default:
		//停止发送消息
		close(a.stopCh)
	}
	ctx.Self().Stop()
}

func (a *RoleActor) ticker() {
	tick := time.Tick(time.Millisecond * 100)
	msg := new(pb.Tick)
	for {
		select {
		case <-a.stopCh:
			glog.Info("robot node ticker closed")
			return
		default:
		}
		select {
		case <-a.stopCh:
			glog.Info("robot node ticker closed")
			return
		case <-tick:
			// glog.Debugf("tick userid:%s", a.Userid)
			a.pid.Tell(msg)
		}
	}
}

func (a *RoleActor) closeRs() {
	a.pid.Tell(new(pb.ServeStop))
}

func (a *RoleActor) leaveDesk() {
	if a.gamePid == nil {
		return
	}
	//下线
	msg3 := new(pb.LeaveDesk)
	if a.User != nil {
		msg3.Userid = a.User.GetUserid()
	}
	//rs.gamePid.Tell(msg3)
	timeout := 3 * time.Second
	res1, err1 := a.gamePid.RequestFuture(msg3, timeout).Result()
	if err1 != nil {
		glog.Errorf("leave desk failed: %v", err1)
	}
	if response1, ok := res1.(*pb.LeftDesk); ok {
		glog.Debugf("left desk response1: %#v", response1)
		if response1.Error != pb.OK {
			glog.Errorf("leave desk failed: %#v", res1)
		} else {
			//不同游戏可能规则不同，由游戏控制是否离开
			a.gamePid = nil
		}
	} else {
		glog.Errorf("leave desk failed: %#v", res1)
	}
}

func (a *RoleActor) ding(ctx actor.Context) {
	switch a.gtype {
	case int32(pb.SEVEN):
		a.upCheckStatus()
	case int32(pb.CRASH):
		a.crashCheckStatus()
	case int32(pb.LHD):
		a.lhdCheckStatus()
	case int32(pb.ABAR):
		a.abCheckStatus()
	case int32(pb.LOTTERY):
		a.cpCheckStatus()
	case int32(pb.PLANE):
		a.planeCheckStatus()
	}
	now := utils.LocalTime().UnixMilli()
	if a.lastSendTime == 0 {
		a.lastSendTime = now
	}
	if len(a.msgQuee) > 0 {
		dq := a.msgQuee[0]
		if a.lastSendTime+dq.sendTime <= now && dq.sendTime > 0 {
			if dq.delayCall == nil || dq.delayCall() { // 延时消息不执行
				a.Sender(dq.msg)
			}
			a.lastSendTime += dq.sendTime
			a.msgQuee = append(make([]DelayQuee, 0), a.msgQuee[1:]...)
		}
	} else {
		// 队列空了就重置时间
		a.lastSendTime = now
	}
	// sort.Slice(a.msgQuee, func(i, j int) bool {
	// 	return a.msgQuee[i].sendTime < a.msgQuee[j].sendTime
	// })
	// 人机表情
	for i, dq := range a.emojiQuee {
		if dq.sendTime <= now && dq.sendTime > 0 {
			a.Sender(dq.msg)
			a.emojiQuee[i].sendTime = 0
		}
	}
	sort.Slice(a.emojiQuee, func(i, j int) bool {
		return a.emojiQuee[i].sendTime < a.emojiQuee[j].sendTime
	})
	removeIndex := -1
	for i, dq := range a.emojiQuee {
		if dq.sendTime > 0 {
			removeIndex = i - 1
		}
	}
	if removeIndex >= 0 {
		a.emojiQuee = append(make([]DelayQuee, 0), a.emojiQuee[removeIndex+1:]...)
	}
	// a.lastSendTime = now
}

func (a *RoleActor) buildUserInfo(arg *pb.RobotMsg) {
	// 随机一个人机vip等级
	vipLv := 0
	vrChoices := config.GetVipRobotChoices()
	if len(vrChoices) > 0 {
		vip, err := utils.WeightedChoice(vrChoices)
		if err != nil {
			glog.Errorf("robot vip weight choice error: %v, %v", err, vrChoices)
		} else {
			vipLv = vip.Item.(int)
		}
	}

	name, photo, sex := GetUserBase(vipLv, true)
	var coin int64 = 0
	if arg.Gtype == int32(pb.LHD) ||
		arg.Gtype == int32(pb.SEVEN) ||
		arg.Gtype == int32(pb.CRASH) ||
		arg.Gtype == int32(pb.ABAR) ||
		arg.Gtype == int32(pb.LOTTERY) ||
		arg.Gtype == int32(pb.PLANE) {
		coin = int64(utils.RandMN(int(arg.Min), int(arg.Max)))
	} else if arg.Gtype == int32(pb.HUA) {
		coin = int64(utils.RandMN(int(arg.Min), int(arg.Max)))
	} else {
		coin = int64(arg.Min * 1000)
	}
	glog.Warning("robot id:", arg.Userid)
	a.User = &data.User{
		Userid:   arg.Userid,
		Nickname: name,
		Photo:    photo,
		Sex:      sex,
		Coin:     coin,
		Robot:    true,
	}
	a.User.Vip.Lv = vipLv

	a.DeskData = new(DeskData)
}

func (r *RoleActor) GetRandSeat() uint32 {
	if len(r.seats) > 0 {
		idx := utils.RandIntN(len(r.seats))
		return r.seats[idx]
	} else {
		return 0
	}
}

func (a *RoleActor) Sender(msg interface{}) {
	if a.gamePid != nil {
		a.gamePid.Request(msg, a.pid)
	}
}

func (a *RoleActor) SendDefer(msg interface{}) {
	s := utils.RandInt64N(2) + 2
	a.SendDefer2(msg, 1000*s)
}

func (a *RoleActor) SendDefer2(msg interface{}, delay int64) {
	// now := utils.LocalTime().UnixMilli()
	d := DelayQuee{msg: msg, sendTime: delay}
	a.msgQuee = append(a.msgQuee, d)
}

func (a *RoleActor) SendDefer3(msg interface{}, delay int64, delayCall func() bool) {
	d := DelayQuee{msg: msg, sendTime: delay, delayCall: delayCall}
	a.msgQuee = append(a.msgQuee, d)
}

func (a *RoleActor) SendEmojiDefer(msg interface{}, delay int64) {
	// now := utils.LocalTime().UnixMilli()
	d := DelayQuee{msg: msg, sendTime: delay}
	a.emojiQuee = append(a.emojiQuee, d)
}
