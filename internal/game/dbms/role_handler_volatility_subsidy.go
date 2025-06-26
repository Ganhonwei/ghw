package dbms

import (
	"context"
	"fmt"
	"goserver/gen/pb"
	"goserver/gen/tb"
	"goserver/pkg/data"
	"goserver/pkg/data/ck"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/myactor"
	"goserver/pkg/myredis"
	"goserver/pkg/table"
	"goserver/pkg/utils"
	"sync"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"gopkg.in/mgo.v2/bson"
)

// 波动返水活动
const (
	volatilityBetsKey    = "activity:volatility:bets"    // 波动返水打码池
	volatilityRewardsKey = "activity:volatility:rewards" // 波动返水返奖池
	volatilitySubsidyKey = "activity:volatility:subsidy" // 波动返水今日已补贴
)

var (
	volatilityPoolUtime          int64 // 水池变动时间
	volatilityPoolLastRecordTime int64 // 水池记录时间
	volatilityPoolTimer          int
	volatilityPoolScore          = new(sync.Map)
)

func (a *RoleActor) volatilitySubsidyTick() {
	// 10s记录一次水池更新
	if volatilityPoolTimer < 4 {
		volatilityPoolTimer++
		return
	} else {
		volatilityPoolTimer = 0
	}

	// 记录水池
	if volatilityPoolLastRecordTime >= volatilityPoolUtime {
		return
	}

	// 更新水池到DB
	now := time.Now().In(location)
	poolTime := now
	if volatilityPoolUtime > 0 {
		poolTime = time.Unix(volatilityPoolUtime, 0).In(location)
	}
	today := poolTime.Format(utils.FORMAT_DATE)

	registAreas := []int32{0, 1, 2}
	for _, registArea := range registAreas {
		_, alreadySubsidy, bottomPool, fixedPool, dynamicPool := getVolatilitySubsidyPool(poolTime, registArea)
		recordVolatilitySubsidyPool(today, 1, registArea, bottomPool)
		recordVolatilitySubsidyPool(today, 2, registArea, fixedPool)
		recordVolatilitySubsidyPool(today, 3, registArea, dynamicPool)
		recordVolatilitySubsidyPool(today, 4, registArea, alreadySubsidy)
	}
	volatilityPoolLastRecordTime = now.Unix()
}

// 记录水池变动
// ptype: 1.保底水池,2.静态水池,3.动态水池,4.已补贴金额
func recordVolatilitySubsidyPool(today string, ptype, registArea int32, amount int64) {
	id := fmt.Sprintf("%s-%d-%d", today, ptype, registArea)
	if score, ok := volatilityPoolScore.Load(id); ok && score.(int64) == amount {
		// 与上个值相同不用修改
		return
	} else {
		volatilityPoolScore.Store(id, amount)
	}

	pool := &pb.VolatilitySubsidyPool{
		Id:         id,
		Date:       today,
		Ptype:      ptype,
		RegistArea: registArea,
		Amount:     amount,
		Utime:      time.Now().UnixMilli(),
	}
	myactor.Logger().Tell(pool)
}

// 获取波动返水3个水池最高值, 已补贴金额
// 保底水池、静态水池和动态水池。
// 1.保底水池金额是配置的一个固定金额。
// 1.静态水池投入比例*(该类玩家昨日总打码-该类玩家昨日总返奖)=该类玩家静态水池总额。
// 2.动态水池投入比例*(该类玩家今日总打码-该类玩家金日总返奖)=该类玩家动态水池总额。
func getVolatilitySubsidyPool(now time.Time, registArea int32) (maxPool, alreadySubsidy, bottomPool, fixedPool, dynamicPool int64) {
	ctx, cencel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cencel()

	today := now.Format(utils.FORMAT_DATE)
	yesterday := now.AddDate(0, 0, -1).Format(utils.FORMAT_DATE)

	var err error
	var todayBets, todayRewards, yesterdayBets, yesterdayRewards int64
	todayBetsKey := fmt.Sprintf("%s:%d:%s", volatilityBetsKey, registArea, today)
	todayRewardsKey := fmt.Sprintf("%s:%d:%s", volatilityRewardsKey, registArea, today)
	yesterdayBetsKey := fmt.Sprintf("%s:%d:%s", volatilityBetsKey, registArea, yesterday)
	yesterdayRewardsKey := fmt.Sprintf("%s:%d:%s", volatilityRewardsKey, registArea, yesterday)
	if todayBets, err = getVolatilitySubsidyPoolValue(ctx, todayBetsKey); err != nil {
		return
	}
	if todayRewards, err = getVolatilitySubsidyPoolValue(ctx, todayRewardsKey); err != nil {
		return
	}
	if yesterdayBets, err = getVolatilitySubsidyPoolValue(ctx, yesterdayBetsKey); err != nil {
		return
	}
	if yesterdayRewards, err = getVolatilitySubsidyPoolValue(ctx, yesterdayRewardsKey); err != nil {
		return
	}

	fixedPool = (yesterdayBets - yesterdayRewards) * int64(table.GetTables().VolatilitySubsidyTable.Get().FixedPoolRate[registArea]) / 10000 // 静态水池
	dynamicPool = (todayBets - todayRewards) * int64(table.GetTables().VolatilitySubsidyTable.Get().DynamicPoolRate[registArea]) / 10000     // 动态水池
	bottomPool = int64(table.GetTables().VolatilitySubsidyTable.Get().BottomPoolAmount[registArea])                                          // 保底水池

	// 今日已补贴金额
	subsidyKey := fmt.Sprintf("%s:%d:%s", volatilitySubsidyKey, registArea, today)
	if alreadySubsidy, err = getVolatilitySubsidyPoolValue(ctx, subsidyKey); err != nil {
		return
	}

	maxPool = utils.Max(bottomPool, utils.Max(fixedPool, dynamicPool))
	return
}

func getVolatilitySubsidyPoolValue(ctx context.Context, pollKey string) (score int64, err error) {
	score, err = myredis.Redis().IncrBy(ctx, pollKey, 0).Result()
	if err != nil {
		glog.Errorf("volatility subsidy get poll score error: %s, %v", pollKey, err)
		return
	}
	return
}

// 波动返水游戏打码逻辑
func volatilitySubsidyBet(userid string, registArea int32, bets int64) {
	ctx, cencel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cencel()

	now := time.Now().In(location)
	today := now.Format(utils.FORMAT_DATE)
	// 加到今日总打码池
	key := fmt.Sprintf("%s:%d:%s", volatilityBetsKey, registArea, today)
	if _, err := myredis.Redis().IncrBy(ctx, key, bets).Result(); err != nil {
		glog.Errorf("volatility subsidy incr bets error: %s, %d, %v", userid, bets, err)
		return
	}

	if err := myredis.Redis().Expire(ctx, key, time.Hour*24*7).Err(); err != nil {
		glog.Errorf("volatility subsidy bets key set expire error: %s, %d, %v", userid, bets, err)
	}

	volatilityPoolUtime = now.Unix()
}

// 波动返水游戏返奖逻辑
func volatilitySubsidyReward(userid string, registArea int32, amount int64) {
	ctx, cencel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cencel()

	now := time.Now().In(location)
	today := now.Format(utils.FORMAT_DATE)
	key := fmt.Sprintf("%s:%d:%s", volatilityRewardsKey, registArea, today)

	// 加到今日总返奖池
	if _, err := myredis.Redis().IncrBy(ctx, key, amount).Result(); err != nil {
		glog.Errorf("volatility subsidy incr reward error: %s, %d,%v", userid, amount, err)
		return
	}

	if err := myredis.Redis().Expire(ctx, key, time.Hour*24*7).Err(); err != nil {
		glog.Errorf("volatility subsidy bets key set expire error: %s, %d, %s, %v", userid, amount, key, err)
	}

	volatilityPoolUtime = now.Unix()
}

// 获取玩家波动返水配置
func getVolatilitySubsidyRecord(registArea int, money uint32) (subsidyRecord *tb.VolatilityVolatilitySubsidyUserRecord) {
	for _, record := range table.GetTables().VolatilitySubsidyUserTable.GetDataList() {
		payRange := record.PayRange[registArea].Value
		money := int32(money)
		if money >= payRange[0] && money <= payRange[1] {
			subsidyRecord = record
			return
		}
	}
	return
}

// 检查是否触发波动补贴, 计算补贴额
// trigger是否触发, subsidy补贴额
func volatilitySubsidyCheck(role *data.User) (trigger bool, subsidy, bets int64) {
	userid := role.Userid

	// 补贴开关
	subsidySwitch := table.GetTables().VolatilitySubsidyTable.Get().SubsidySwitch[role.RegistArea]
	if subsidySwitch != 1 {
		return
	}

	// 补贴次数上限
	if role.VolatilitySubsidyTimes >= table.GetTables().VolatilitySubsidyTable.Get().SubsidyTimes[role.RegistArea] {
		return
	}

	subsidyRecord := getVolatilitySubsidyRecord(role.RegistArea, role.Money)
	if subsidyRecord == nil {
		return
	}

	//（充-提-钱包总余额）/充 ≥ 触发补贴的亏损比例
	lossMoney := (int64(role.Money) - int64(role.CashOut) - role.GetScore())
	lossRate := lossMoney * 10000 / int64(role.Money)
	if lossRate < int64(subsidyRecord.LossRate[role.RegistArea]) {
		return
	}

	// 计算补贴额 当累计补贴给玩家的金额已经超过“个人补贴金额总额上限”的时候，不再给该玩家补贴
	subsidyRate := subsidyRecord.SubsidyRate[role.RegistArea]
	subsidy = lossMoney * int64(subsidyRate) / 10000
	// 个人补贴总额上限
	subsidyLimitRate := table.GetTables().VolatilitySubsidyTable.Get().SubsidyLimitRate[role.RegistArea]
	subsidyLimit := lossMoney * int64(subsidyLimitRate) / 10000

	// 已累计补给该玩家的金额
	userSubsidy := role.VolatilitySubsidyAmounts
	if subsidy+userSubsidy > subsidyLimit {
		subsidy = subsidyLimit - userSubsidy
	}
	if subsidy <= 0 {
		return
	}

	// 当总补贴出去的金额超过三个水池中的最高值时，对应的所有人不再补贴
	now := time.Now().In(location)
	pool, alreadySubsidy, _, _, _ := getVolatilitySubsidyPool(now, int32(role.RegistArea))
	if pool <= alreadySubsidy {
		return
	}

	// 查询打码量
	err := ck.Select(&bets, `
		SELECT SUM(bet_amount) bet_amount FROM (
			SELECT SUM(bet_amount) bet_amount
			FROM game.col_detail t1 FINAL WHERE userid = ? AND begin_time > ? AND robot = 0 and win_type in (1,2,3) 
				UNION ALL
			SELECT SUM(amount) bet_amount
			FROM game.col_nsq_log_external_bet t2 FINAL WHERE user_id = ? AND ctime > ? AND amount != 0 
		) s1
	`, userid, role.Ctime.Unix(), userid, role.Ctime.Unix())
	if err != nil {
		glog.Error("select user bets error:", err)
		return
	}
	// 当充投比≤个人充投比上限的时候才给补贴
	if float64(bets)/float64(role.Money) > table.GetTables().VolatilitySubsidyTable.Get().BetPayRateLimit[role.RegistArea] {
		return
	}

	trigger = true
	return
}

// 检查触发波动补贴没
func (a *RoleActor) VolatilitySubsidyCheck(ctx actor.Context) {
	arg := ctx.Message().(*pb.VolatilitySubsidyCheck)
	rsp := new(pb.VolatilitySubsidyChecked)
	rsp.Userid = arg.Userid
	defer ctx.Respond(rsp)

	role := a.getUser(arg.Userid)
	if role == nil {
		return
	}
	trigger, subsidy, bets := volatilitySubsidyCheck(role)
	rsp.Trigger = trigger
	rsp.Subsidy = subsidy
	rsp.Bets = bets
}

// 玩家回到大厅, 检查触发波动补贴没
func (a *RoleActor) VolatilitySubsidyToLobby(ctx actor.Context) {
	arg := ctx.Message().(*pb.VolatilitySubsidyToLobby)
	role := a.getUser(arg.Userid)
	if role == nil {
		return
	}
	trigger, _, _ := volatilitySubsidyCheck(role)
	if trigger {
		// 派奖弹窗 ntf 通知
		if role, ok := a.roles[arg.Userid]; ok {
			role.Pid.Tell(&pb.VolatilitySubsidyNtf{WheelTimes: 1})
		}
	}
}

// 波动返水转盘请求
func (a *RoleActor) VolatilitySubsidyWheelReq(ctx actor.Context) {
	arg := ctx.Message().(*pb.VolatilitySubsidyWheelReq)
	rsp := new(pb.VolatilitySubsidyWheelRsp)
	defer ctx.Respond(rsp)
	rsp.WheelTimes = 0

	role := a.getUserById(arg.Userid)
	if role == nil {
		rsp.Error = pb.Failed
		return
	}

	subsidyRecord := getVolatilitySubsidyRecord(role.RegistArea, role.Money)
	if subsidyRecord == nil {
		rsp.Error = pb.Failed
		return
	}

	trigger, subsidy, bets := volatilitySubsidyCheck(role)
	if !trigger {
		rsp.Error = pb.Failed
		return
	}

	// 加到今日已补贴
	c, cencel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cencel()
	now := time.Now().In(location)
	today := now.Format(utils.FORMAT_DATE)
	subsidyKey := fmt.Sprintf("%s:%d:%s", volatilitySubsidyKey, role.RegistArea, today)
	alreadySubsidy, err := myredis.Redis().IncrBy(c, subsidyKey, subsidy).Result()
	if err != nil {
		glog.Errorf("volatility subsidy incr already subsidy poll error: %s, %v", subsidyKey, err)
		rsp.Error = pb.Failed
		return
	}
	glog.Info("volatility subsidy user: %s, %d, already=%d", arg.Userid, subsidy, alreadySubsidy)
	if err := myredis.Redis().Expire(c, subsidyKey, time.Hour*24*7).Err(); err != nil {
		glog.Errorf("volatility subsidy already subsidy poll set expire error: %s, %v", subsidyKey, err)
	}

	volatilityPoolUtime = now.Unix()

	// 玩家已补贴次数和金额
	role.VolatilitySubsidyTimes++
	role.VolatilitySubsidyAmounts += subsidy
	if role.VolatilitySubsidyTimes == 1 {
		role.VolatilitySubsidyFirstTime = time.Now().UnixMilli()
	}

	subsidyType := subsidyRecord.SubsidyType[role.RegistArea]
	// 发送奖励 1代表bonus,2代表cash,3代表withdrawable
	var bonus, diamond, outDiamond int64
	switch subsidyType {
	case 1:
		bonus = subsidy
	case 2:
		diamond = subsidy
	case 3:
		diamond = subsidy
		outDiamond = subsidy
	}

	msg := handler.ChangeCurrencyMsg(diamond, 0, 0, 0, bonus, outDiamond,
		int32(pb.LOG_TYPE145), arg.Userid, "波动返水补贴领取", "")
	// 在线时
	if v, ok := a.roles[arg.Userid]; ok {
		// 同步gate
		v.Pid.Tell(&pb.VolatilitySubsidySync{
			VolatilitySubsidyTimes:     role.VolatilitySubsidyTimes,
			VolatilitySubsidyAmounts:   role.VolatilitySubsidyAmounts,
			VolatilitySubsidyFirstTime: role.VolatilitySubsidyFirstTime,
		})
		v.Pid.Tell(msg)
	} else {
		// 离线时
		rolePid.Tell(msg)
	}

	rsp.SubsidyAmount = subsidy

	// 补贴记录
	log := &data.VolatilitySubsidy{
		Id:          bson.NewObjectId().Hex(),
		Userid:      arg.Userid,
		First:       role.VolatilitySubsidyTimes == 1,
		FirstTime:   role.VolatilitySubsidyFirstTime,
		Sdate:       today,
		Subsidy:     subsidy,
		SubsidyType: subsidyType,
		Ctime:       time.Now().UnixMilli(),
		RegistArea:  int32(role.RegistArea),
		Pays:        int64(role.Money),
		Bets:        bets,
	}
	log.Save()
}
