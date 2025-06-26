package redblack

import (
	"time"

	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func (a *DeskActor) handlerMsg(msg interface{}, ctx actor.Context) {
	switch msg.(type) {
	case *pb.Connected:
		//连接成功
		arg := msg.(*pb.Connected)
		glog.Infof("Connected %s", arg.Name)
	case *pb.Disconnected:
		//成功断开
		arg := msg.(*pb.Disconnected)
		glog.Infof("Disconnected %s", arg.Name)
	case *pb.CloseDesk:
		arg := msg.(*pb.CloseDesk)
		glog.Debugf("CloseDesk %#v", arg)
		//移除
		delete(a.desks, arg.Roomid)
		delete(a.rules, arg.Unique)
		//TODO 优化,重复了
		//a.roomPid.Request(msg, ctx.Self())
		//响应
		//rsp := new(pb.ClosedDesk)
		//ctx.Respond(rsp)
	case *pb.LeaveDesk:
		arg := msg.(*pb.LeaveDesk)
		glog.Debugf("LeaveDesk %#v", arg)
		if v, ok := a.desks[arg.Roomid]; ok &&
			v.Number > 0 {
			v.Number--
		}
		//响应
		//rsp := new(pb.LeftDesk)
		//ctx.Respond(rsp)
	case *pb.JoinDesk:
		arg := msg.(*pb.JoinDesk)
		glog.Debugf("JoinDesk %#v", arg)
		//房间数据变更
		if v, ok := a.desks[arg.Roomid]; ok {
			v.Number++
		}
		//响应
		//rsp := new(pb.EnteredRoom)
		//ctx.Respond(rsp)
	case *pb.EnterDesk:
		arg := msg.(*pb.EnterDesk)
		glog.Debugf("EnterDesk %#v", arg.Sender)
		a.enterDesk(arg, ctx)
	//case *pb.CreateDesk:
	//	arg := msg.(*pb.CreateDesk)
	//	glog.Debugf("CreateDesk %#v", arg)
	//	a.createDesk(arg, ctx)
	case *pb.SyncConfig:
		//同步配置
		arg := msg.(*pb.SyncConfig)
		// glog.Debugf("SyncConfig %#v", arg)
		a.syncDesk(arg, ctx)
	case *pb.GetRoomList:
		arg := msg.(*pb.GetRoomList)
		glog.Debugf("GetRoomList %#v", arg)
		if arg.Sender == nil {
			return
		}
	case *pb.CloseServer:
		a.CloseServer(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}

func (a *DeskActor) CloseServer(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.CloseServer)
	glog.Debugf("CloseServer %#v", arg)
	a.closeServer = true
	a.broadcast(arg)
}

// '后台添加房间
func (a *DeskActor) syncDesk(arg *pb.SyncConfig, ctx actor.Context) {
	// switch arg.Type {
	// case pb.CONFIG_GAMES:
	// 	b := make(map[string]data.Game)
	// 	err = json.Unmarshal(arg.Data, &b)
	// 	if err != nil {
	// 		glog.Errorf("syncConfig Unmarshal err %v", err)
	// 		return
	// 	}
	// 	for _, v := range b {
	// 		//不是当前节点房间
	// 		if v.Node != nodeName {
	// 			continue
	// 		}
	// 		switch arg.Atype {
	// 		case pb.CONFIG_DELETE:
	// 			//关闭房间
	// 			v2 := v
	// 			a.closeDesk(&v2, ctx)
	// 		case pb.CONFIG_UPSERT:
	// 			//创建房间
	// 			v2 := v
	// 			//已经存在
	// 			if id, ok := a.rules[v2.Id]; ok {
	// 				if p, ok2 := a.desks[id]; ok2 {
	// 					p.Pid.Tell(arg)
	// 					continue
	// 				}
	// 			}
	// 			a.spawnDesk(&v2, ctx)
	// 		}
	// 	}
	// default:
	// }
	handler.SyncConfig(arg, a.Name)
}

// 关闭一张桌子
func (a *DeskActor) closeDesk(uid string, ctx actor.Context) {
	glog.Debugf("close Desk %#v", uid)
	//可以去room服务中取
	if k, ok := a.rules[uid]; ok {
		glog.Debugf("close Desk %s", k)
		a.stopDesk(k, ctx)
		delete(a.rules, uid)
	}
}

// 停止服务
func (a *DeskActor) stopDesk(roomid string, ctx actor.Context) {
	if v, ok := a.desks[roomid]; ok {
		//关闭房间消息
		msg1 := new(pb.ServeStop)
		v.Pid.Request(msg1, ctx.Self())
		//关闭房间消息
		//msg2 := new(pb.CloseDesk)
		//msg2.Roomid = roomid
		//a.roomPid.Request(msg2, ctx.Self())
		//delete(a.desks, roomid)
	}
}

//.

// '进入百人或者匹配房间
func (a *DeskActor) enterDesk(arg *pb.EnterDesk, ctx actor.Context) {
	rsp := new(pb.EnteredDesk)
	rsp.Gtype = arg.Gtype
	rsp.Gameid = arg.Gameid
	rsp.Rtype = arg.Rtype

	if a.closeServer {
		rsp.Error = pb.ServerClose
		ctx.Respond(rsp)
		return
	}

	//指定房间(人机)
	if d, ok := a.desks[arg.Roomid]; ok {
		d.Pid.Tell(arg)
		return
	}
	if arg.Roomid != "" {
		rsp.Error = pb.Failed
		ctx.Respond(rsp)
		return
	}
	//查找房间
	for _, v := range a.desks {
		if arg.Dtype == int32(pb.DESK_TYPE_POINTCONTROL) {
			if v.Dtype != arg.Dtype ||
				v.DeskData.Rtype != arg.Rtype {
				continue
			}
			// 点控桌
			rsp, err := v.Pid.RequestFuture(arg, 2*time.Second).Result()
			if err != nil {
				continue
			}
			if r, ok := rsp.(*pb.EnteredDesk); ok {
				if r.Error == pb.OK {
					arg.Sender.Tell(r)
					// 新建一张桌子
					gameData := handler.NewFreeGameData(a.Name, "90023", int32(pb.REDBLACK), int32(pb.DESK_TYPE_POINTCONTROL))
					if _, ok := a.spawnDesk(gameData, ctx); ok {
						return
					}
				}
			}
			continue
		}

		// if arg.Dtype == int32(pb.DESK_TYPE_NEWBIEW) || arg.Dtype == int32(pb.DESK_TYPE_NORMAL_B) {
		// 	if v.Dtype != arg.Dtype ||
		// 		v.DeskData.Rtype != arg.Rtype {
		// 		continue
		// 	}
		// 	// 新手模式,只能进一个人
		// 	if v.Number >= 1 {
		// 		continue
		// 	}
		// 	v.Pid.Tell(arg)
		// 	return
		// }

		if v.DeskData.Rtype == arg.Rtype &&
			v.DeskData.Ltype == arg.Ltype &&
			v.DeskData.Dtype == arg.Dtype &&
			int32(pb.ROOM_TYPE1) != arg.Rtype &&
			v.Number < 100 {
			v.Pid.Tell(arg)
			return
		}
	}
	//创建一个新的房间
	switch arg.Rtype {
	case int32(pb.ROOM_TYPE2): //百人
		//TODO 优化为后台添加配置
		gameData := handler.NewFreeGameData(a.Name, "90011", int32(pb.REDBLACK), int32(pb.DESK_TYPE_NORMAL))
		switch arg.Dtype {
		case int32(pb.DESK_TYPE_NEWBIEW):
			gameData = handler.NewFreeGameData(a.Name, "90011", int32(pb.REDBLACK), int32(pb.DESK_TYPE_NEWBIEW))
		case int32(pb.DESK_TYPE_POINTCONTROL):
			gameData = handler.NewFreeGameData(a.Name, "90023", int32(pb.REDBLACK), int32(pb.DESK_TYPE_POINTCONTROL))
		case int32(pb.DESK_TYPE_NORMAL_B):
			gameData = handler.NewFreeGameData(a.Name, "90011", int32(pb.REDBLACK), int32(pb.DESK_TYPE_NORMAL_B))
		}
		if deskPid, ok := a.spawnDesk(gameData, ctx); ok {
			deskPid.Tell(arg)
			return
		}
	case int32(pb.ROOM_TYPE1): //私人
		glog.Errorf("enter Desk err %#v", arg)
	case int32(pb.ROOM_TYPE0): //自由
		//TODO 查找匹配房间,
		//测试时动态添加,正式时后台配置
		gameData := handler.NewCoinGameData(a.Name,
			int32(pb.REDBLACK), arg.Dtype, arg.Ltype)
		if deskPid, ok := a.spawnDesk(gameData, ctx); ok {
			deskPid.Tell(arg)
			return
		}
	default:
	}
	rsp.Error = pb.Failed
	ctx.Respond(rsp)
	//arg.Sender.Tell(rsp)
}

//.

// '匹配房间
func (a *DeskActor) changeDesk(arg *pb.ChangeDesk, ctx actor.Context) {
	rsp := new(pb.ChangedDesk)
	rsp.Gtype = arg.Gtype
	//查找房间
	for _, v := range a.desks {
		if v.DeskData.Rtype == arg.Rtype &&
			v.DeskData.Ltype == arg.Ltype &&
			int32(pb.ROOM_TYPE1) != arg.Rtype &&
			v.DeskData.Rid != arg.Roomid &&
			v.Number < v.DeskData.Count {
			rsp.Desk = v.Pid
			arg.Sender.Tell(rsp)
			return
		}
	}
	//创建一个新的房间
	switch arg.Rtype {
	case int32(pb.ROOM_TYPE2): //百人
		glog.Errorf("change Desk err %#v", arg)
	case int32(pb.ROOM_TYPE1): //私人
		glog.Errorf("change Desk err %#v", arg)
	case int32(pb.ROOM_TYPE0): //自由
		//TODO 查找匹配房间,
		//测试时动态添加,正式时后台配置
		gameData := handler.NewCoinGameData(a.Name,
			int32(pb.REDBLACK), arg.Dtype, arg.Ltype)
		if deskPid, ok := a.spawnDesk(gameData, ctx); ok {
			rsp.Desk = deskPid
			arg.Sender.Tell(rsp)
			return
		}
	default:
	}
	rsp.Error = pb.Failed
	arg.Sender.Tell(rsp)
}

//.

// '启动新服务,新开的房间同步状态
func (a *DeskActor) spawnDesk(gameData *data.Game,
	ctx actor.Context) (deskPid *actor.PID, ok bool) {
	deskData := handler.NewDeskData(gameData)
	return a.spawnDesk2(deskData, ctx)
}

func (a *DeskActor) spawnDesk2(deskData *data.DeskData,
	ctx actor.Context) (deskPid *actor.PID, ok bool) {
	glog.Debugf("spawn Desk %#v", deskData)
	//新桌子
	newDesk1 := NewDesk(deskData)
	//spawn desk
	deskPid = newDesk1.newDesk()
	glog.Debugf("deskPid: %#v", deskPid.String())
	//添加桌子
	if !a.addDesk(deskData, deskPid, ctx) {
		//关闭房间消息
		msg1 := new(pb.ServeStop)
		deskPid.Request(msg1, ctx.Self())
		ok = false
		return
	}
	newDesk1.dbmsPid = a.dbmsPid
	newDesk1.roomPid = a.roomPid
	newDesk1.rolePid = a.rolePid
	// newDesk1.loggerPid = a.loggerPid
	newDesk1.selfPid = deskPid
	//添加新桌子
	a.desks[deskData.Rid] = &data.DeskBase{
		DeskData: deskData,
		Pid:      deskPid,
	}
	//规则暂时只关闭时用到
	a.rules[deskData.Unique] = deskData.Rid
	//启动
	deskPid.Tell(new(pb.ServeStart))
	glog.Debugf("spawn Desk successfully %s, %s",
		deskData.Rid, deskPid.String())
	glog.Debugf("spawn Desk %#v", deskData)
	ok = true
	return
}

// 添加桌子
func (a *DeskActor) addDesk(deskData *data.DeskData,
	deskPid *actor.PID, ctx actor.Context) bool {
	//添加桌子
	msg2 := new(pb.AddDesk)
	msg2.Desk = deskPid
	msg2.Roomid = deskData.Rid
	msg2.Rtype = deskData.Rtype
	msg2.Gtype = deskData.Gtype
	msg2.Unique = deskData.Unique
	res2 := a.reqRoom(msg2, ctx)
	var response2 *pb.AddedDesk
	var ok bool
	if response2, ok = res2.(*pb.AddedDesk); !ok {
		glog.Errorf("add desk failed: %#v", res2)
		return false
	}
	if response2.Error != pb.OK {
		glog.Errorf("add desk failed: %v", response2.Error)
		return false
	}
	glog.Debugf("add Desk successfully %s, %s",
		response2.Roomid, response2.Code)
	deskData.Rid = response2.Roomid
	deskData.Code = response2.Code
	return true
}

//获取桌子唯一id
////deskData.Rid = a.genDesk(deskData.Rtype, deskData.Gtype, ctx)
//func (a *DeskActor) genDesk(rtype, gtype int32, ctx actor.Context) string {
//	msg1 := new(pb.GenDesk)
//	msg1.Rtype = rtype
//	msg1.Gtype = gtype
//	res1 := a.reqRoom(msg1, ctx)
//	if response1, ok := res1.(*pb.GenedDesk); ok {
//		glog.Debugf("response1: %#v", response1)
//		return response1.Roomid
//	}
//	glog.Error("spawn desk failed: rtype %d, gtype %d",
//		rtype, gtype)
//	return ""
//}

//.

// '创建房间
func (a *DeskActor) createDesk(arg *pb.CreateDesk, ctx actor.Context) {
	rsp := new(pb.CreatedDesk)
	rsp.Rtype = arg.Rtype
	rsp.Gtype = arg.Gtype
	//创建一个新的房间
	switch arg.Rtype {
	case int32(pb.ROOM_TYPE2): //百人
	case int32(pb.ROOM_TYPE1): //私人
		gameData := handler.NewPrivGameData(arg)
		if deskPid, ok := a.spawnDesk2(gameData, ctx); ok {
			//创建成功消息
			// msg := handler.PackNNCreateMsg(gameData)
			// arg.Sender.Tell(msg)
			//扣除玩家消耗
			msg2 := handler.ChangeCurrencyMsg((-1 * int64(gameData.Cost)),
				0, 0, 0, 0, 0, int32(pb.LOG_TYPE2), arg.Cid, "redblack创建私人房间", "")
			arg.Sender.Tell(msg2)
			//响应消息
			rsp.Desk = deskPid
			arg.Sender.Tell(rsp)
			return
		}
	case int32(pb.ROOM_TYPE0): //自由
	}
	rsp.Error = pb.Failed
	arg.Sender.Tell(rsp)
}

//.

// 登录成功数据处理
func (a *DeskActor) reqRoom(msg interface{}, ctx actor.Context) interface{} {
	timeout := 3 * time.Second
	res1, err1 := a.roomPid.RequestFuture(msg, timeout).Result()
	if err1 != nil {
		glog.Errorf("reqRoom err: %v, msg %#v", err1, msg)
		return nil
	}
	return res1
}

// 广播消息
func (a *DeskActor) broadcast(msg interface{}) {
	for _, v := range a.desks {
		v.Pid.Tell(msg)
	}
}

// 关闭多余房间
func (a *DeskActor) stopIdleRoom(ctx actor.Context) {
	for k := range pb.DeskType_name {
		idles := make([]*data.DeskBase, 0)
		for _, d := range a.desks {
			if d.Dtype != k {
				continue
			}
			// 获取空闲状态的桌子
			if d.Number <= 0 {
				idles = append(idles, d)
			}
		}
		size := len(idles)
		if size <= 5 {
			return
		}
		for i := 5; i < size; i++ {
			// 关闭多余的
			idle := idles[i]
			a.closeDesk(idle.Unique, ctx)
		}
	}
}

func (a *DeskActor) createPCRoom(ctx actor.Context) {
	var num int = 0
	for _, d := range a.desks {
		if d.Dtype == int32(pb.DESK_TYPE_POINTCONTROL) {
			num++
		}
	}
	if num <= 0 {
		//创建一个点控桌子
		gameData := handler.NewFreeGameData(a.Name, "90023", int32(pb.REDBLACK), int32(pb.DESK_TYPE_POINTCONTROL))
		if deskPid, ok := a.spawnDesk(gameData, ctx); ok {
			glog.Infof("deskPid: %s", deskPid.String())
			// return
		}
	}
}

func (a *DeskActor) stopNewbiewRoom(ctx actor.Context) {
	idles := make([]*data.DeskBase, 0)
	for _, d := range a.desks {
		if (d.Dtype == int32(pb.DESK_TYPE_NEWBIEW) || d.Dtype == int32(pb.DESK_TYPE_NORMAL_B)) && d.Number <= 0 {
			idles = append(idles, d)
		}
	}
	size := len(idles)
	if size <= 1 {
		return
	}
	for i := 1; i < size; i++ {
		// 关闭多余的
		idle := idles[i]
		a.closeDesk(idle.Unique, ctx)
	}
}

// vim: set foldmethod=marker foldmarker=//',//.:
