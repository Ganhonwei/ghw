package crash

import (
	"fmt"
	"time"

	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"goserver/pkg/table"
	"goserver/pkg/utils"

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
	tick := time.Tick(time.Millisecond * 50)
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
	tick := time.Tick(time.Millisecond * 200)
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
		// 爆炸检测
		a.boomCheck()
	}

	// 人机撤离
	a.robotTimer++
	if a.robotTimer == 10 {
		a.robotTimer = 0
		robot := table.GetTables().CrashRobotTable.Get()
		leavePro1 := robot.RobotLeave1[1]
		leavePro2 := robot.RobotLeave2[1]
		for k := range a.NewbiewRobot {
			if _, ok := a.CrashBets[k]; ok {
				if _, ok := a.CRASHBack[k]; !ok {
					if utils.RandWan(leavePro1) {
						req := &pb.CRASHBackReq{Userid: k, Pos: 0}
						a.selfPid.Tell(req)
					}
				}
			}

			if _, ok := a.CrashBets1[k]; ok {
				if _, ok := a.CRASHBack1[k]; !ok {
					if utils.RandWan(leavePro2) {
						req := &pb.CRASHBackReq{Userid: k, Pos: 1}
						a.selfPid.Tell(req)
					}
				}
			}
		}
	}
}

// 钟声
func (a *Desk) newbiewDing(ctx actor.Context) {
	// 人机下注
	a.robotBet()
	// if len(a.BetQeuee) <= 0 {
	// 	return
	// }
	// // if len(a.BetQeuee) <= 0 || a.BetLastTime == 0 {
	// // 	a.BetLastTime = utils.BsonNow().UnixMilli()
	// // }

	// now := utils.BsonNow().UnixMilli()
	// sort.Slice(a.BetQeuee, func(i, j int) bool {
	// 	return a.BetQeuee[i].BetTime < a.BetQeuee[j].BetTime
	// })

	// removeIndex := -1
	// for i, q := range a.BetQeuee {
	// 	if now >= q.BetTime {
	// 		a.selfPid.Tell(&pb.CRASHBetReq{Userid: q.Robotid, Value: uint32(q.Chip)})
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
			a.takeoff()      // 开牌
			a.freeGameOver() // 结算
		case int32(pb.STATE_LEAD):
			a.freeGameOver() // 开完牌了，直接结算
		}
	}

	// 返还下一轮下注
	for _, pos := range []int32{0, 1} {
		crashNextBets := a.PosCrashNextBetsMap(pos)
		for userid, bet := range crashNextBets {
			role, ok := a.roles[userid]
			if !ok || role == nil {
				continue
			}
			glog.Infof("停机返还下注: userid=%v, rid=%v, pos=%v, bet=%#v", userid, a.DeskData.Rid, pos, bet)
			a.sendCurrency(userid, bet.Coin, bet.Diamond, int32(pb.LOG_TYPE146), fmt.Sprintf("CRASH房间%s位置%d返还下注", a.DeskData.Rid, pos+1))
			// 可提现金加回去
			if bet.Out > 0 {
				a.checkGiveDiamond(userid, data.Currency{Diamond: bet.Out})
			}
		}
	}
	clear(a.CrashNextBets)
	clear(a.CrashNextBets1)

	//改为结算状态
	a.state = int32(pb.STATE_OVER)
	//玩家退出
	for k := range a.roles {
		errcode := a.leave(k, 0)
		glog.Debugf("stop userid %s, err %v", k, errcode)
		if errcode != pb.OK {
			//continue
		}
		ntf := &pb.CRASHLeaveNtf{Userid: k}
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
