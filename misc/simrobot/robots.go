package main

import (
	"encoding/json"
	"time"

	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/glog"
)

//' 接口

// Msg2Robots 消息通知
func Msg2Robots(msg interface{}, num uint32) {
	for num > 0 {
		rbs.Send2rbs(msg)
		num--
	}
}

// RegistRoom 房间列表
// func RegistRoom(roomid string, ltype int32) {
// 	msg := &pb.RobotRoomList{
// 		Roomid: roomid,
// 		Ltype:  ltype,
// 	}
// 	glog.Debugf("regist room %s ltype %d", roomid, ltype)
// 	Msg2Robots(msg, 1)
// }

// EnterRoom 进入房间
// func EnterRoom(phone, roomid string, ltype int32) {
// 	msg := &pb.RobotEnterRoom{
// 		Phone:  phone,
// 		Roomid: roomid,
// 		Ltype:  ltype,
// 	}
// 	Msg2Robots(msg, 1)
// }

// Logined 登录成功
// func Logined(phone string, ltype int32) {
// 	msg := &pb.RobotLogin{
// 		Phone: phone,
// 		Ltype: ltype,
// 	}
// 	Msg2Robots(msg, 1)
// }

// Logout 登出成功
// func Logout(roomid, phone, code string, chip int64) {
// 	msg := &pb.RobotLogout{
// 		Roomid: roomid,
// 		Phone:  phone,
// 		Code:   code,
// 		Chip:   chip,
// 	}
// 	Msg2Robots(msg, 1)
// }

// ReLogined 已经注册,重新登录
// func ReLogined(roomid, phone, code string, rtype, envBet int32) {
// 	msg := &pb.RobotReLogin{
// 		Roomid: roomid,
// 		Phone:  phone,
// 		Code:   code,
// 		Rtype:  rtype,
// 		EnvBet: envBet,
// 	}
// 	Msg2Robots(msg, 1)
// }

// 将机器人放入池子
func (r *RobotServer) PutRobot(robot *Robot, gtype int32) {
	if robot.sim {
		r.simrobotPoolMutex.Lock()
		defer r.simrobotPoolMutex.Unlock()

		robot.Reset()
		r.simrobotPool = append(r.simrobotPool, robot)
	} else {
		r.robotPoolMutex.Lock()
		defer r.robotPoolMutex.Unlock()

		robot.Reset()

		if pool, ok := r.robotPool[gtype]; ok {
			pool = append(pool, robot)
			r.robotPool[gtype] = pool
			glog.Infof("PutRobot gtype:%d pool: %d", gtype, len(r.robotPool[gtype]))
		} else {
			pool := []*Robot{robot}
			r.robotPool[gtype] = pool
			glog.Infof("PutRobot gtype:%d pool: %d", gtype, len(r.robotPool[gtype]))
		}

	}
}

// 将机器人拿出池子
func (r *RobotServer) GetRobot(sim bool, gtype int32) *Robot {
	if sim {
		r.simrobotPoolMutex.Lock()
		defer r.simrobotPoolMutex.Unlock()

		if len(r.simrobotPool) != 0 {
			robot := r.simrobotPool[0]
			r.simrobotPool = r.simrobotPool[1:]
			return robot
		} else {
			return nil
		}
	} else {
		r.robotPoolMutex.Lock()
		defer r.robotPoolMutex.Unlock()

		if pool, ok := r.robotPool[gtype]; ok {
			if len(pool) > 0 {
				robot := pool[0]
				r.robotPool[gtype] = pool[1:]
				glog.Infof("GetRobot gtype:%d pool: %d", gtype, len(r.robotPool[gtype]))
				return robot
			} else {
				glog.Infof("GetRobot gtype:%d pool: 0", gtype)
				// 制造机器人
				return nil
			}
		} else {
			glog.Infof("gtype:%d pool: 0", gtype)
			return nil
		}
	}

}

// Send2rbs 发送消息
func (r *RobotServer) Send2rbs(msg interface{}) {
	select {
	case <-r.stopCh:
		return
	default:
		r.msgCh <- msg
	}
}

//.

// ' 机器人测试
func (r *RobotServer) runTest() {
	glog.Infof("runTest started phone -> %s", r.phone1)
	tick := time.Tick(20 * time.Second)
	msg4 := &pb.RobotMsg{
		Num: 1,
	}
	for {
		select {
		case <-tick:
			glog.Infof("r.online -> %d", len(r.online))
			glog.Infof("r.offline -> %d", len(r.offline))
			glog.Infof("r.phone -> %s", r.phone1)
			//TODO:优化,按时间段运行
			//运行指定数量机器人(每个创建一个牌局)
			//code = "create" 表示机器人创建房间
			//go Msg2Robots(msg1, 5)
			if len(r.online) < 3 {
				go Msg2Robots(msg4, 1)
			}
		case <-r.stopCh:
			return
		}
	}
}

// .hua压测
func (r *RobotServer) runHua() {
	tick := time.Tick(1 * time.Second)
	for {
		select {
		case <-tick:
			if !r.loginFinish {
				continue
			}
			for k, v := range r.hua {
				if v > 0 {
					r.hua[k] -= 1
					msg := &pb.RobotMsg{
						Gameid: k,
						Sim:    true,
						Gtype:  int32(pb.HUA),
					}
					go Msg2Robots(msg, 1)
				}
			}
		case <-r.stopCh:
			return
		}
	}
}

// .lhd压测
func (r *RobotServer) runLHD() {
	tick := time.Tick(10 * time.Second)
	for {
		select {
		case <-tick:
			if !r.loginFinish {
				continue
			}
			if r.lhd > 0 {
				r.lhd--
				msg := &pb.RobotMsg{
					Sim:   true,
					Gtype: int32(pb.LHD),
				}
				go Msg2Robots(msg, 1)
			}
		case <-r.stopCh:
			return
		}
	}
}

// .7UP压测
func (r *RobotServer) runUP() {
	tick := time.Tick(10 * time.Second)
	for {
		select {
		case <-tick:
			if !r.loginFinish {
				continue
			}
			if r.updown > 0 {
				r.updown--
				msg := &pb.RobotMsg{
					Sim:   true,
					Gtype: int32(pb.SEVEN),
				}
				go Msg2Robots(msg, 1)
			}
		case <-r.stopCh:
			return
		}
	}
}

// .CRASH压测
func (r *RobotServer) runCRASH() {
	tick := time.Tick(10 * time.Second)
	for {
		select {
		case <-tick:
			if !r.loginFinish {
				continue
			}
			if r.crash > 0 {
				r.crash--
				msg := &pb.RobotMsg{
					Sim:   true,
					Gtype: int32(pb.CRASH),
				}
				go Msg2Robots(msg, 1)
			}
		case <-r.stopCh:
			return
		}
	}
}

func (r *RobotServer) CheckRobotNum(robotnum int) int {
	if robotnum == 0 {
		return 0
	}

	if robotnum/r.count > 0 {
		return r.count
	} else {
		return robotnum
	}
}

func (r *RobotServer) CheckSimRobotNum() int {
	if r.simrobotNum == 0 {
		return 0
	}

	if r.simrobotNum/r.count > 0 {
		return r.count
	} else {
		return r.simrobotNum
	}
}

// 预登录
func (r *RobotServer) runPreLogin() {
	tick := time.Tick(1 * time.Second)
	var robotnum int = 0
	for _, v := range r.robotNum {
		robotnum += v
	}
	for {
		select {
		case <-tick:
			// glog.Info("robot pool: %d", robotnum)
			// glog.Info("simrobot pool: %d", len(r.simrobotPool))

			if count := r.CheckRobotNum(robotnum); count > 0 {
				msg := &pb.RobotPreLogin{
					Sim: false,
				}
				Msg2Robots(msg, uint32(count))
				robotnum -= count
			} else if count := r.CheckSimRobotNum(); count > 0 {
				msg := &pb.RobotPreLogin{
					Sim: true,
				}
				Msg2Robots(msg, uint32(count))
				r.simrobotNum -= count
			} else if !r.loginFinish {
				r.loginFinish = true
			}
		case <-r.stopCh:
			return
		}
	}
}

// ' 消息处理服务
func (r *RobotServer) run() {
	defer func() {
		glog.Infof("Robots closed online -> %d", len(r.online))
		glog.Infof("Robots closed offline -> %d", len(r.offline))
		glog.Infof("Robots closed phone -> %s", r.phone1)
	}()
	glog.Infof("Robots started -> %s", r.phone1)
	tick := time.Tick(time.Minute)
	for {
		select {
		case m, ok := <-r.msgCh:
			if !ok {
				glog.Errorf("Robots msgCh closed phone -> %s", r.phone1)
				return
			}
			switch m.(type) {
			case *pb.RobotPreLogin:
				msg := m.(*pb.RobotPreLogin)
				// glog.Infof("robot pre login -> %#v", msg)
				r.preLogin(msg)
			case *pb.RobotMsg:
				msg := m.(*pb.RobotMsg)
				// glog.Infof("run msg -> %#v", msg)
				r.callRobot(msg)
			// case *pb.RobotReLogin:
			// 	//重新尝试登录
			// 	msg := m.(*pb.RobotReLogin)
			// 	glog.Infof("ReLogin -> %#v", msg)
			// 	// go r.runRobot(msg.Roomid, msg.Phone, msg.Code, msg.Rtype, msg.Ltype, msg.Gtype, msg.EnvBet, false)
			// case *pb.RobotLogin:
			// 	//登录成功
			// 	msg := m.(*pb.RobotLogin)
			// 	glog.Infof("login -> %#v", msg)
			// 	delete(r.offline, msg.Phone)
			// 	r.online[msg.Phone] = true
			// case *pb.RobotLogout:
			// 	//登出断开
			// 	msg := m.(*pb.RobotLogout)
			// 	glog.Infof("logout -> %#v", msg)
			// 	if _, ok := r.online[msg.Phone]; ok {
			// 		delete(r.online, msg.Phone)
			// 	}
			// 	if msg.Chip < 20000 {
			// 		//TODO 自动充值
			// 		//r.unused[msg.Phone] = msg.Chip
			// 	} else {
			// 		//TODO 暂时不重复
			// 		//r.offline[msg.Phone] = true
			// 	}
			// 	if v, ok := r.rooms[msg.Roomid]; ok && v > 0 {
			// 		r.rooms[msg.Roomid]--
			// 	}
			// case *pb.RobotStop:
			// 	msg := m.(*pb.RobotStop)
			// 	glog.Infof("robot stop %#v", msg)
			// r.mutexConns.Lock()
			// for conn := range r.conns {
			// 	conn.Close()
			// }
			// r.conns = nil
			// r.mutexConns.Unlock()
			// case *pb.RobotRoomList:
			// 	msg := m.(*pb.RobotRoomList)
			// 	if v, ok := r.ltypes[msg.Ltype]; ok {
			// 		var have bool
			// 		for _, val := range v {
			// 			if val == msg.Roomid {
			// 				have = true
			// 				break
			// 			}
			// 		}
			// 		if !have {
			// 			v = append(v, msg.Roomid)
			// 			r.ltypes[msg.Ltype] = v
			// 		}
			// 	} else {
			// 		v := make([]string, 0)
			// 		v = append(v, msg.Roomid)
			// 		r.ltypes[msg.Ltype] = v
			// 	}
			// 	glog.Debugf("room list %s ltypes %#v", msg.Roomid, r.ltypes)
			// case *pb.RobotEnterRoom:
			// 	msg := m.(*pb.RobotEnterRoom)
			// 	glog.Debugf("RobotEnterRoom -> %#v", msg)
			// 	r.rooms[msg.Roomid]++
			// 	glog.Debugf("rooms -> %#v", r.rooms)
			case closeFlag:
				//停止发送消息
				close(r.stopCh)
				return
			}
		case <-tick:
			//逻辑处理
		}
	}
}

// 机器人预登录
func (r *RobotServer) preLogin(msg *pb.RobotPreLogin) {
	sim := msg.Sim
	if sim {
		r.simrobotId++
		if r.simrobotId > r.maxId {
			return
		}
		go r.runRobot(r.simrobotId, sim)
	} else {
		r.robotId++
		if r.robotId > r.maxId {
			return
		}
		go r.runRobot(r.robotId, sim)
	}
}

// 召唤机器人
func (r *RobotServer) callRobot(msg *pb.RobotMsg) {
	if !r.loginFinish {
		return
	}
	robot := r.GetRobot(msg.Sim, msg.Gtype)
	if robot == nil {
		return
	}
	robot.gtype = msg.Gtype //游戏类型
	robot.gameid = msg.Gameid
	robot.roomid = msg.Roomid
	robot.min = msg.Min
	robot.max = msg.Max

	var emoji_config data.EmojiConfig
	err := json.Unmarshal(msg.Emoji, &emoji_config)
	if err == nil {
		robot.emoji = emoji_config
	}

	go robot.adjustCoin()
}

// ' 启动机器人
// func (r *RobotServer) run2(msg *pb.RobotMsg) {
// 	var code string = msg.Code
// 	var rtype int32 = msg.Rtype
// 	var ltype int32 = msg.Ltype
// 	var envBet int32 = msg.EnvBet
// 	var gtype int32 = msg.Gtype
// 	var min int32 = msg.Min
// 	var max int32 = msg.Max
// 	var gameid string = msg.Gameid
// 	var roomid string = msg.Roomid
// 	var sim bool = msg.Sim
// 	if sim {
// 		r.phone2 = utils.StringAdd(r.phone2)
// 		phone := "simrobot" + r.phone2
// 		go r.runRobot(gameid, roomid, phone, code, rtype, ltype, gtype, envBet, min, max, sim, false)
// 	} else {
// 		r.phone1 = utils.StringAdd(r.phone1)
// 		phone := "robot" + r.phone1
// 		go r.runRobot(gameid, roomid, phone, code, rtype, ltype, gtype, envBet, min, max, sim, false)
// 	}
// 	//选择一个房间
// 	// glog.Debugf("ltypes %#v", r.ltypes)

// 	// if len(msg.Roomid) != 0 {
// 	// 	roomid = msg.Roomid
// 	// } else {
// 	// 	if s, ok := r.ltypes[ltype]; ok {
// 	// 		for _, v := range s {
// 	// 			//每个房间5个人
// 	// 			if r.rooms[v] < 8 {
// 	// 				roomid = v
// 	// 				break
// 	// 			}
// 	// 		}
// 	// 	}
// 	// }
// 	//房间已经存在列表
// 	// if roomid == "" && len(r.ltypes[ltype]) != 0 {
// 	// 	return
// 	// }
// 	// glog.Debugf("roomid %s, ltype %d", roomid, ltype)
// 	// for k, v := range r.offline {
// 	// 	if v { //已经断开
// 	// 		phone = k
// 	// 		r.offline[k] = false //登录中
// 	// 		glog.Infof("run offline robot -> %s, %s", roomid, phone)
// 	// 		// go r.runRobot(gameid, roomid, phone, code, rtype, ltype, gtype, envBet, false)
// 	// 		break
// 	// 	}
// 	// }
// 	// glog.Infof("offline phone -> %s", phone)
// 	// if len(phone) == 0 {
// 	// 	phone = r.phone
// 	// 	r.phone = utils.StringAdd(r.phone)
// 	// 	if _, ok := r.unused[phone]; ok {
// 	// 		//TODO 自动充值
// 	// 		//Msg2Robots(msg, 1)
// 	// 		//TODO 会出现死循环
// 	// 	} else if _, ok := r.online[phone]; ok {
// 	// 		//重复, TODO 会出现死循环
// 	// 		//Msg2Robots(msg, 1)
// 	// 	} else {
// 	// 		//新机器人不用注册
// 	// 		glog.Infof("run new robot -> %s, %s", roomid, phone)
// 	// 		phone = "robot" + phone
// 	// 		go r.runRobot(gameid, roomid, phone, code, rtype, ltype, gtype, envBet, min, max, false)
// 	// 	}
// 	// }
// 	// glog.Infof("new phone -> %s", phone)
// 	//重置
// 	// phone1 := cfg.Section("robot").Key("phone").Value()
// 	// //if r.phone > utils.StringAdd2(phone1, "750") {
// 	// if r.phone > utils.StringAdd2(phone1, "286") {
// 	// 	r.phone = phone1
// 	// }
// }

//.

// vim: set foldmethod=marker foldmarker=//',//.:
