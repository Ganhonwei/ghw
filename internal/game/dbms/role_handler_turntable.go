package dbms

import (
	"context"
	"encoding/base64"
	"fmt"
	"goserver/gen/pb"
	"goserver/gen/tb"
	"goserver/pkg/data"
	"goserver/pkg/data/ck"
	"goserver/pkg/glog"
	"goserver/pkg/myredis"
	"goserver/pkg/table"
	"goserver/pkg/utils"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/redis/go-redis/v9"
	"gopkg.in/mgo.v2/bson"
)

const (
	turntableDrawLogKey         = "activity:turn:user_draws"         // 玩家抽奖日志
	turntablePrizeMarqueeKey    = "activity:turn:marquees"           // 跑马灯奖励记录
	turntableUserPrizeKey       = "activity:turn:prize"              // 转盘奖励领取记录缓存
	turntableMarqueeNextTimeKey = "activity:turn:marquees_next_time" // 转盘跑马灯下次假数据生成时间
)

var (
	turntableMarqueeNextTime int64
)

// InitTurntable 初始化转盘活动
func (a *RoleActor) ActivityTurnInit() {
}

// 转盘时钟
func (a *RoleActor) activityTurnTick() {
	a.activityTurnMarqueeGenTimeTick()
}

// 转盘跑马灯假数据生成
func (a *RoleActor) activityTurnMarqueeGenTimeTick() {
	genTime := table.GetTables().TurnBaseTable.Get().MarqueeGenTime
	if len(genTime) != 2 || (genTime[0] <= 0 && genTime[1] <= 0) {
		return
	}
	nowSec := time.Now().Unix()

	if turntableMarqueeNextTime > 0 {
		if nowSec > turntableMarqueeNextTime {
			defer func() {
				// 重置下一次生成时间
				turntableMarqueeNextTime = 0
				if err := myredis.Redis().Del(context.Background(), turntableMarqueeNextTimeKey).Err(); err != nil {
					glog.Error(err)
				}
			}()

			// 生成领奖记录
			list := table.GetTables().TurnTable.GetDataList()
			turnTargets := list[utils.RandMN(0, len(list)-1)].Targets
			var scoreType, scoreTarget int32
			for i, target := range turnTargets {
				if target > 0 {
					scoreTarget = target
					scoreType = int32(i) + 1
					break
				}
			}

			// 获取一条人机基本信息
			robotVipLv := int32(utils.RandMN(0, 10))
			req := &pb.RobotBaseGet{VipCount: map[int32]int32{robotVipLv: 1}}
			r, err := nodePid.RequestFuture(req, 5*time.Second).Result()
			rsp, ok := r.(*pb.RobotBaseGeted)
			if !ok || rsp.Error != pb.OK {
				glog.Errorf("fetch robot base error: %#v,%v", rsp, err)
				return
			}
			if robots, ok := rsp.VipRobots[robotVipLv]; ok && len(robots.Robots) > 0 {
				robot := robots.Robots[0]
				marquee := &pb.ActivityTurnMarquee{
					Id:        bson.NewObjectId().Hex(),
					Userid:    robot.Id,
					Username:  robot.Nickname,
					Photo:     robot.Photo,
					Score:     int64(scoreTarget),
					ScoreType: scoreType,
				}
				a.activityTurnMarqueeSave(marquee)
			}
		}
		return
	}

	var nextTime int64
	r, err := myredis.Redis().Get(context.Background(), turntableMarqueeNextTimeKey).Result()
	if err != nil {
		if err != redis.Nil {
			glog.Error("turntableMarqueeNextTimeKey get error:", err)
		}
	} else {
		if a, e := strconv.ParseInt(r, 10, 64); e == nil {
			nextTime = a
		}
	}
	if nextTime == 0 {
		nextSec := int64(utils.RandMN(int(genTime[0])*60, int(genTime[1])*60))
		nextTime = nowSec + nextSec
	}
	turntableMarqueeNextTime = nextTime
	if err = myredis.Redis().Set(context.Background(), turntableMarqueeNextTimeKey, strconv.Itoa(int(turntableMarqueeNextTime)), 0).Err(); err != nil {
		glog.Error(err)
	}
}

func (a *RoleActor) activityTurnDrawLogSave(userid string, log *data.ActivityTurnDrawLog) {
	log.Id = bson.NewObjectId().Hex()

	// save draw log redis cache
	r := &pb.ActivityTurnDrawLog{
		Rid:         time.Now().Unix(),
		Id:          log.Id,
		Userid:      log.Userid,
		Username:    log.Username,
		Photo:       log.Photo,
		Reason:      log.Reason,
		DrawScore:   log.DrawScore,
		Ctime:       time.Unix(log.Ctime, 0).In(location).Format(utils.FORMAT),
		ScoreBefore: log.ScoreBefore,
		Score:       log.Score,
	}
	logBytes, err := r.Marshal()
	if err != nil {
		glog.Error("activity turn draw log marshal error:", err)
	} else {
		logCache := base64.StdEncoding.EncodeToString(logBytes)
		if err := myredis.Redis().ZAdd(context.Background(), fmt.Sprintf("%s:%s", turntableDrawLogKey, userid), redis.Z{Score: float64(r.Rid), Member: logCache}).Err(); err != nil {
			glog.Error("activity turn draw log cache save error:", err)
		}
	}

	go log.Save()
}

func (a *RoleActor) activityTurnGetUserTurnTable(user *data.User) (record tb.ActivityTurnRecord) {
	turnTable := table.GetTables().TurnTable.Get(user.ActivityTurnPrizeTimes)
	if turnTable == nil {
		list := table.GetTables().TurnTable.GetDataList()
		turnTable = list[len(list)-1]
	}
	return *turnTable
}

func (a *RoleActor) activityTurnGetUserTurn(user *data.User) (turn *data.ActivityTurn, err error) {
	nowSec := time.Now().Unix()
	if user.ActivityTurn != nil {
		turn = user.ActivityTurn
		// 老用户其他3个展示序幕金
		if len(turn.GiveAmount3) == 0 {
			turnTable := a.activityTurnGetUserTurnTable(user)
			// 其他3个展示序幕金
			give3 := []int64{
				int64(utils.RandMN(int(turnTable.Gives[0])+1, int(turnTable.Gives[1])-1)),
				int64(utils.RandMN(int(turnTable.Gives[0])+1, int(turnTable.Gives[1])-1)),
				int64(utils.RandMN(int(turnTable.Gives[0])+1, int(turnTable.Gives[1])-1)),
			}
			turn.GiveAmount3 = give3
			a.activityTurnSync(user)
		}
		// 本轮还未打开序幕礼盒
		if turn.TurnEtime == 0 {
			return
		}
		// 本轮结束时间未到
		if nowSec < turn.TurnEtime {
			if turn.Score < turn.ScoreTarget &&
				turn.DrawTimesFree == 0 &&
				nowSec >= turn.NextFreeDrawTime {
				turn.DrawTimesFree++ // 免费抽奖次数增加
				// turn同步gate
				a.activityTurnSync(user)
			}
			return
		}
	}

	// 初始化用户活动数据
	turnTable := a.activityTurnGetUserTurnTable(user)
	var scoreType, scoreTarget int32
	for i, target := range turnTable.Targets {
		if target > 0 {
			scoreTarget = target
			scoreType = int32(i) + 1
			break
		}
	}
	if scoreType == 0 {
		err = fmt.Errorf("activity turn targe prize error: [%v]", turnTable.Targets)
		return
	}
	// 赠送序幕金
	give := utils.RandMN(int(turnTable.Gives[0])+1, int(turnTable.Gives[1])-1)
	if give <= 0 {
		err = fmt.Errorf("activity turn targe gives error: [%d-%d]", turnTable.Gives[0], turnTable.Gives[1])
		return
	}
	// 其他3个展示序幕金
	give3 := []int64{
		int64(utils.RandMN(int(turnTable.Gives[0])+1, int(turnTable.Gives[1])-1)),
		int64(utils.RandMN(int(turnTable.Gives[0])+1, int(turnTable.Gives[1])-1)),
		int64(utils.RandMN(int(turnTable.Gives[0])+1, int(turnTable.Gives[1])-1)),
	}

	turn = &data.ActivityTurn{
		TurnTime:      0,
		TurnEtime:     0,
		GiveAmount:    int64(give), // 序幕金
		GiveAmount3:   give3,
		DrawTimesFree: 1, // 免费赠送一次抽奖
		Score:         int64(give),
		ScoreTarget:   int64(scoreTarget),
		ScoreType:     scoreType,
	}
	user.ActivityTurn = turn

	// 删除上一轮抽奖记录
	if err = myredis.Redis().Del(context.Background(), fmt.Sprintf("%s:%s", turntableDrawLogKey, user.Userid)).Err(); err != nil {
		glog.Error(err)
		return
	}

	// turn同步gate
	a.activityTurnSync(user)

	return user.ActivityTurn, nil
}

// 转盘数据同步gate
func (a *RoleActor) activityTurnSync(user *data.User) {
	if role, ok := a.roles[user.Userid]; ok {
		sync := &pb.ActivityTurnSync{
			TurnTime:               user.ActivityTurn.TurnTime,
			TurnEtime:              user.ActivityTurn.TurnEtime,
			GiveSelected:           user.ActivityTurn.GiveSelected,
			GiveAmount:             user.ActivityTurn.GiveAmount,
			GiveAmount3:            user.ActivityTurn.GiveAmount3,
			GiveIndex:              user.ActivityTurn.GiveIndex,
			DrawedTimes:            user.ActivityTurn.DrawedTimes,
			DrawTimesFree:          user.ActivityTurn.DrawTimesFree,
			DrawTimesInvite:        user.ActivityTurn.DrawTimesInvite,
			NextFreeDrawTime:       user.ActivityTurn.NextFreeDrawTime,
			Score:                  user.ActivityTurn.Score,
			ScoreTarget:            user.ActivityTurn.ScoreTarget,
			ScoreType:              user.ActivityTurn.ScoreType,
			TackedPrize:            user.ActivityTurn.TackedPrize,
			DrawedTimesFree:        user.ActivityTurn.DrawedTimesFree,
			DrawedTimesInvite:      user.ActivityTurn.DrawedTimesInvite,
			DrawTimesLucky:         user.ActivityTurn.DrawTimesLucky,
			ActivityTurnPrizeTimes: user.ActivityTurnPrizeTimes,
		}
		role.Pid.Tell(sync)
	}
}

// 打码排行榜弹窗
func (a *RoleActor) ActivityBetRankWindowNtfs(ctx actor.Context) {
	arg := ctx.Message().(*pb.ActivityBetRankWindowNtfs)
	if role, ok := a.roles[arg.Userid]; ok {
		for _, ntf := range arg.Ntf0 {
			role.Pid.Tell(ntf)
		}
		for _, ntf := range arg.Ntf1 {
			role.Pid.Tell(ntf)
		}
		for _, ntf := range arg.Ntf2 {
			role.Pid.Tell(ntf)
		}
		for _, ntf := range arg.Ntf3 {
			role.Pid.Tell(ntf)
		}
	}
}

// ActivityTurnShareRegist 转盘分享注册消息
func (a *RoleActor) ActivityTurnShareRegist(ctx actor.Context) {
	arg := ctx.Message().(*pb.ActivityTurnShareRegist)

	superior := a.getUserById(arg.ShareSuperior)
	if superior == nil {
		return
	}
	turn, err := a.activityTurnGetUserTurn(superior)
	if err != nil {
		glog.Error("activityTurnGetUserTurn error: ", err)
		return
	}
	// 邀请抽奖次数加一
	if turn.TurnEtime > 0 && turn.GiveSelected {
		turn.DrawTimesInvite++
		if arg.NewAdid {
			turn.DrawTimesLucky++
		}
		// turn同步gate
		a.activityTurnSync(superior)
		// turn同步前端
		a.activityTurnNtf(superior.Userid)
	}
}

// ActivityTurnReq 转盘主界面数据主界面
func (a *RoleActor) ActivityTurnReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ActivityTurnReq)
	rsp := a.activityTurnRspPack(arg.Userid)
	ctx.Respond(rsp)
}

func (a *RoleActor) activityTurnRspPack(userid string) (rsp *pb.ActivityTurnRsp) {
	rsp = new(pb.ActivityTurnRsp)

	user := a.getUser(userid)
	if user == nil {
		rsp.Error = pb.Failed
		return
	}
	turn, err := a.activityTurnGetUserTurn(user)
	if err != nil {
		glog.Error("activityTurnGetUserTurn error: ", err)
		rsp.Error = pb.Failed
		return
	}

	turnTable := a.activityTurnGetUserTurnTable(user)
	rsp.GiveSelected = turn.GiveSelected
	rsp.GiveAmount = turn.GiveAmount
	rsp.GiveIndex = turn.GiveIndex
	rsp.DrawTimes = turn.DrawTimesFree + turn.DrawTimesInvite
	rsp.DrawTimesFree = turn.DrawTimesFree
	rsp.DrawTimesInvite = turn.DrawTimesInvite
	rsp.Etime = turn.TurnEtime
	rsp.NextFreeDrawTime = turn.NextFreeDrawTime
	rsp.Score = turn.Score
	rsp.ScoreTarget = turn.ScoreTarget
	rsp.ScoreType = turn.ScoreType
	rsp.CanTackPrize = turn.Score >= turn.ScoreTarget
	rsp.TackedPrize = turn.TackedPrize
	rsp.TurnRoundTime = turnTable.TurnRoundTime
	rsp.FreeDrawTime = int32(turnTable.FreeDrawTime)
	rsp.GiveAmount3 = turn.GiveAmount3

	// 抽奖历史
	results, err := myredis.Redis().ZRevRange(context.Background(), fmt.Sprintf("%s:%s", turntableDrawLogKey, userid), 0, 20).Result()
	if err != nil && err != redis.Nil {
		glog.Error(err)
		rsp.Error = pb.Failed
		return
	}

	for _, ret := range results {
		bytes, err := base64.StdEncoding.DecodeString(ret)
		if err != nil {
			glog.Error(err)
			rsp.Error = pb.Failed
			return
		}
		log := new(pb.ActivityTurnDrawLog)
		if err = log.Unmarshal(bytes); err != nil {
			glog.Error(err)
			rsp.Error = pb.Failed
			return
		}
		rsp.DrawLogs = append(rsp.DrawLogs, log)
	}

	// 跑马灯
	marquees, err := myredis.Redis().LRange(context.Background(), turntablePrizeMarqueeKey, 0, 99).Result()
	if err != nil && err != redis.Nil {
		glog.Error(err)
		rsp.Error = pb.Failed
		return
	}
	for _, marqueeB64 := range marquees {
		marqueeBytes, err := base64.StdEncoding.DecodeString(marqueeB64)
		if err != nil {
			glog.Error(err)
			rsp.Error = pb.Failed
			return
		}
		marquee := new(pb.ActivityTurnMarquee)
		if err = marquee.Unmarshal(marqueeBytes); err != nil {
			glog.Error(err)
			rsp.Error = pb.Failed
			return
		}
		rsp.Marquees = append(rsp.Marquees, marquee)
	}
	return
}

// ActivityTurnDrawReq 转盘抽奖请求
func (a *RoleActor) ActivityTurnDrawReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ActivityTurnDrawReq)
	rsp := new(pb.ActivityTurnDrawRsp)

	user := a.getUser(arg.Userid)
	if user == nil {
		rsp.Error = pb.Failed
		ctx.Respond(rsp)
		return
	}
	turn, err := a.activityTurnGetUserTurn(user)
	if err != nil {
		glog.Error("activityTurnGetUserTurn error: ", err)
		rsp.Error = pb.Failed
		ctx.Respond(rsp)
		return
	}

	// 没有抽奖次数
	if turn.DrawTimesFree <= 0 && turn.DrawTimesInvite <= 0 {
		glog.Errorf("activityTurn turnDrawTimesNotEnough: %s", user.Userid)
		rsp.Error = pb.Failed
		ctx.Respond(rsp)
		return
	}
	// 目标金额已达到可以领奖
	if turn.Score >= turn.ScoreTarget {
		glog.Errorf("activityTurn score = target: %s", user.Userid)
		rsp.Error = pb.Failed
		ctx.Respond(rsp)
		return
	}
	turnTable := a.activityTurnGetUserTurnTable(user)
	drawOutTimes := turnTable.DrawOutTimes[user.RegistArea] // 抽满次数
	clawBackRate := turnTable.ClawBackRate                  // 递减率

	log := &data.ActivityTurnDrawLog{
		TurnStime:   turn.TurnTime,
		TurnEtime:   turn.TurnEtime,
		Userid:      user.Userid,
		Username:    user.Nickname,
		Photo:       user.Photo,
		ChannelId:   user.AD_BundleId,
		RegistArea:  int8(user.RegistArea),
		Rtime:       user.Ctime.Unix(),
		Ctime:       time.Now().Unix(),
		ScoreBefore: turn.Score,
		ScoreTarget: turn.ScoreTarget,
		PrizeType:   turn.ScoreType,
	}

	var drawScore int64
	if turn.DrawedTimes >= drawOutTimes-1 {
		// 已达到抽奖次数
		drawScore = turn.ScoreTarget - turn.Score
	} else {
		drawScore = int64(float64(turn.ScoreTarget-turn.Score) * clawBackRate)
	}
	if turn.DrawTimesFree > 0 {
		turn.DrawTimesFree--
		log.Reason = 1
	} else if turn.DrawTimesInvite > 0 {
		turn.DrawTimesInvite--
		log.Reason = 2

		// 有效邀请中满额概率
		if turn.DrawTimesLucky > 0 {
			turn.DrawTimesLucky--
			log.LuckyDraw = 1
			luckyRate := turnTable.LuckyInviteRate[user.RegistArea]
			if utils.RandWan(luckyRate) {
				// 直接抽满
				drawScore = turn.ScoreTarget - turn.Score
				log.LuckyDraw = 2
			}
		}
	}

	turn.DrawedTimes++
	turn.Score += drawScore

	nowSec := time.Now().Unix()
	if turn.Score < turn.ScoreTarget &&
		turn.DrawTimesFree == 0 &&
		nowSec > turn.NextFreeDrawTime {
		// 重置玩家下次免费抽奖时间
		turn.NextFreeDrawTime = nowSec + int64(math.Ceil(turnTable.FreeDrawTime*60*60))
	}

	// turn同步gate
	a.activityTurnSync(user)

	log.DrawScore = drawScore
	log.Score = turn.Score
	a.activityTurnDrawLogSave(user.Userid, log)

	rsp.DrawScore = drawScore
	rsp.Score = turn.Score
	rsp.ScoreTarget = turn.ScoreTarget
	rsp.CanTackPrize = turn.Score >= turn.ScoreTarget
	rsp.Turn = a.activityTurnRspPack(arg.Userid)

	ctx.Respond(rsp)
}

// ActivityTurnClickLogReq 点击序幕礼盒、点击invite按钮记录
func (a *RoleActor) ActivityTurnClickLogReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ActivityTurnClickLogReq)
	rsp := &pb.ActivityTurnClickLogRsp{}

	user := a.getUser(arg.Userid)
	if user == nil {
		rsp.Error = pb.Failed
		ctx.Respond(rsp)
		return
	}
	turn, err := a.activityTurnGetUserTurn(user)
	if err != nil {
		glog.Error("activityTurnGetUserTurn error: ", err)
		rsp.Error = pb.Failed
		ctx.Respond(rsp)
		return
	}

	nowSec := time.Now().Unix()
	if arg.Ctype == 0 { // 打开序幕礼盒
		turn.GiveIndex = arg.Index

	} else if arg.Ctype == 1 { // 领取序幕金
		if turn.GiveSelected {
			glog.Error("activity turn give repeat: ", user.Userid)
			rsp.Error = pb.Failed
			ctx.Respond(rsp)
			return
		}
		turn.GiveSelected = true
		turn.TurnTime = nowSec
		turn.TurnEtime = nowSec + int64(a.activityTurnGetUserTurnTable(user).TurnRoundTime*60*60)

		// 抽奖日志
		log := &data.ActivityTurnDrawLog{
			TurnStime:   turn.TurnTime,
			TurnEtime:   turn.TurnEtime,
			Userid:      user.Userid,
			Username:    user.Nickname,
			Photo:       user.Photo,
			ChannelId:   user.AD_BundleId,
			RegistArea:  int8(user.RegistArea),
			Rtime:       user.Ctime.Unix(),
			Reason:      0,
			DrawScore:   turn.GiveAmount,
			Ctime:       time.Now().Unix(),
			ScoreBefore: 0,
			Score:       turn.Score,
			ScoreTarget: turn.ScoreTarget,
			PrizeType:   turn.ScoreType,
		}
		a.activityTurnDrawLogSave(user.Userid, log)
	}

	// turn同步gate
	a.activityTurnSync(user)

	ctx.Respond(rsp)
}

// ActivityTurnDrawLogReq 本轮抽奖记录 滚动分页请求
func (a *RoleActor) ActivityTurnDrawLogReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ActivityTurnDrawLogReq)
	rsp := &pb.ActivityTurnDrawLogRsp{}

	rangeBy := &redis.ZRangeBy{Max: fmt.Sprint(arg.PrevRid - 1), Min: "0", Offset: 0, Count: arg.PageSize}
	results, err := myredis.Redis().ZRevRangeByScore(context.Background(), fmt.Sprintf("%s:%s", turntableDrawLogKey, arg.Userid), rangeBy).Result()
	if err != nil {
		glog.Error(err)
		rsp.Error = pb.Failed
		ctx.Respond(rsp)
		return
	}

	for _, ret := range results {
		bytes, err := base64.StdEncoding.DecodeString(ret)
		if err != nil {
			glog.Error(err)
			rsp.Error = pb.Failed
			ctx.Respond(rsp)
			return
		}
		log := new(pb.ActivityTurnDrawLog)
		if err = log.Unmarshal(bytes); err != nil {
			glog.Error(err)
			rsp.Error = pb.Failed
			ctx.Respond(rsp)
			return
		}
		rsp.Logs = append(rsp.Logs, log)
	}
	ctx.Respond(rsp)
}

// ActivityTurnTackPrizeReq 转盘申请领奖
func (a *RoleActor) ActivityTurnTackPrizeReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ActivityTurnTackPrizeReq)
	rsp := &pb.ActivityTurnTackPrizeRsp{}

	user := a.getUser(arg.Userid)
	if user == nil {
		rsp.Error = pb.Failed
		ctx.Respond(rsp)
		return
	}
	turn, err := a.activityTurnGetUserTurn(user)
	if err != nil {
		glog.Error("activityTurnGetUserTurn error: ", err)
		rsp.Error = pb.Failed
		ctx.Respond(rsp)
		return
	}

	// 已领取过
	if turn.TackedPrize {
		rsp.Error = pb.Failed
		ctx.Respond(rsp)
		return
	}
	// 分数不够
	if turn.Score < turn.ScoreTarget {
		rsp.Error = pb.Failed
		ctx.Respond(rsp)
		return
	}
	rsp.Prize = turn.ScoreTarget
	rsp.PrizeType = turn.ScoreType

	// 领奖记录保存
	prize := &data.ActivityTurnPrizeLog{
		Id:         bson.NewObjectId().Hex(),
		TurnStime:  turn.TurnTime,
		TurnEtime:  turn.TurnEtime,
		Userid:     user.Userid,
		ChannelId:  user.AD_BundleId,
		RegistArea: int8(user.RegistArea),
		Rtime:      user.Ctime.Unix(),
		Prize:      turn.ScoreTarget,
		PrizeType:  turn.ScoreType,
		Ctime:      time.Now().Unix(),
		State:      0,
		Stime:      0,
		Stype:      0,
	}
	prize.DrawedTimesFree = turn.DrawedTimesFree
	prize.DrawedTimesInvite = turn.DrawedTimesInvite
	prize.Gives = turn.GiveAmount
	prize.Save()

	// 缓存领奖记录
	prizeBytes, err := (&pb.ActivityTurnPrize{
		Id:        prize.Id,
		State:     prize.State,
		Prize:     prize.Prize,
		PrizeType: prize.PrizeType,
		Ctime:     time.Unix(prize.Ctime, 0).In(location).Format(utils.FORMAT),
		CtimeSec:  prize.Ctime,
	}).Marshal()
	if err != nil {
		glog.Error(err)
	} else {
		prizeB64 := base64.StdEncoding.EncodeToString(prizeBytes)
		if err = myredis.Redis().HSet(context.Background(), fmt.Sprintf("%s:%s", turntableUserPrizeKey, prize.Userid), prize.Id, prizeB64).Err(); err != nil {
			glog.Error(err)
		}
	}

	turn.TackedPrize = true
	user.ActivityTurnPrizeTimes++

	// 机审
	go a.activityTurnPrizeAudit(prize, turn.TurnTime, turn.TurnEtime, user.ActivityTurnPrizeTimes)

	// 领取奖励本轮结束, 重置用户转盘数据
	turn.TurnEtime = time.Now().Unix()
	_, err = a.activityTurnGetUserTurn(user)
	if err != nil {
		glog.Error("tack prize reset turn error:", err)
	}

	// turn同步gate
	a.activityTurnSync(user)

	// 缓存领奖记录
	marquee := &pb.ActivityTurnMarquee{
		Id:        prize.Id,
		Userid:    user.Userid,
		Username:  user.Nickname,
		Photo:     user.Photo,
		Score:     prize.Prize,
		ScoreType: prize.PrizeType,
	}
	a.activityTurnMarqueeSave(marquee)

	ctx.Respond(rsp)
	a.activityTurnNtf(arg.Userid)
}

// 缓存领奖记录到跑马灯
func (a *RoleActor) activityTurnMarqueeSave(marquee *pb.ActivityTurnMarquee) {
	if marqueeBytes, err := marquee.Marshal(); err == nil {
		ctx := context.Background()
		marqueeBase64 := base64.StdEncoding.EncodeToString(marqueeBytes)
		if err = myredis.Redis().LPush(ctx, turntablePrizeMarqueeKey, marqueeBase64).Err(); err != nil {
			glog.Error(err)
		}
		// 保留最新100条
		if err = myredis.Redis().LTrim(ctx, turntablePrizeMarqueeKey, 0, 99).Err(); err != nil {
			glog.Error(err)
		}
	}
}

// 通知转盘更新
func (a *RoleActor) activityTurnNtf(userid string) {
	turnNtf := a.activityTurnRspPack(userid)
	ntf := &pb.ActivityTurnNtf{Turn: turnNtf}
	if role, ok := a.roles[userid]; ok {
		role.Pid.Tell(ntf)
	}
}

// 转盘领奖机审
func (a *RoleActor) activityTurnPrizeAudit(prize *data.ActivityTurnPrizeLog, turnTime, turnEtime int64, activityTurnPrizeTimes int32) {
	pass, reasons, err := a.activityTurnTackPrizeAuditMachine(prize, turnTime, turnEtime, activityTurnPrizeTimes)
	if err != nil {
		glog.Error(err)
		reasons = append(reasons, fmt.Sprintf("机审错误:%s", err.Error()))
	} else if pass {
		prize.State = 1
	}
	prize.Stype = 1
	prize.Stime = time.Now().Unix()
	prize.Reason = strings.Join(reasons, ", ")
	prize.Save()
	if prize.State == 1 {
		rolePid.Tell(&pb.ActivityTurnPrizeAuditSuccess{PrizeId: prize.Id})
	}

	// 缓存领奖记录
	prizeBytes, err := (&pb.ActivityTurnPrize{
		Id:        prize.Id,
		State:     prize.State,
		Prize:     prize.Prize,
		PrizeType: prize.PrizeType,
		Ctime:     time.Unix(prize.Ctime, 0).In(location).Format(utils.FORMAT),
		CtimeSec:  prize.Ctime,
	}).Marshal()
	if err != nil {
		glog.Error(err)
	} else {
		prizeB64 := base64.StdEncoding.EncodeToString(prizeBytes)
		if err = myredis.Redis().HSet(context.Background(), fmt.Sprintf("%s:%s", turntableUserPrizeKey, prize.Userid), prize.Id, prizeB64).Err(); err != nil {
			glog.Error(err)
		}
	}
}

// 转盘领奖审核通过发送奖励
func (a *RoleActor) ActivityTurnPrizeAuditSuccess(ctx actor.Context) {
	arg := ctx.Message().(*pb.ActivityTurnPrizeAuditSuccess)
	prize := &data.ActivityTurnPrizeLog{Id: arg.PrizeId}
	prize.GetById()

	glog.Infof("机审通过发送奖励: %s, %#v", arg.PrizeId, prize)
	if prize.State != 1 {
		return
	}

	// 审核已通过
	// 发送奖励 1代表bonus,2代表cash,3代表withdrawable
	var bonus, diamond, out int64
	switch prize.PrizeType {
	case 1:
		bonus = prize.Prize
	case 2:
		diamond = prize.Prize
	case 3:
		diamond = prize.Prize
		out = prize.Prize
	}
	if role, ok := a.roles[prize.Userid]; ok {
		glog.Infof("转盘活动奖励领取在线: user=%s, id=%s", prize.Userid, prize.Id)
		msg := &pb.ChangeCurrency{
			Userid:  prize.Userid,
			Type:    int32(pb.LOG_TYPE135),
			Diamond: diamond,
			Give:    bonus,
			Out:     out,
			Desc:    "转盘活动奖励领取",
			WaterId: prize.Id,
		}
		role.Pid.Tell(msg)
	} else {
		glog.Infof("转盘活动奖励领取离线: user=%s, id=%s", prize.Userid, prize.Id)
		a.syncCurrency(diamond, 0, bonus, out, 0, int32(pb.LOG_TYPE135), prize.Userid, "转盘活动奖励领取", prize.Id, false)
	}
}

// activityTurnTackPrizeAuditMachine 转盘领奖机审
func (a *RoleActor) activityTurnTackPrizeAuditMachine(prize *data.ActivityTurnPrizeLog, turnStime, turnEtime int64, activityTurnPrizeTimes int32) (pass bool, reasons []string, err error) {
	userid := prize.Userid

	if prize.PrizeType == 1 {
		reasons = append(reasons, "奖金类型bonus")
		// 发 bonus
		var isTurnInvitesPayed bool // 本轮邀请来的人是否至少有一人充值
		if isTurnInvitesPayed, err = activityTurnIsInvitesPayedLessOne(userid, true, turnStime, turnEtime); err != nil {
			return
		}
		if isTurnInvitesPayed {
			reasons = append(reasons, "本轮邀请用户有充值")
			pass = true
			return
		} else {
			reasons = append(reasons, "本轮邀请用户无充值")
			if activityTurnPrizeTimes == 1 {
				reasons = append(reasons, "第一次申请领奖")
				pass = true
				return
			} else {
				reasons = append(reasons, "非第一次申请领奖")
				if activityTurnPrizeTimes == 2 {
					reasons = append(reasons, "第二次申请领奖")
					var isTurnInviteIpNoRef bool // 本轮邀请来的人至少一人IP无关联账号
					if isTurnInviteIpNoRef, err = activityTurnInvitesIpNoRefLessOne(userid, turnStime, turnEtime); err != nil {
						return
					}
					if isTurnInviteIpNoRef {
						reasons = append(reasons, "本轮邀请来的人包含IP无关联账号")
						pass = true
						return
					} else {
						reasons = append(reasons, "本轮邀请来的人不包含IP无关联账号")
						return
					}
				} else {
					reasons = append(reasons, "非第二次申请领奖")
					var isInvitesPayed bool // 所有邀请来的人是否至少有一人充值
					if isInvitesPayed, err = activityTurnIsInvitesPayedLessOne(userid, false, 0, 0); err != nil {
						return
					}
					if !isInvitesPayed {
						reasons = append(reasons, "邀请的所有人无人充值")
						return
					} else {
						reasons = append(reasons, "邀请的所有人有人充值")
						var isTurnInviteIpNoRef bool // 本轮邀请来的人至少一人IP无关联账号
						if isTurnInviteIpNoRef, err = activityTurnInvitesIpNoRefLessOne(userid, turnStime, turnEtime); err != nil {
							return
						}
						if isTurnInviteIpNoRef {
							reasons = append(reasons, "本轮邀请来的人包含IP无关联账号")
							pass = true
							return
						} else {
							reasons = append(reasons, "本轮邀请来的人不包含IP无关联账号")
							return
						}
					}
				}
			}
		}
	} else {
		// 发 cash 或 withdraw
		var prizeT = "cash"
		if prize.PrizeType == 2 {
			prizeT = "withdrawalble"
		}
		reasons = append(reasons, fmt.Sprintf("奖金类型%s", prizeT))
		// 第一次申请领奖
		if activityTurnPrizeTimes == 1 {
			reasons = append(reasons, "第一次申请领奖")
			var isTurnInvitesPayed bool // 本轮邀请来的人是否至少有一人充值
			if isTurnInvitesPayed, err = activityTurnIsInvitesPayedLessOne(userid, true, turnStime, turnEtime); err != nil {
				return
			}
			if isTurnInvitesPayed {
				reasons = append(reasons, "本轮邀请用户有充值")
				pass = true
				return
			} else {
				reasons = append(reasons, "本轮邀请用户无充值")
				var isInvitesIpNoRef bool // 邀请来的人是否IP关联账号数都是0
				if isInvitesIpNoRef, err = activityTurnInvitesAllIpNoRef(userid); err != nil {
					return
				}
				if isInvitesIpNoRef {
					reasons = append(reasons, "邀请来的人IP关联账号数都是0")
					pass = true
					return
				} else {
					reasons = append(reasons, "邀请来的人有IP关联账号")
					return
				}
			}
		} else {
			reasons = append(reasons, "非第一次申请领奖")
			var isTurnInvitesPayed bool // 本轮邀请来的人是否至少有一人充值(未邀请到人也算无充值)
			if isTurnInvitesPayed, err = activityTurnIsInvitesPayedLessOne(userid, true, turnStime, turnEtime); err != nil {
				return
			}
			if !isTurnInvitesPayed {
				reasons = append(reasons, "本轮邀请用户无充值")
				return
			} else {
				reasons = append(reasons, "本轮邀请用户有充值")
				var isTurnInviteIpNoRef bool // 本轮邀请来的人IP关联账号是否都为0
				if isTurnInviteIpNoRef, err = activityTurnInvitesIpNoRefLessOne(userid, turnStime, turnEtime); err != nil {
					return
				}
				if isTurnInviteIpNoRef {
					reasons = append(reasons, "本轮邀请来的人包含IP无关联账号")
					pass = true
					return
				} else {
					reasons = append(reasons, "本轮邀请来的人不包含IP无关联账号")
					return
				}
			}
		}
	}
}

// activityTurnIsInvitesPayedLessOne 邀请来的人是否至少有一人充值
// thisTurn 是否只计算本轮邀请
func activityTurnIsInvitesPayedLessOne(userid string, thisTurn bool, turnStime, turnEtime int64) (payed bool, err error) {
	sql := `
		SELECT COUNT(*) pay_users FROM (
			SELECT userid FROM game.col_trade_record FINAL WHERE userid IN (
				SELECT userid FROM game.col_user FINAL WHERE share_superior = ? %s
			) AND order_status = 4
			GROUP BY userid
		) t1
	`
	args := []any{userid}
	var turnTimeF = ""
	if thisTurn {
		turnTimeF = " AND ctime BETWEEN ? AND ?"
		stime := time.Unix(turnStime, 0).In(location)
		etime := time.Unix(turnEtime, 0).In(location)
		args = append(args, stime)
		args = append(args, etime)
	}
	sql = fmt.Sprintf(sql, turnTimeF)

	var datas []map[string]any
	if err = ck.Select(&datas, sql, args...); err != nil {
		return
	}
	if len(datas) > 0 {
		payed = utils.ToInt64(datas[0]["pay_users"]) > 0
	}
	return
}

// 本轮邀请来的人至少一人IP无关联账号
func activityTurnInvitesIpNoRefLessOne(userid string, turnStime, turnEtime int64) (payed bool, err error) {
	sql := `
		SELECT count(*) no_ip_ref_count FROM (
			SELECT userid, regist_ip FROM game.col_user FINAL WHERE userid IN (
				SELECT userid FROM game.col_user FINAL WHERE share_superior = ? AND ctime BETWEEN ? AND ?
			)
		) t1 JOIN (
			SELECT regist_ip, count(*) ip_regist_count FROM game.col_user FINAL WHERE regist_ip IN (
				SELECT regist_ip FROM game.col_user FINAL WHERE share_superior = ? AND ctime BETWEEN ? AND ?
			)
			GROUP BY regist_ip
		) t2 ON t1.regist_ip = t2.regist_ip
		WHERE t2.ip_regist_count == 1
	`
	stime := time.Unix(turnStime, 0).In(location)
	etime := time.Unix(turnEtime, 0).In(location)

	var datas []map[string]any
	if err = ck.Select(&datas, sql, userid, stime, etime, userid, stime, etime); err != nil {
		return
	}
	if len(datas) > 0 {
		payed = utils.ToInt64(datas[0]["no_ip_ref_count"]) > 0
	}
	return
}

// 邀请来的人是否IP关联账号数都是0
func activityTurnInvitesAllIpNoRef(userid string) (payed bool, err error) {
	sql := `
		SELECT count(*) ip_ref_count FROM (
			SELECT userid, regist_ip FROM game.col_user FINAL WHERE userid IN (
				SELECT userid FROM game.col_user FINAL WHERE share_superior = ?
			)
		) t1 JOIN (
			SELECT regist_ip, count(*) ip_regist_count FROM game.col_user FINAL WHERE regist_ip IN (
				SELECT regist_ip FROM game.col_user FINAL WHERE share_superior = ?
			)
			GROUP BY regist_ip
		) t2 ON t1.regist_ip = t2.regist_ip
		WHERE t2.ip_regist_count > 1
	`

	var datas []map[string]any
	if err = ck.Select(&datas, sql, userid, userid); err != nil {
		return
	}
	if len(datas) > 0 {
		payed = utils.ToInt64(datas[0]["ip_ref_count"]) == 0
	}
	return
}

// 后台转盘活动审核
func (a *RoleActor) webTurnAudit(rsp *pb.WebResponse, op *data.ActivityTurnPrizeOperate) {
	prizes := data.ListActivityTurnPrizeLogByIds(op.PrizeIds)
	for _, prize := range prizes {
		if prize.State != 0 {
			continue
		}
		prize.Stype = 2
		prize.State = int32(op.Op)
		prize.Stime = time.Now().Unix()
		prize.AuditUser = op.UserName
		prize.Save()
		if prize.State == 1 {
			rolePid.Tell(&pb.ActivityTurnPrizeAuditSuccess{PrizeId: prize.Id})
		}

		// 缓存领奖记录
		prizeBytes, err := (&pb.ActivityTurnPrize{
			Id:        prize.Id,
			State:     prize.State,
			Prize:     prize.Prize,
			PrizeType: prize.PrizeType,
			Ctime:     time.Unix(prize.Ctime, 0).In(location).Format(utils.FORMAT),
			CtimeSec:  prize.Ctime,
		}).Marshal()
		if err != nil {
			glog.Error(err)
		} else {
			prizeB64 := base64.StdEncoding.EncodeToString(prizeBytes)
			if err = myredis.Redis().HSet(context.Background(), fmt.Sprintf("%s:%s", turntableUserPrizeKey, prize.Userid), prize.Id, prizeB64).Err(); err != nil {
				glog.Error(err)
			}
		}
	}
}

// ActivityTurnPrizesReq 转盘奖金审核记录
func (a *RoleActor) ActivityTurnPrizesReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.ActivityTurnPrizesReq)
	rsp := new(pb.ActivityTurnPrizesRsp)

	c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	prizes, err := myredis.Redis().HGetAll(c, fmt.Sprintf("%s:%s", turntableUserPrizeKey, arg.Userid)).Result()
	if err != nil {
		glog.Error(err)
		rsp.Error = pb.Failed
		ctx.Respond(rsp)
		return
	}
	for _, prizeB64 := range prizes {
		prize := new(pb.ActivityTurnPrize)
		prizeBytes, err := base64.StdEncoding.DecodeString(prizeB64)
		if err != nil {
			glog.Error(err)
			rsp.Error = pb.Failed
			ctx.Respond(rsp)
			return
		}
		if err = prize.Unmarshal(prizeBytes); err != nil {
			glog.Error(err)
			rsp.Error = pb.Failed
			ctx.Respond(rsp)
			return
		}
		rsp.Prizes = append(rsp.Prizes, prize)
	}
	// 申请时间倒序
	sort.Slice(rsp.Prizes, func(i, j int) bool {
		return rsp.Prizes[i].CtimeSec > rsp.Prizes[j].CtimeSec
	})
	ctx.Respond(rsp)
}
