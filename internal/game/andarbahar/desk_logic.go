package andarbahar

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/algo"
	"goserver/pkg/game/config"
	"goserver/pkg/game/handler"
	"goserver/pkg/game/login"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"math"
	"strconv"
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

func (a *Desk) GameRechargeAmount(ctx actor.Context) {
	// userid := a.getRouter(ctx)
	arg := ctx.Message().(*pb.GameRechargeAmount)
	rsp := new(pb.GameRechargeAmounted)
	seat := a.getSeat(arg.Userid)
	rsp.Amount, rsp.GiveAmount = a.getRechargeAmount(seat)
	ctx.Respond(rsp)
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
	prevGame := t.DeskGame
	t.DeskGame = new(data.DeskGame)
	switch t.DeskData.Rtype {
	case int32(pb.ROOM_TYPE0): //自由
	case int32(pb.ROOM_TYPE1): //私人
		// 私人房庄家数据保留
		if prevGame != nil {
			t.DeskGame.Dealer = prevGame.Dealer
			t.DeskGame.DealerSeat = prevGame.DealerSeat
		}
		t.changingSeat = make(map[string]*ChangingSeat)
		t.DeskPriv = new(data.DeskPriv)
		t.DeskPriv.PrivPlayer = make(map[string][3]string)
		t.DeskPriv.PrivScore = make(map[string]int64)
		t.DeskPriv.PrivWins = make(map[string]uint32)
		t.DeskPriv.PrivLoses = make(map[string]uint32)
		t.DeskPriv.Joins = make(map[string]uint32)
		t.ABDeskPriv = new(data.ABDeskPriv)
		t.ABDeskPriv.Joker = 0
		t.ABDeskPriv.Bets = make(map[string]int64)
		t.ABDeskPriv.AndarCards = make([]uint32, 0)
		t.ABDeskPriv.BaharCards = make([]uint32, 0)
		t.ABDeskPriv.AndarBets = make(map[string]int64)
		t.ABDeskPriv.BaharBets = make(map[string]int64)
		t.ABDeskPriv.ABScore = make(map[string]int64)
		t.ABDeskPriv.CardRoundTime = 0
		t.ABDeskPriv.CardRound = 0
		t.ABDeskPriv.Winner = 0

		t.ActRechargeTimes = make(map[uint32]int) // 局内充值次数
		// 娱乐模式重置娱乐分
		if t.Gmode == 1 {
			for _, role := range t.roles {
				role.User.SetFraction(int64(config.GetPvpRoom().AbFunInitScore))
				t.sendFraction(role.Userid, 0, int32(pb.LOG_TYPE2)) // 发送娱乐分变更通知
			}
		}
	case int32(pb.ROOM_TYPE2): //百人
		t.DeskFree = new(data.DeskFree)
		t.FreeDeskNewbiew = new(data.FreeDeskNewbiew)
		// t.DeskFree.Dealers = make(map[string]int64) //上庄列表,userid: carry
		// t.DeskFree.Carry = SysCarry
		t.freeInit()
		room := t.DeskData.Game
		initScore := room.AB.Robot.InitScore
		t.NewbiewRobot = make(map[string]*data.User)
		// 新手房间,生成假人,20个
		for i := 0; i < 20; i++ {
			user := &data.User{
				Nickname:   login.RandName(),
				Userid:     fmt.Sprintf("%d", i+1),
				Photo:      strconv.Itoa(utils.RandIntN(30) + 1),
				Diamond:    int64(utils.RandMN(int(initScore[0]), int(initScore[1]))),
				FreeWinMap: make(map[int32][]data.FreeWin),
				Robot:      true,
			}
			// 人机vip头像
			handler.RobotVip(user)
			// 自定义头像
			t.getHead(user)
			user.FreeWinMap[int32(pb.ABAR)] = handler.GenNewbiewFreeWin(int32(pb.ABAR))
			t.NewbiewRobot[user.Userid] = user
		}
		for i := 0; i < 100; i++ {
			// c, _ := utils.WeightedChoice(choices)
			t.ABWinnerSeat = append(t.ABWinnerSeat, uint32(utils.RandInt32N(2)+1)) // 生成对局记录

			// 8条joker 牌记录
			if i >= 92 {
				index := utils.RandIntN(len(algo.NiuCARDS))
				t.ABJoker = append(t.ABJoker, algo.NiuCARDS[index])
			}
		}
		// if t.Dtype == int32(pb.DESK_TYPE_NEWBIEW) || t.Dtype == int32(pb.DESK_TYPE_NORMAL_B) {
		// }
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
		if offline {
			v.PrivOfflineTimeout = utils.Timestamp() + 180
		} else {
			v.PrivOfflineTimeout = 0
		}
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
	ab := a.DeskData.Game.AB
	if score.GetSum() >= ab.Msg_Score {
		if s, ok := a.roles[userid]; ok {
			marquee := &pb.MarQueeNtf{
				Userid:  userid,
				Content: handler.BuildABMarquee(s.Nickname, score.GetSum()),
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

// 配置开什么
func (a *Desk) cfgWinner() int {
	// if a.Dtype != int32(pb.DESK_TYPE_NORMAL) {
	// 	return -1
	// }

	err := cfg.Reload()
	if err != nil {
		return a.cfgWinner()
	}
	conf := cfg.Section(nodeName).Key("times").Value()
	if conf == "" {
		return -1
	}
	w, err := strconv.Atoi(conf)
	if err != nil {
		return -1
	}
	if w <= 0 {
		return -1
	}

	if w%2 == 0 {
		a.ABDeskFree.Winner = 2
	} else {
		a.ABDeskFree.Winner = 1
	}

	if w >= 1 && w <= 5 {
		a.ABDeskFree.SideWinner = 3
	} else if w >= 6 && w <= 10 {
		a.ABDeskFree.SideWinner = 4
	} else if w >= 11 && w <= 15 {
		a.ABDeskFree.SideWinner = 5
	} else if w >= 16 && w <= 25 {
		a.ABDeskFree.SideWinner = 6
	} else if w >= 26 && w <= 30 {
		a.ABDeskFree.SideWinner = 7
	} else if w >= 31 && w <= 35 {
		a.ABDeskFree.SideWinner = 8
	} else if w >= 36 && w <= 40 {
		a.ABDeskFree.SideWinner = 9
	} else {
		a.ABDeskFree.SideWinner = 10
	}
	return w
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

// 获取role
func (t *Desk) getRole(userid string) *data.DeskRole {
	if v, ok := t.roles[userid]; ok && v != nil {
		return v
	}
	return nil
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

// vim: set foldmethod=marker foldmarker=//',//.:
