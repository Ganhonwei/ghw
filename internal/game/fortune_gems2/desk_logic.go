package fortune_gems2

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"goserver/pkg/myactor"
	"math"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// NewDesk 新建一张牌桌
func NewDesk(deskData *data.DeskData) *Desk {
	desk := &Desk{
		router: make(map[string]string),
		stopCh: make(chan struct{}),
	}
	desk.DeskData = deskData
	return desk
}

// InitDesk 初始化
func (t *Desk) InitDesk() {
	t.state = int32(pb.STATE_FREE)
	t.minesPits = make([]int32, 25)
	t.minesPitsUser = make([]int32, 25)

	// t.gameInit()
	// t.DeskGame = new(data.DeskGame)
	// switch t.DeskData.Rtype {
	// case int32(pb.ROOM_TYPE0): //自由
	// 	// t.FakeSeats = handler.CreateFakeSeats(1, t.seats, t.DeskData.Count)
	// case int32(pb.ROOM_TYPE1): //私人
	// case int32(pb.ROOM_TYPE2): //百人
	// }
}

// 给玩家发送消息
func (t *Desk) send2user(msg interface{}) {
	if t.role != nil {
		t.role.Pid.Tell(msg)
	}
}

// 获取路由
func (t *Desk) getRouter(ctx actor.Context) string {
	glog.Debugf("getRouter %s", ctx.Sender().String())
	return t.router[ctx.Sender().String()]
}

// 获取进程pid
func (t *Desk) getPid() *actor.PID {
	if t.role != nil {
		return t.role.Pid
	}
	return nil
}

// 获取玩家数据
func (t *Desk) getPlayer() *data.User {
	if t.role != nil {
		return t.role.User
	}
	return nil
}

// 获取role
func (t *Desk) getRole() *data.DeskRole {
	return t.role
}

// 玩家是否在线
func (t *Desk) isOnline() bool {
	if t.role != nil {
		return !t.role.Offline
	}
	return false
}

// 设置玩家是否离线
func (t *Desk) setOffline(offline bool) {
	if t.role != nil {
		t.role.Offline = offline
	}
}

// 事件发送
func (t *Desk) eventPost(eventId int32, bean any) {
	body, err := json.Marshal(bean)
	if err != nil {
		glog.Errorf("lhd event Marshal fail")
		return
	}
	msg := new(pb.EventPost)
	msg.EventId = eventId
	msg.Data = body
	t.send2user(msg)
}

// 检查赠送金和提现金
func (a *Desk) checkGiveDiamond(score int64) {
	if r := a.role; r != nil {
		if r.Robot {
			return
		}
		msg := &pb.GiveAndOutCash{
			Userid: r.Userid,
			GameId: a.GameId,
			Desc:   fmt.Sprintf("房间%s", a.Rid),
		}

		// 正常玩家才加可提现金，新手不加
		if score > 0 && r.Money > 0 {
			// msg.Out = score.Diamond + a.BackOutDiamond[userid]
			msg.Out = score
			r.OutDiamond += score
		}
		if r.OutDiamond > r.Diamond {
			msg.Out = r.Diamond - r.OutDiamond
		}
		r.OutDiamond = int64(math.Min(float64(r.OutDiamond), float64(r.Diamond)))
		if r.Offline {
			a.rolePid.Tell(msg)
			return
		}
		a.send2user(msg)
	}
}

func (t *Desk) shareAmount(score int64) {
	if role := t.role; role != nil {
		if role.ShareSuperior == "" || role.State != 2 {
			return
		}
		msg := &pb.ShareBetAmount{
			Score:    score,
			Userid:   role.Userid,
			Superior: role.ShareSuperior,
			Name:     role.Nickname,
		}
		t.roomPid.Tell(msg)
	}
}

func (t *Desk) GameTime() {
	if role := t.role; role != nil && t.detail != nil {
		if role.Robot {
			return
		}

		msg := &pb.LogGameTime{
			Userid: role.Userid,
			Gtype:  int32(pb.FORTUNE_GEMS2),
			Time:   time.Now().Unix() - t.detail.BeginTime,
		}
		myactor.Logger().Tell(msg)
		t.send2user(msg)
	}
}
