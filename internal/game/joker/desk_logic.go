package joker

import (
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/glog"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// internal function

// NewDesk 新建一张牌桌
func NewDesk(deskData *data.DeskData) *Desk {
	desk := &Desk{
		roles:  make(map[string]*data.DeskRole),
		seats:  make(map[uint32]*data.DeskSeat),
		router: make(map[string]string),
		stopCh: make(chan struct{}),
	}
	desk.DeskData = deskData
	return desk
}

// InitDesk 初始化
func (t *Desk) InitDesk() {
	t.DeskGame = new(data.DeskGame)
	switch t.DeskData.Rtype {
	case int32(pb.ROOM_TYPE0): //自由
		t.state = int32(pb.STATE_FREE)
		t.robotNum = t.randRobotNum()
	case int32(pb.ROOM_TYPE1): //私人
		t.DeskPriv = new(data.DeskPriv)
		t.DeskPriv.PrivScore = make(map[string]int64)
		t.DeskPriv.Joins = make(map[string]uint32)

	case int32(pb.ROOM_TYPE2): //百人
		t.DeskFree = new(data.DeskFree)
		t.DeskFree.Dealers = make(map[string]int64) //上庄列表,userid: carry
		t.DeskFree.Carry = SysCarry
		t.freeInit()
	}
}

// 房间消息广播
func (t *Desk) broadcast(msg interface{}) {
	for _, v := range t.roles {
		if v == nil {
			continue
		}
		if v.Pid == nil {
			continue
		}
		if v.Offline {
			continue
		}
		v.Pid.Tell(msg)
	}
}

// 房间消息广播(除userid外)
func (t *Desk) broadcast2(userid string, msg interface{}) {
	for k, v := range t.roles {
		if v == nil {
			continue
		}
		if v.Pid == nil {
			continue
		}
		if v.Offline {
			continue
		}
		if k != userid {
			v.Pid.Tell(msg)
		}
	}
}

// 房间消息广播(除seat外)
func (t *Desk) broadcast3(seat uint32, msg interface{}) {
	if v, ok := t.seats[seat]; ok && v != nil {
		t.broadcast2(v.Userid, msg)
	}
}

// 广播给指定座位列表
func (t *Desk) broadcast4(seats []uint32, msg interface{}) {
	for _, seatid := range seats {
		userid := t.getUserid(seatid)
		role := t.getRole(userid)
		if role == nil {
			continue
		}
		if role.Pid == nil {
			continue
		}
		if role.Offline {
			continue
		}
		role.Pid.Tell(msg)
	}
}

// 广播给非人机
func (t *Desk) broadcast5(msg interface{}) {
	for _, v := range t.roles {
		if v == nil {
			continue
		}
		if v.Pid == nil {
			continue
		}
		if v.Offline {
			continue
		}
		if v.Robot {
			continue
		}
		v.Pid.Tell(msg)
	}
}

// 获取其他座位列表
func (t *Desk) GetOtherSeats(seats []uint32) (others []uint32) {
	for k, _ := range t.seats {
		found := false
		for _, v := range seats {
			if k == v {
				found = true
			}
		}
		if !found {
			others = append(others, k)
		}
	}
	return
}

// 给玩家发送消息
func (t *Desk) send2userid(userid string, msg interface{}) {
	if v, ok := t.roles[userid]; ok && v != nil {
		v.Pid.Tell(msg)
	}
}

// 给位置发送消息
func (t *Desk) send2seat(seat uint32, msg interface{}) {
	if v, ok := t.seats[seat]; ok && v != nil {
		t.send2userid(v.Userid, msg)
	}
}

// 给玩家发送消息,离线也发送
func (t *Desk) send3userid(userid string, msg interface{}) {
	if v, ok := t.roles[userid]; ok && v != nil {
		if v.Offline {
			return
		}
		v.Pid.Tell(msg)
	}
}

// 给位置发送消息,离线也发送
func (t *Desk) send3seat(seat uint32, msg interface{}) {
	if v, ok := t.seats[seat]; ok && v != nil {
		t.send3userid(v.Userid, msg)
	}
}

// 获取路由
func (t *Desk) getRouter(ctx actor.Context) string {
	glog.Debugf("getRouter %s", ctx.Sender().String())
	return t.router[ctx.Sender().String()]
}

// 获取进程pid
func (t *Desk) getPid(userid string) *actor.PID {
	if v, ok := t.roles[userid]; ok && v != nil {
		return v.Pid
	}
	return nil
}

// 获取玩家数据
func (t *Desk) getPlayer(userid string) *data.User {
	if v, ok := t.roles[userid]; ok && v != nil {
		return v.User
	}
	return nil
}

// 获取role
func (t *Desk) getRole(userid string) *data.DeskRole {
	if v, ok := t.roles[userid]; ok && v != nil {
		return v
	}
	return nil
}

// 是否是人机
func (t *Desk) isRobot(seat uint32) bool {
	u := t.getUserBySeat(seat)
	if u == nil {
		return false
	}
	if u.Robot {
		return true
	}
	return false
}

// 是否还在牌桌上
func (t *Desk) isAlive(seat uint32) bool {
	s := t.getStatus(seat)
	if s == nil {
		return false
	}
	if s.Alive {
		return true
	}
	return false
}

// 获取deskseat
func (t *Desk) getSeat(seat uint32) *data.DeskSeat {
	if v, ok := t.seats[seat]; ok && v != nil {
		return v
	}
	return nil
}

// 获取actstatus
func (t *Desk) getStatus(seat uint32) *data.ActStatus {
	if v, ok := t.ActSeats[seat]; ok && v != nil {
		return v
	}
	return nil
}

// 获取玩家数据
func (t *Desk) getUserBySeat(seat uint32) *data.User {
	if v, ok := t.seats[seat]; ok && v != nil {
		return t.getPlayer(v.Userid)
	}
	return nil
}

// 获取位置
func (t *Desk) getSeatid(userid string) uint32 {
	if v, ok := t.roles[userid]; ok && v != nil {
		return v.Seat
	}
	return 0
}

// 获取位置
func (t *Desk) getUserid(seat uint32) string {
	if v, ok := t.seats[seat]; ok && v != nil {
		return v.Userid
	}
	return ""
}

// 玩家是否在线
func (t *Desk) isOnline(userid string) bool {
	if v, ok := t.roles[userid]; ok && v != nil {
		return !v.Offline
	}
	return false
}

// 设置玩家是否离线
func (t *Desk) setOffline(userid string, offline bool) {
	if v, ok := t.roles[userid]; ok && v != nil {
		v.Offline = offline
	}
}

// 获取手牌
func (t *Desk) getHandCards(seat uint32) []uint32 {
	//房间类型 百人场
	if t.DeskData.Rtype == int32(pb.ROOM_TYPE2) &&
		t.DeskFree != nil {
		return t.DeskFree.Cards[seat]
	}
	//房间类型 非百人场
	if v, ok := t.seats[seat]; ok && v != nil {
		return v.Cards
	}
	glog.Errorf("getHandCards %d", seat)
	t.printOver()
	return []uint32{}
	//panic(fmt.Sprintf("getHandCards error:%d", seat))
}

// 玩家牌力
func (t *Desk) getPower(seat uint32) uint32 {
	if v, ok := t.seats[seat]; ok && v != nil {
		return v.Power
	}
	return 0
}

// 位置下注
func (t *Desk) getBets(seat uint32) int64 {
	if v, ok := t.seats[seat]; ok && v != nil {
		return v.Bet
	}
	return 0
}

// 是否翻牌
func (t *Desk) isSee(seat uint32) bool {
	if t.DeskAct == nil {
		return false
	}
	if v, ok := t.DeskAct.ActSeats[seat]; ok {
		return v.See
	}
	return false
}

func (t *Desk) isChaal(seat uint32) bool {
	if t.DeskAct == nil {
		return false
	}
	if v, ok := t.DeskAct.ActSeats[seat]; ok {
		return v.Chaal
	}
	return false
}

// 是否可以操作
func (t *Desk) CanOper(seat uint32) bool {
	status := t.getStatus(seat)
	if status == nil {
		return false
	}
	if !status.Alive {
		return false
	}
	return true
}

// 事件发送
func (t *Desk) eventPost(userid string, eventId int32, bean any) {
	body, err := json.Marshal(bean)
	if err != nil {
		glog.Errorf("lhd event Marshal fail")
		return
	}
	msg := new(pb.EventPost)
	msg.EventId = eventId
	msg.Data = body
	t.send2userid(userid, msg)
}

// vim: set foldmethod=marker foldmarker=//',//.:
