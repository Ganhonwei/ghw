package crash

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/table"
	"goserver/pkg/utils"
	"math"
	"strings"
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

// 玩家状态更新
func (t *Desk) UserGameState(ctx actor.Context) {
	arg := ctx.Message().(*pb.UserGameState)
	glog.Debugf("UserGameState %#v", arg)
	if role, ok := t.roles[arg.Userid]; ok {
		role.State = int(arg.State)
		// t.stateCheck()
	}
}

// 桌子状态
func (t *Desk) DeskStatus(ctx actor.Context) {
	rsp := new(pb.DeskedStatus)
	r, _ := t.roleCountNum()
	rsp.Idle = r == 0
	ctx.Respond(rsp)
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

func (t *Desk) CrashBoom(ctx actor.Context) {
	t.boomPoint()
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
	glog.Infof("----------------Init Desk: %s, %d", t.Rid, t.DeskData.Rtype)

	t.DeskGame = new(data.DeskGame)
	switch t.DeskData.Rtype {
	case int32(pb.ROOM_TYPE0): //自由
	case int32(pb.ROOM_TYPE1): //私人
		t.DeskPriv = new(data.DeskPriv)
		t.DeskPriv.PrivScore = make(map[string]int64)
		t.DeskPriv.Joins = make(map[string]uint32)
	case int32(pb.ROOM_TYPE2): //百人
		t.DeskFree = new(data.DeskFree)
		t.DeskFree.Dealers = make(map[string]int64) //上庄列表,userid: carry
		t.DeskFree.FreeDeskNewbiew = new(data.FreeDeskNewbiew)
		// t.DeskFree.Carry = SysCarry
		t.NewbiewRobot = make(map[string]*data.User)
		t.freeInit()

		// 生成录单
		t.createHistory()
		// t.CrashJackpot = utils.RandInt64N(1000000)
		// if t.Dtype == int32(pb.DESK_TYPE_NEWBIEW) || t.Dtype == int32(pb.DESK_TYPE_NORMAL_B) {
		// }
	}
}

func (t *Desk) createHistory() {
	t.CRASHHistory = handler.CreateFakeHistory(50)
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
	timeout := 5 * time.Second
	res1, err1 := a.roomPid.RequestFuture(msg, timeout).Result()
	if err1 != nil {
		glog.Errorf("reqRoom err: %v, msg %#v", err1, msg)
		return nil
	}
	return res1
}

// 检查跑马灯
func (a *Desk) checkMarquee(userid string, score data.Currency) {
	msgScore := table.GetTables().CrashRoomTable.Get().MsgScore
	if score.GetSum() >= msgScore {
		if s, ok := a.roles[userid]; ok {
			marquee := &pb.MarQueeNtf{
				Userid:  userid,
				Content: handler.BuildCRASHMarquee(s.Nickname, score.GetSum()),
			}
			a.rolePid.Tell(marquee)
		}
	}
}

// 检查提现金
func (a *Desk) checkOutDiamond(userid string) {
	if r, ok := a.roles[userid]; ok {
		if r.Robot {
			return
		}
		if r.OutDiamond > r.Diamond {
			msg := &pb.GiveAndOutCash{
				Userid: userid,
				GameId: a.GameId,
				Desc:   fmt.Sprintf("房间%s", a.Rid),
			}
			msg.Out = r.Diamond - r.OutDiamond
			r.OutDiamond = min(r.OutDiamond, r.Diamond)
			if r.Offline {
				a.rolePid.Tell(msg)
				return
			}
			a.send2userid(userid, msg)
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
			Userid: userid,
			GameId: a.GameId,
			Desc:   fmt.Sprintf("房间%s", a.Rid),
		}

		// 正常玩家才加可提现金，新手不加
		if score.Diamond > 0 && r.Money > 0 {
			// msg.Out = score.Diamond + a.BackOutDiamond[userid]
			msg.Out = score.Diamond
			r.OutDiamond += score.Diamond
		}
		if r.OutDiamond > r.Diamond {
			msg.Out = r.Diamond - r.OutDiamond
		}
		r.OutDiamond = int64(math.Min(float64(r.OutDiamond), float64(r.Diamond)))
		if r.Offline {
			a.rolePid.Tell(msg)
			return
		}
		a.send2userid(userid, msg)
	}
}

func (t *Desk) getHead(user *data.User) {
	if user.Vip.Lv <= 0 {
		return
	}
	head := handler.GetHead()
	if head != nil {
		user.Nickname = head.Name
		user.Photo = fmt.Sprintf("https://cdn.g2qdh.com/avatar/man/%d.jpg", head.Id)
		user.Sex = head.Sex
	}
}

func (t *Desk) returnHead(u *data.User) {
	if strings.HasPrefix(u.Photo, "http") {
		index := strings.LastIndex(u.Photo, "/")
		str := u.Photo[index+1 : len(u.Photo)-4]
		id := utils.Uint64(str)
		// 归还头像
		handler.ReturnHead(u.Sex, uint32(id))
		glog.Debugf("%s return head %s", u.Userid, u.Photo)
	}
}

// pos 下注位置0或下注位置1
func (t *Desk) checkBetPos(pos int32) bool {
	return pos == 0 || pos == 1
}

// 获取下注位置
func (t *Desk) PosCrashBetsMap(pos int32) map[string]data.Currency {
	if pos == 1 {
		return t.CRASHDeskFree.CrashBets1
	}
	return t.CRASHDeskFree.CrashBets
}

// 获取下一轮下注位置
func (t *Desk) PosCrashNextBetsMap(pos int32) map[string]data.Currency {
	if pos == 1 {
		return t.CRASHDeskFree.CrashNextBets1
	}
	return t.CRASHDeskFree.CrashNextBets
}

// 获取撤离位置
func (t *Desk) PosCrashBackMap(pos int32) map[string]int32 {
	if pos == 1 {
		return t.CRASHDeskFree.CRASHBack1
	}
	return t.CRASHDeskFree.CRASHBack
}

// 获取输赢分
func (t *Desk) PosCrashScoreMap(pos int32) map[string]data.Currency {
	if pos == 1 {
		return t.CRASHDeskFree.CRASHScoreMap1
	}
	return t.CRASHDeskFree.CRASHScoreMap
}

// vim: set foldmethod=marker foldmarker=//',//.:
