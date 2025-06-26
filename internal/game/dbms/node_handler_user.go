package dbms

import (
	"regexp"
	"strings"
	"time"

	"goserver/gen/pb"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/utils"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/go-resty/resty/v2"
)

func (a *DBMSActor) GetConfig(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.GetConfig)
	glog.Debugf("GetConfig %#v", arg)
	//ctx.Respond(handler.GetSyncConfig(arg.Type))
	//同步配置
	a.syncConfig2(ctx.Sender())
}

func (a *DBMSActor) SyncConfig(ctx actor.Context) {
	msg := ctx.Message()
	//同步配置
	arg := msg.(*pb.SyncConfig)
	// glog.Debugf("SyncConfig %#v", arg)
	handler.SyncConfig(arg, a.Name)
}

func (a *DBMSActor) WebRequest(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.WebRequest)
	glog.Debugf("WebRequest %#v", arg)
	rsp := new(pb.WebResponse)
	rsp.Code = arg.Code
	a.handlerWeb(arg, rsp, ctx)
	ctx.Respond(rsp)
}

func (a *DBMSActor) MatchDesk(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.MatchDesk)
	glog.Debugf("MatchDesk %#v", arg)
	a.matchDesk(arg, ctx)
}

func (a *DBMSActor) CreateDesk(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.CreateDesk)
	glog.Debugf("CreateDesk %#v", arg)
	a.createDesk(arg, ctx)
}

func (a *DBMSActor) GetRoomList(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.GetRoomList)
	glog.Debugf("GetRoomList %#v", arg)
	a.getRoomList(arg, ctx)
}

// func (a *DBMSActor) GetGameListReq(ctx actor.Context) {
// 	msg := ctx.Message()
// 	arg := msg.(*pb.GetGameListReq)
// 	glog.Debugf("GetGameListReq %#v", arg)
// 	a.getGameListReq(arg, ctx)
// }

// func (a *DBMSActor) NoticeReq(ctx actor.Context) {
// 	msg := ctx.Message()
// 	arg := msg.(*pb.NoticeReq)
// 	glog.Debugf("NoticeReq %#v", arg)
// 	//TODO 缓存
// 	rsp := handler.PackUserNotice(arg)
// 	ctx.Respond(rsp)
// }

// func (a *DBMSActor) ActivityReq(ctx actor.Context) {
// 	msg := ctx.Message()
// 	arg := msg.(*pb.ActivityReq)
// 	glog.Debugf("ActivityReq %#v", arg)
// 	//TODO 缓存
// 	rsp := handler.PackUserActivity(arg)
// 	ctx.Respond(rsp)
// }

// func (a *DBMSActor) JoinActivityReq(ctx actor.Context) {
// 	msg := ctx.Message()
// 	arg := msg.(*pb.JoinActivityReq)
// 	glog.Debugf("JoinActivityReq %#v", arg)
// 	rsp := handler.JoinActivity(arg)
// 	ctx.Respond(rsp)
// }

// func (a *DBMSActor) AgentProfitRankReq(ctx actor.Context) {
// 	msg := ctx.Message()
// 	arg := msg.(*pb.AgentProfitRankReq)
// 	glog.Debugf("AgentProfitRankReq %#v", arg)
// 	//TODO 缓存
// 	rsp := handler.PackAgentProfitRankMsg(arg)
// 	ctx.Respond(rsp)
// }

// func (a *DBMSActor) AgentManageReq(ctx actor.Context) {
// 	msg := ctx.Message()
// 	arg := msg.(*pb.AgentManageReq)
// 	glog.Debugf("AgentManageReq %#v", arg)
// 	rsp := handler.PackAgentManageMsg(arg)
// 	ctx.Respond(rsp)
// }

// func (a *DBMSActor) AgentProfitManageReq(ctx actor.Context) {
// 	msg := ctx.Message()
// 	arg := msg.(*pb.AgentProfitManageReq)
// 	glog.Debugf("AgentProfitManageReq %#v", arg)
// 	//rsp := handler.PackAgentProfitManageMsg(arg)
// 	rsp := handler.PackAgentProfitManageMsg2(arg)
// 	ctx.Respond(rsp)
// }

// func (a *DBMSActor) AgentPlayerManageReq(ctx actor.Context) {
// 	msg := ctx.Message()
// 	arg := msg.(*pb.AgentPlayerManageReq)
// 	glog.Debugf("AgentPlayerManageReq %#v", arg)
// 	rsp := handler.PackPlayerManageMsg(arg)
// 	ctx.Respond(rsp)
// }

// func (a *DBMSActor) AgentProfitReq(ctx actor.Context) {
// 	msg := ctx.Message()
// 	arg := msg.(*pb.AgentProfitReq)
// 	glog.Debugf("AgentProfitReq: %v", arg)
// 	rsp := handler.PackAgentProfitMsg(arg)
// 	ctx.Respond(rsp)
// }

// func (a *DBMSActor) AgentDayProfitReq(ctx actor.Context) {
// 	msg := ctx.Message()
// 	arg := msg.(*pb.AgentDayProfitReq)
// 	glog.Debugf("AgentDayProfitReq: %v", arg)
// 	rsp := handler.PackAgentDayProfitMsg(arg)
// 	ctx.Respond(rsp)
// }

// func (a *DBMSActor) AgentProfitOrderReq(ctx actor.Context) {
// 	msg := ctx.Message()
// 	arg := msg.(*pb.AgentProfitOrderReq)
// 	glog.Debugf("AgentProfitOrderReq %#v", arg)
// 	rsp := handler.PackAgentProfitOrderMsg(arg)
// 	ctx.Respond(rsp)
// }

// func (a *DBMSActor) AgentOauth2Confirm(ctx actor.Context) {
// 	msg := ctx.Message()
// 	arg := msg.(*pb.AgentOauth2Confirm)
// 	glog.Debugf("AgentOauth2Confirm: %v", arg)
// 	rsp, msg2 := handler.AgentOauth2Confirm(arg)
// 	ctx.Respond(rsp)
// 	if msg2 != nil {
// 		rolePid.Tell(msg2)
// 	}
// }

// func (a *DBMSActor) GetRoomRecord(ctx actor.Context) {
// 	msg := ctx.Message()
// 	arg := msg.(*pb.GetRoomRecord)
// 	glog.Debugf("GetRoomRecord %#v", arg)
// 	rsp := handler.PackRecordMsg(arg)
// 	ctx.Respond(rsp)
// }

func (a *DBMSActor) RobotMsg(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.RobotMsg)
	glog.Debugf("RobotMsg %#v", arg)
	var robotName string
	// robotName = cfg.Section("robot.node" + utils.String(arg.Gtype)).Name()
	if env == "dev" {
		robotName = cfg.Section("robot.node1").Name()
	} else {
		robotName = cfg.Section("robot.node" + utils.String(arg.Gtype)).Name()
	}
	if v, ok := a.serve[robotName]; ok {
		v.Tell(arg)
	}
}

// 匹配节点
func (a *DBMSActor) matchDesk(msg *pb.MatchDesk, ctx actor.Context) {
	rsp := new(pb.MatchedDesk)
	rsp.Gameid = msg.Gameid
	rsp.Roomid = msg.Roomid
	rsp.Rtype = msg.Rtype
	rsp.Gtype = msg.Gtype
	rsp.Dtype = msg.Dtype
	rsp.Ltype = msg.Ltype
	rsp.IsChangeTable = msg.IsChangeTable
	if msg.Name == "" {
		rsp.Error = pb.Failed
		ctx.Respond(rsp)
		return
	}
	for k, v := range a.serve {
		re := regexp.MustCompile(`^(game\.[a-zA-Z0-9_]+)\.(\d+)$`)
		matches := re.FindStringSubmatch(k)
		if len(matches) == 3 {
			nodeName := matches[1]
			nodeId := matches[2]

			_ = nodeId

			if nodeName == msg.Name {
				rsp.Desk = v
				ctx.Respond(rsp)
				return
			}
		} else if strings.Contains(k, msg.Name) &&
			utils.IsInteger(strings.ReplaceAll(k, msg.Name, "")) {
			rsp.Desk = v
			ctx.Respond(rsp)
			return
		}
	}
	glog.Errorf("match desk filed serve %#v, msg %#v", a.serve, msg)
	rsp.Error = pb.Failed
	ctx.Respond(rsp)
}

// 创建房间
func (a *DBMSActor) createDesk(msg *pb.CreateDesk, ctx actor.Context) {
	for k, v := range a.serve {
		if strings.Contains(k, msg.Name) &&
			utils.IsInteger(strings.ReplaceAll(k, msg.Name, "")) {
			v.Tell(msg)
			return
		}
	}
	glog.Errorf("create desk filed serve %#v, msg %#v", a.serve, msg)
	rsp := new(pb.CreatedDesk)
	rsp.Rtype = msg.Rtype
	rsp.Gtype = msg.Gtype
	rsp.Error = pb.Failed
	ctx.Respond(rsp)
}

// 获取房间列表
func (a *DBMSActor) getRoomList(msg *pb.GetRoomList, ctx actor.Context) {
	for k, v := range a.serve {
		if strings.Contains(k, msg.Name) &&
			utils.IsInteger(strings.ReplaceAll(k, msg.Name, "")) {
			v.Tell(msg)
			return
		}
	}
	glog.Errorf("get room filed serve %#v, msg %#v", a.serve, msg)
	rsp := new(pb.GotRoomList)
	rsp.Rtype = msg.Rtype
	rsp.Gtype = msg.Gtype
	rsp.Error = pb.Failed
	ctx.Respond(rsp)
}

// 游戏是否开启
func (a *DBMSActor) IsGameOpen(game_type int32) bool {
	var game_name string
	switch game_type {
	case int32(pb.HUA):
		game_name = "hua"
	case int32(pb.LHD):
		game_name = "lhd"
	case int32(pb.SEVEN):
		game_name = "seven"
	case int32(pb.RUMMY):
		game_name = "rummy"
	case int32(pb.AK47):
		game_name = "ak47"
	case int32(pb.JOKER):
		game_name = "joker"
	case int32(pb.CRASH):
		game_name = "crash"
	case int32(pb.RUMMY2):
		game_name = "rummy_2"
	case int32(pb.HUA2):
		game_name = "hua_2"
	}

	for k, _ := range a.serve {
		if strings.Contains(k, game_name) &&
			utils.IsInteger(strings.ReplaceAll(k, game_name, "")) {
			return true
		}
	}
	return false
}

// 获取游戏列表
// func (a *DBMSActor) getGameListReq(msg *pb.GetGameListReq, ctx actor.Context) {
// 	rsp := &pb.GetGameListRsp{}
// 	all_games := config.GetGames()
// 	var games data.GameList
// 	for _, v := range all_games {
// 		if msg.GameType != v.Gtype {
// 			continue
// 		}
// 		if v.Status == 0 {
// 			continue
// 		}
// 		// if a.IsGameOpen(v.Gtype) {
// 		// 	games = append(games, v)
// 		// }
// 	}

// 	sort.Sort(games)
// 	for _, v := range games {
// 		info := &pb.GameListInfo{
// 			Id:        v.Id,
// 			GameType:  v.Gtype,
// 			MinAccess: int32(v.Min_Access),
// 			MaxAccess: int32(v.Max_Access),
// 			Sort:      int32(v.SortId),
// 			Bet:       v.GetBottom(),
// 		}
// 		rsp.List = append(rsp.List, info)
// 	}

// 	ctx.Respond(rsp)
// }

func (a *DBMSActor) GamePeopleNtf(ctx actor.Context) {
	arg := ctx.Message().(*pb.GamePeopleNtf)
	for k, p := range a.serve {
		if strings.Contains(k, "gate.") {
			p.Tell(arg)
		}
	}
}

func (a *DBMSActor) SendSmsCode(ctx actor.Context) {
	arg := ctx.Message().(*pb.SendSmsCode)
	// for k, p := range a.serve {
	// 	if strings.Contains(k, "sms") {
	// 		p.Tell(arg)
	// 		return
	// 	}
	// }
	type SmsRequest struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}

	type SmsResponse struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	go func() {
		client := resty.New().SetTimeout(3 * time.Second)
		var result SmsResponse
		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(SmsRequest{
				Phone: arg.Phone,
				Code:  arg.Code,
			}).
			SetResult(&result).
			Post(smsService + "/api/sendSMSCode")

		if err != nil {
			glog.Errorf("send sms code error with %v:", err)
			return
		}

		if resp.StatusCode() != 200 || result.Code != 200 {
			glog.Errorf("发送短信API返回错误: code=%d, message=%s",
				result.Code, result.Message)
			return
		}
	}()
}

func (a *DBMSActor) RobotBaseGet(ctx actor.Context) {
	arg := ctx.Message().(*pb.RobotBaseGet)
	glog.Debugf("RobotBaseGet %#v", arg)
	var robotName string
	var robotNode *actor.PID
	// 查询robot节点
	if env == "dev" {
		robotName = cfg.Section("robot.node1").Name()
		if v, ok := a.serve[robotName]; ok {
			robotNode = v
		}
	} else {
		for i := 1; i <= 10; i++ {
			robotName = cfg.Section("robot.node" + utils.String(i)).Name()
			if v, ok := a.serve[robotName]; ok {
				robotNode = v
				break
			}
		}
	}
	rsp := new(pb.RobotBaseGeted)
	if robotNode == nil {
		glog.Error("get robot base not found robot server node")
		rsp.Error = pb.Failed
		ctx.Respond(rsp)
		return
	}

	r, err := robotNode.RequestFuture(arg, 3*time.Second).Result()
	if err != nil {
		glog.Error("get robot base error with %v:", err)
		rsp.Error = pb.Failed
		ctx.Respond(rsp)
		return
	}
	ctx.Respond(r)
}
