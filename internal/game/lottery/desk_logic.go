package lottery

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"strconv"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// 关闭桌子
func (t *Desk) CloseDesk(ctx actor.Context) {
	arg := ctx.Message().(*pb.CloseDesk)
	glog.Debugf("CloseDesk %#v", arg)
	t.closeDesk(arg, ctx)
}

// 离开桌子
func (t *Desk) LeaveDesk(ctx actor.Context) {
	arg := ctx.Message().(*pb.LeaveDesk)
	glog.Debugf("LeaveDesk %#v", arg)
	t.leaveDesk(arg, ctx)
}

// 同步配置
func (t *Desk) SyncConfig(ctx actor.Context) {
	arg := ctx.Message().(*pb.SyncConfig)
	// glog.Debugf("SyncConfig %#v", arg)
	t.syncConfig(arg, ctx)
}

// 进入桌子
func (t *Desk) EnterDesk(ctx actor.Context) {
	arg := ctx.Message().(*pb.EnterDesk)
	glog.Debugf("EnterDesk %#v", arg.Sender)
	t.enterDesk(arg, ctx)
}

// 离线消息
func (t *Desk) OfflineDesk(ctx actor.Context) {
	arg := ctx.Message().(*pb.OfflineDesk)
	glog.Debugf("OfflineDesk %#v", arg)
	//离线消息
	t.offlineDesk(arg.Userid)
}

// 货币变更
func (t *Desk) ChangeCurrency(ctx actor.Context) {
	arg := ctx.Message().(*pb.ChangeCurrency)
	//充值或购买同步
	t.changeCurrency(arg)
}

// 桌子状态
func (t *Desk) DeskStatus(ctx actor.Context) {
	rsp := new(pb.DeskedStatus)
	r, _ := t.roleCountNum()
	rsp.Idle = r == 0
	ctx.Respond(rsp)
}

// 玩家状态更新
func (t *Desk) UserGameState(ctx actor.Context) {
	arg := ctx.Message().(*pb.UserGameState)
	glog.Debugf("UserGameState %#v", arg)
	if role, ok := t.roles[arg.Userid]; ok {
		role.State = int(arg.State)
		// t.stateCheck()
	}
}

// 点控状态更新
func (t *Desk) PointControl(ctx actor.Context) {
	arg := ctx.Message().(*pb.PointControl)
	glog.Debugf("PointControl %#v", arg)
	if role, ok := t.roles[arg.UserId]; ok {
		role.PCSwitch = arg.Switch
		role.PCFactor = arg.Factor
		role.PCScore = arg.Score
	}
}

func (t *Desk) TriggerFreeWelfare(ctx actor.Context) {
	arg := ctx.Message().(*pb.TriggerFreeWelfare)
	glog.Debugf("TriggerFreeWelfare %#v", arg)
	if role, ok := t.roles[arg.Userid]; ok {
		role.FreeWelfare++
	}
}

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
	case int32(pb.ROOM_TYPE2): //百人
		t.DeskFree = new(data.DeskFree)
		t.FreeDeskNewbiew = new(data.FreeDeskNewbiew)
		t.CPJackpot = 300000000 // 初始化奖池300w
		t.freeInit()
		t.newbiewInit()
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

// 房间消息广播(robot除外)
func (t *Desk) broadcast4(msg interface{}) {
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
		if v.GetRobot() {
			continue
		}
		v.Pid.Tell(msg)
	}
}

// 给玩家发送消息
func (t *Desk) send2userid(userid string, msg interface{}) {
	if v, ok := t.roles[userid]; ok && v != nil {
		if v.Offline {
			return
		}
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
	// glog.Debugf("getRouter %s", ctx.Sender().String())
	return t.router[ctx.Sender().String()]
}

// 获取进程pid,离线也可发送
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

// 获取玩家数据
func (t *Desk) getUserBySeat(seat uint32) *data.User {
	if v, ok := t.seats[seat]; ok && v != nil {
		return t.getPlayer(v.Userid)
	}
	return nil
}

// 获取位置
func (t *Desk) getSeat(userid string) uint32 {
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

func (a *Desk) reqRoom(msg interface{}) interface{} {
	timeout := 3 * time.Second
	res1, err1 := a.roomPid.RequestFuture(msg, timeout).Result()
	if err1 != nil {
		glog.Errorf("reqRoom err: %v, msg %#v", err1, msg)
		return nil
	}
	return res1
}

// 检查跑马灯
func (a *Desk) checkMarquee(userid string, score data.Currency) {
	cp := a.DeskData.Game.LOTTERY
	if score.GetSum() >= cp.Msg_Score {
		if s, ok := a.roles[userid]; ok {
			marquee := &pb.MarQueeNtf{
				Userid:  userid,
				Content: handler.BuildCPMarquee(s.Nickname, score.GetSum()),
			}
			a.rolePid.Tell(marquee)
		}
	}
}

// 检查赠送金和提现金
func (a *Desk) checkGiveDiamond(userid string, score data.Currency) {
	if r, ok := a.roles[userid]; ok {
		if r.Robot {
			return
		}
		msg := &pb.GiveAndOutCash{
			Userid:   userid,
			GameId:   a.GameId,
			Desc:     fmt.Sprintf("房间%s", a.Rid),
			SyncGame: true,
		}
		// 正常玩家才加可提现金，新手不加
		if score.Diamond > 0 && r.Money > 0 {
			// msg.Out = score.Diamond + a.BackOutDiamond[userid]
			msg.Out = score.Diamond
			// r.OutDiamond += score.Diamond
		}
		if score.Diamond < 0 && r.OutDiamond > r.Diamond {
			msg.Out = r.Diamond - r.OutDiamond
		}
		// r.OutDiamond = int64(math.Min(float64(r.OutDiamond), float64(r.Diamond)))
		if r.Offline {
			a.rolePid.Tell(msg)
			return
		}

		a.send2userid(userid, msg)
	}
}

// 配置开什么
func (a *Desk) cfgWinner() int32 {
	// if a.Dtype != int32(pb.DESK_TYPE_NORMAL) {
	// 	return -1
	// }

	err1 := cfg.Reload()
	if err1 != nil {
		return a.cfgWinner()
	}
	conf := cfg.Section(nodeName).Key("winner").Value()
	if conf != "" {
		w, err := strconv.Atoi(conf)
		if err == nil {
			return int32(w)
		}
	}
	return 0
}

func (t *Desk) getUser(userid string) *data.User {
	if u, ok := t.roles[userid]; ok {
		return u.User
	}
	if u, ok := t.NewbiewRobot[userid]; ok {
		return u
	}
	return nil
}

// vim: set foldmethod=marker foldmarker=//',//.:
