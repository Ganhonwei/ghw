package handler

import (
	"sort"
	"time"

	"goserver/gen/pb"
	"goserver/gen/tb"
	"goserver/pkg/data"
	"goserver/pkg/game/config"
	"goserver/pkg/glog"
	"goserver/pkg/table"
	"goserver/pkg/utils"

	"github.com/globalsign/mgo/bson"
)

func packLogActivityMsg(v *data.LogActivity) (msg *pb.Activity) {
	msg = &pb.Activity{
		Id:       v.Actid,
		JoinTime: utils.Time2LocalStr(v.Jtime),
	}
	switch v.Type {
	case int32(pb.ACT_TYPE0):
		if v.Num >= 500000 {
			msg.Over = true
		}
	case int32(pb.ACT_TYPE1):
		//TODO 当天是否完成
	case int32(pb.ACT_TYPE2):
		if v.Num >= 15 {
			msg.Over = true
		}
	}
	return
}

func packActivityMsg(v data.Activity) (msg *pb.Activity) {
	msg = &pb.Activity{
		Id:        v.Id,
		Title:     v.Title,
		Content:   v.Content,
		StartTime: utils.Time2LocalStr(v.StartTime),
		// EndTime:   utils.Time2LocalStr(v.EndTime),
		Type: v.Type,
	}
	return
}

// SetActivityList 配置活动数据,测试数据
func SetActivityList() {
	// var startTime, endTime time.Time
	// startTime = utils.TimestampTodayTime()
	// endTime = startTime.AddDate(0, 0, 30)
	// NewActivity(int32(pb.ACT_TYPE0), "翻倍奖", "翻倍奖", startTime, endTime, 0, 0, 0)
	// NewActivity(int32(pb.ACT_TYPE1), "翻倍奖", "翻倍奖", startTime, endTime, 10000, 100, 100)
	// NewActivity(int32(pb.ACT_TYPE1), "翻倍奖", "翻倍奖", startTime, endTime, 50000, 500, 100)
	// NewActivity(int32(pb.ACT_TYPE1), "翻倍奖", "翻倍奖", startTime, endTime, 70000, 700, 80)
	// NewWeeklyCard(int32(pb.ACT_TYPE2), "翻倍奖", "翻倍奖", startTime, endTime, 70000, 700, 80, 3, 0)
	// NewWeeklyCard(int32(pb.ACT_TYPE2), "翻倍奖", "翻倍奖", startTime, endTime, 70000, 700, 80, 5, 1)
	// NewWeeklyCard(int32(pb.ACT_TYPE2), "翻倍奖", "翻倍奖", startTime, endTime, 70000, 700, 80, 7, 2)

	// endTime = startTime.AddDate(0, 0, 60)
	// NewActivity(int32(pb.ACT_TYPE1), "增加奖", "增加奖", startTime, endTime)
	// endTime = startTime.AddDate(0, 0, 90)
	// NewActivity(int32(pb.ACT_TYPE2), "激活奖", "激活奖", startTime, endTime)
}

// NewActivity 添加新活动
func NewActivity(Type int32, title, content string, startTime, endTime time.Time, price, number, proportion int32) {
	t := data.Activity{
		//Id:      bson.NewObjectId().String(),
		Id:        data.ObjectIdString(bson.NewObjectId()),
		Ctime:     bson.Now(),
		Type:      Type,
		Title:     title,
		Content:   content,
		StartTime: startTime,
		// EndTime:      endTime,
		F_price:      20000,
		F_number:     200,
		F_firstDay:   40,
		F_secondDay:  60,
		F_thirdDay:   80,
		F_forthDay:   80,
		F_fifthDay:   140,
		N_price:      price,
		N_number:     number,
		N_proportion: proportion,
	}
	config.SetActivity(t)
	t.Save()
}

func NewWeeklyCard(Type int32, title, content string, startTime, endTime time.Time, price, number, proportion, day, index int32) {
	t := data.Activity{
		//Id:      bson.NewObjectId().String(),
		Id:        data.ObjectIdString(bson.NewObjectId()),
		Ctime:     bson.Now(),
		Type:      Type,
		Title:     title,
		Content:   content,
		StartTime: startTime,
		// EndTime:      endTime,
		W_price:      price,
		W_number:     number,
		W_proportion: proportion,
		W_duringDay:  day,
		W_index:      index,
		W_daily:      10000,
	}
	config.SetActivity(t)
	t.Save()
}

func SetOnlineReward() {
	NewOnlineReward(1, 300, [][]int32{{1, 50, 1000}, {1, 100, 1000}, {1, 200, 1000}, {1, 300, 1000}, {1, 500, 3190}, {1, 800, 1500}, {1, 1000, 1000}, {1, 1500, 200}, {1, 2000, 100}, {1, 3000, 10}})
	NewOnlineReward(2, 900, [][]int32{{1, 50, 1000}, {1, 100, 1000}, {1, 200, 1000}, {1, 300, 1000}, {1, 500, 3190}, {1, 800, 1500}, {1, 1000, 1000}, {1, 1500, 200}, {1, 2000, 100}, {1, 3000, 10}})
	NewOnlineReward(3, 1800, [][]int32{{1, 50, 1000}, {1, 100, 1000}, {1, 200, 1000}, {1, 300, 1000}, {1, 500, 3190}, {1, 800, 1500}, {1, 1000, 1000}, {1, 1500, 200}, {1, 2000, 100}, {1, 3000, 10}})
	NewOnlineReward(4, 5400, [][]int32{{1, 50, 1000}, {1, 100, 1000}, {1, 200, 1000}, {1, 300, 1000}, {1, 500, 3190}, {1, 800, 1500}, {1, 1000, 1000}, {1, 1500, 200}, {1, 2000, 100}, {1, 3000, 10}})
	NewOnlineReward(5, 9000, [][]int32{{1, 50, 1000}, {1, 100, 1000}, {1, 200, 1000}, {1, 300, 1000}, {1, 500, 3190}, {1, 800, 1500}, {1, 1000, 1000}, {1, 1500, 200}, {1, 2000, 100}, {1, 3000, 10}})
}

func SetPokerHands() {
	NewPokerhands(1, 2, 1, [][]int32{{1, 300, 400}, {1, 500, 300}, {1, 1000, 200}, {1, 2000, 99}, {1, 3000, 1}})
	NewPokerhands(2, 3, 3, [][]int32{{1, 300, 200}, {1, 500, 300}, {1, 1000, 400}, {1, 2000, 99}, {1, 3000, 1}})
	NewPokerhands(3, 4, 5, [][]int32{{1, 300, 200}, {1, 500, 300}, {1, 1000, 400}, {1, 2000, 99}, {1, 3000, 1}})
	NewPokerhands(4, 5, 10, [][]int32{{1, 300, 1}, {1, 500, 1}, {1, 1000, 20}, {1, 2000, 60}, {1, 3000, 18}})
	NewPokerhands(5, -1, 15, [][]int32{{1, 300, 1}, {1, 500, 1}, {1, 1000, 20}, {1, 2000, 60}, {1, 3000, 18}})
}

func SetDailySign() {
	NewDailySign(1, int32(data.DIAMOND), 500)
	NewDailySign(2, int32(data.DIAMOND), 500)
	NewDailySign(3, int32(data.DIAMOND), 1000)
	NewDailySign(4, int32(data.DIAMOND), 500)
	NewDailySign(5, int32(data.DIAMOND), 1500)
}

func NewDailySign(day, ctype, number int32) {
	t := data.DailySign{
		//Id:      bson.NewObjectId().String(),
		Id:     day,
		Ctype:  ctype,
		Number: number,
	}
	config.SetDailySign(t)
	t.Save()
}

// 在线奖励
func NewOnlineReward(id int32, onlineTime int64, reward [][]int32) {
	t := data.OnlineReward{
		Id:         id,
		OnlineTime: onlineTime,
		Reward:     reward,
	}
	config.SetOnlineReward(t)
	t.Save()
}

// 牌型任务
func NewPokerhands(id, nextid, progress int32, reward [][]int32) {
	t := data.PokerHands{
		Id:       id,
		NextId:   nextid,
		Progress: progress,
		Rewards:  reward,
		Del:      0,
	}
	config.SetPhTask(t)
	t.Save()
}

// 在线奖励数据
func BuildOnlineRewadMsg() (beans []*pb.OnlineReward) {
	// onlineMap := config.GetOnlineReward2()
	onlineMap := table.GetTables().OnlineRewardTable.GetDataList()
	for _, v := range onlineMap {
		build := &pb.OnlineReward{
			Id:         v.Id,
			DuringTime: int32(v.OnlineTime),
		}
		for _, r := range v.Rewards {
			if len(r.Nums) < 2 {
				glog.Errorf("online reward length error id:%d", v.Id)
				continue
			}
			item := &pb.Item{
				Itype:  r.Nums[0],
				Number: int64(r.Nums[1]),
			}
			build.Items = append(build.Items, item)
		}
		beans = append(beans, build)
	}
	sort.Slice(beans, func(i, j int) bool {
		return beans[i].Id < beans[j].Id
	})
	return
}

// 入门礼包数据
func BuildRechargeActMsg(user *data.User) *pb.RechargeActivity {
	if user.NoviceGift != "" {
		// 重复参加
		return nil
	}
	activityMap := config.GetActivitys2()
	for _, v := range activityMap {
		if v.Del > 0 {
			// TODO
			continue
		}
		if v.Type == int32(pb.ACT_TYPE1) {
			bean := &pb.RechargeActivity{
				Id:             v.Id,
				GiveProportion: v.N_proportion,
				Number:         v.N_number,
				Price:          v.N_price,
			}
			return bean
		}
	}
	return nil
}

// 分享数据
func BuildShareDataMsg(user *data.User) (bean *pb.ShareData) {
	bean = new(pb.ShareData)
	bean.Referrals = int32(len(user.ShareBelow))
	if user.State != 2 {
		return
	}
	bean.TotalCash = user.ShareTotal
	bean.Withdrawable = user.ShareWithdraw
	for _, sd := range user.ShareBelow {
		// sharebean := table.GetTables().ShareConfigTable.Get()
		sharebean := config.GetShare(1)
		if sharebean.Id == 0 {
			break
		}
		if sd.Receive {
			continue
		}
		if sharebean.Recharge <= int32(sd.Recharge) {
			bean.ReRecharge += int64(sharebean.FirstRecharge)
		}
	}
	return
}

// 任务数据
func BuildTaskDataMsg(user *data.User) (beans []*pb.TaskData) {
	tasks := config.GetTasks()
	for _, t := range tasks {
		info := user.Task[t.ID]
		if info == nil {
			// 没有，新建任务
			info = CreateTask(t)
			user.Task[t.ID] = info
		}
		bean, err := BuildTaskData(info)
		if err != nil {
			continue
		}
		beans = append(beans, bean)
	}
	return
}

// vb任务数据
func BuildVBTaskDataMsg(user *data.User) (beans []*pb.VBTaskData) {
	for _, task := range user.VBTask {
		bean := &pb.VBTaskData{
			Id:         task.Id,
			RewardMode: task.RewardMode,
			Reward:     task.Reward,
			Prize:      task.Prize,
		}
		for _, typeId := range task.TaskTypeId {
			t := table.GetTables().VBGameTaskTypeTable.Get(typeId)
			if t == nil {
				glog.Errorf("vb task type %d not found", typeId)
				continue
			}
			taskType := &pb.VBTaskType{
				Id:          typeId,
				Name:        t.Name,
				MaxProgress: GetVBTaskTypeMaxProgress(t),
				Progress:    0,
				Gtype:       t.Gtype,
			}
			// 任务进度
			t0 := task.TaskTypes[typeId]
			if t0 != nil {
				taskType.Progress = t0.Progress
			}
			// 任务已完成，所有子任务进度拉满
			if task.Prize == 2 || task.Prize == 3 {
				taskType.Progress = taskType.MaxProgress
			}
			bean.TaskTypes = append(bean.TaskTypes, taskType)
		}
		beans = append(beans, bean)
	}
	return
}

// vb任务数据
func BuildVBLoseCompensationMsg(user *data.User) *pb.VBLoseCompensation {
	loseCom := table.GetTables().LoseCompensationTable.Get()
	nextReceive := (int64(loseCom.RewardInterval)*60 - (utils.BsonNow().Unix() - user.VBOutDiamondRewardTime))
	if nextReceive < 0 {
		nextReceive = 0
	}
	nextReceive = 0
	var canReward bool = user.VBOutDiamond >= int64(loseCom.Withdrawal) && user.VBOutDiamondDayTimes < loseCom.DayLimit && nextReceive <= 0
	// if loseCom.Mod == 2 {
	// 	// 释放vb金不足 领取时提示错误
	// 	canReward = canReward && user.VBBank >= int64(loseCom.Withdrawal)
	// }
	// loseCom.Withdrawal

	return &pb.VBLoseCompensation{
		VbOutDiamond: int32(user.VBOutDiamond),
		LoseRate:     loseCom.Compensate,
		Withdrawal:   loseCom.Withdrawal,
		CanReward:    canReward,
		NextReceive:  nextReceive, // 下次领取时间
	}
}

// 获取VB任务最大进度
func GetVBTaskTypeMaxProgress(record *tb.VbGameTaskTypeRecord) int32 {
	switch record.TaskType {
	case 1: // 局数任务
		return record.Arg1
	case 2: // 赢局数任务
		return record.Arg1
	case 101: // tp,ak47,joker特定牌型赢n局
		return record.Arg1
	case 102: // rummy在n回合获胜m局
		return record.Arg1
	case 103: // crash aviator在n倍以上逃脱n次
		return record.Arg1
	}

	if record.Arg1 > 0 {
		return record.Arg1
	}
	return 99999
}

// 牌型任务数据
func BuildPokerHandsDataMsg(user *data.User) *pb.PokerHandsTask {
	if user.PhRewardId == 0 {
		user.PhRewardId = 1
	}
	task := config.GetPhTask(user.PhRewardId)
	if task.Id == 0 {
		return nil
	}
	bean := new(pb.PokerHandsTask)
	bean.Progress = user.PhProgress
	bean.MaxProgress = task.Progress
	return bean
}

// 限时礼包数据
// func BuildLimitedGiftDataMsg(user *data.User) *pb.LimitedGift {
// 	if user.LimitedGiftId == 0 {
// 		return nil
// 	}
// 	if utils.LocalTime().UnixMilli() > user.LimitedGiftOverTime {
// 		user.OverlimitedGift = append(user.OverlimitedGift, user.LimitedGiftId)
// 		user.LimitedGiftId = 0
// 		return nil
// 	}
// 	b := config.GetLimitedGift(user.LimitedGiftId)
// 	bean := &pb.LimitedGift{
// 		Id:       b.Id,
// 		Price:    b.Price,
// 		Cash:     int32(b.Reward),
// 		OverTime: user.LimitedGiftOverTime,
// 	}
// 	return bean
// }

// 周卡数据
func BuildWeeklyCardDataMsg(user *data.User, loc *time.Location) (beans []*pb.WeeklyCardData) {
	if user.WeeklyCardMap == nil {
		user.WeeklyCardMap = make(map[int32]*data.WeekCard)
	}
	// weeklyCardMap := config.GetWeeklyCardMap()
	tables := table.GetTables().GiftRechargeTable.GetDataList()
	for _, card := range tables {
		if card.GiftType != 3 {
			continue
		}
		data := config.BuildWeekCardData(card.Id, user, loc)
		beans = append(beans, data)
	}
	return
}

// 签到数据
func BuildDailySignDataMsg(user *data.User) (beans []*pb.DaiySignData) {
	if user.SignDay == 0 {
		user.SignDay = (1 << 0)
	}
	signs := config.GetDailySigns2()
	for _, v := range signs {
		data := &pb.DaiySignData{
			Day:    v.Id,
			Ctype:  v.Ctype,
			Number: v.Number,
			Price:  v.Price,
			State:  1,
		}
		// 判断能不能领
		if user.LoginTimes >= v.Id {
			if (user.SignDay & (1 << v.Id)) == 0 {
				// 还没领
				data.State = 2
			} else {
				data.State = 3
			}
		}
		beans = append(beans, data)
	}
	return
}

// 商店罐子
func BuildShopPotDataMsg(user *data.User) (beans []*pb.ShopPot) {
	pots := user.ActivatedPot
	for _, p := range pots {
		bean := &pb.ShopPot{
			FlowWater: p.FlowWater,
			Number:    p.Number,
			StartTime: p.CTime,
			OverTime:  p.OverTime,
			Id:        p.Id,
		}
		beans = append(beans, bean)
	}
	return
}

// 玩游戏分享抽奖
func BuildPlayShareDataMsg(user *data.User) (bean *pb.PlayShareActivityData) {
	c := config.GetPlayShareConfig(1)
	if c.Id == 0 {
		glog.Error("no found config")
		return
	}

	// open := false
	// switchMap := config.GetSettingMap()
	// for _, sw := range switchMap {
	// 	if sw.Stype == 1 && sw.Name == "拼多多" && sw.Status == 1 {
	// 		open = true
	// 	}
	// }

	// 查看活动是否开启
	if !config.SettingIsOpenByName(1, "拼多多") {
		return nil
	}

	if user.PlayShareData.Config.Id == 0 {
		user.PlayShareData.Config = c
	}
	// 使用玩家身上的配置
	c = user.PlayShareData.Config

	data := user.PlayShareData
	if data.OverTime == 0 {
		// data.OverTime = utils.BsonNow().Unix() + int64(c.ValidityTime*60*60)
		return
	}

	if data.OverTime < utils.BsonNow().Unix() || data.Over {
		return
	}

	bean = &pb.PlayShareActivityData{
		Rounds:        int32(data.PlayTimes),
		InviteFriends: int32(len(data.InviteFriends)),
		DrawTimes:     int32(data.PlayDraws + data.ShareDraws),
		Rewards:       data.Rewards,
		OverTime:      data.OverTime,
		PlayRounds:    int32(c.PlayRounds[0]),
	}
	for _, v := range c.Reward {
		bean.ReceiveLine += int64(v)
	}
	if data.PlayUsed >= int(c.MaxPlayDraw) {
		bean.PopType = 1 // 切成分享
	}
	return
}
