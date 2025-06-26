package up7

import (
	"time"

	"goserver/gen/pb"
	"goserver/pkg/glog"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// 启动服务
func (a *Desk) start(ctx actor.Context) {
	glog.Infof("desk start: %v", ctx.Self().String())
	//初始化
	a.InitDesk()
	//启动
	go a.ticker(ctx)
	go a.newBiewTicker(ctx)
}

// 时钟
func (a *Desk) ticker(ctx actor.Context) {
	tick := time.Tick(time.Second)
	msg := new(pb.Tick)
	for {
		select {
		case <-a.stopCh:
			glog.Info("desk ticker closed")
			return
		default: //防止阻塞
		}
		select {
		case <-a.stopCh:
			glog.Info("desk ticker closed")
			return
		case <-tick:
			ctx.Self().Tell(msg)
		}
	}
}

// 时钟
func (a *Desk) newBiewTicker(ctx actor.Context) {
	tick := time.Tick(time.Millisecond * 100)
	msg := new(pb.Tick2)
	for {
		select {
		case <-a.stopCh:
			glog.Info("desk ticker closed")
			return
		default: //防止阻塞
		}
		select {
		case <-a.stopCh:
			glog.Info("desk ticker closed")
			return
		case <-tick:
			ctx.Self().Tell(msg)
		}
	}
}

// 钟声
func (a *Desk) ding(ctx actor.Context) {
	//glog.Debugf("ding: %v", ctx.Self().String())
	//逻辑处理
	switch a.DeskData.Rtype {
	case int32(pb.ROOM_TYPE0): //自由
	case int32(pb.ROOM_TYPE1): //私人
	case int32(pb.ROOM_TYPE2): //百人
		// 检查桌子状态
		a.checkDeskStatus()
		// a.freeTimeout()
	}
	// if a.Dtype != int32(pb.DESK_TYPE_NEWBIEW) && a.Dtype != int32(pb.DESK_TYPE_NORMAL_B) {
	// 	a.loadRobot()
	// } else {
	// }
}

// 钟声
func (a *Desk) newbiewDing(ctx actor.Context) {
	// 人机下注
	a.robotBet()
	// if a.Dtype != int32(pb.DESK_TYPE_NEWBIEW) || len(a.BetQeuee) <= 0 {
	// 	return
	// }
	// if len(a.BetQeuee) <= 0 {
	// 	return
	// }

	// now := utils.BsonNow().UnixMilli()
	// sort.Slice(a.BetQeuee, func(i, j int) bool {
	// 	return a.BetQeuee[i].BetTime < a.BetQeuee[j].BetTime
	// })

	// removeIndex := -1
	// for i, q := range a.BetQeuee {
	// 	room := config.GetGame(q.Roomid)
	// 	if room.Id == "" {
	// 		continue
	// 	}
	// 	if now >= q.BetTime {
	// 		a.newBiewFreeBet(q.Robotid, q.Seat, q.Chip)
	// 	} else {
	// 		removeIndex = i
	// 		break
	// 	}
	// 	removeIndex = i
	// }
	// if removeIndex >= 0 {
	// 	a.BetQeuee = append(make([]data.NewbiewFreeBetQeuee, 0), a.BetQeuee[removeIndex+1:]...)
	// }
}

// 关闭时钟
func (a *Desk) closeTick() {
	select {
	case <-a.stopCh:
		return
	default:
		//停止发送消息
		close(a.stopCh)
	}
}

func (a *Desk) handlerStop(ctx actor.Context) {
	glog.Infof("handlerStop: %s", a.Name)
	//关闭
	a.closeTick()
	//玩家结算退出
	switch a.DeskData.Rtype {
	case int32(pb.ROOM_TYPE0): //自由
	case int32(pb.ROOM_TYPE1): //私人
	case int32(pb.ROOM_TYPE2): //百人
		switch a.state {
		case int32(pb.STATE_BET):
			a.lead()         // 开牌
			a.freeGameOver() // 结算
		case int32(pb.STATE_LEAD):
			a.freeGameOver() // 开完牌了，直接结算
		}
	}
	// 切换为结算
	a.state = int32(pb.STATE_OVER)
	//玩家退出
	for k := range a.roles {
		errcode := a.leave(k, 0)
		glog.Debugf("stop userid %s, err %v", k, errcode)
		if errcode != pb.OK {
			//continue
		}
		ntf := &pb.UPLeaveNtf{Userid: k}
		a.send2userid(k, ntf)
		//离开状态消息
		a.userLeaveDesk(k)
	}
	for _, u := range a.NewbiewRobot {
		// 归还头像
		a.returnHead(u)
	}
	//关闭房间消息
	msg := new(pb.CloseDesk)
	msg.Roomid = a.DeskData.Rid
	msg.Code = a.DeskData.Code
	msg.Rtype = a.DeskData.Rtype
	msg.Gtype = a.DeskData.Gtype
	msg.Unique = a.DeskData.Unique
	nodePid.Tell(msg)
	//TODO 优化
	if a.roomPid != nil {
		a.roomPid.Request(msg, ctx.Self())
	}
	//停掉服务
	ctx.Self().Stop()
}
