package config

import (
	"sort"
	"strconv"
	"sync"
	"time"

	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/table"
	"goserver/pkg/utils"
)

// ActivityMap 活动列表
var ActivityMap *sync.Map

// 签到数据列表
var DailySignMap *sync.Map

// 在线奖励数据列表
var OnlineRewardMap *sync.Map

// 限时礼包数据列表
var LimitedGiftMap *sync.Map

// 周卡
var WeeklyCardMap *sync.Map

// 大富翁
var ScratchTicketConfigs []*data.ScratchTicketConfig

// 玩游戏分享抽奖
var PlayShareMap *sync.Map

// InitActivity 启动初始化
func InitActivity() {
	ActivityMap = new(sync.Map)
	WeeklyCardMap = new(sync.Map)
	PlayShareMap = new(sync.Map)
	l := data.GetActivityList()
	for _, v := range l {
		SetActivity(v)
	}
	w := data.GetWeeklyCardList()
	for _, v := range w {
		SetWeeklyCard(v)
	}
	p := data.GetPlayShareList()
	for _, v := range p {
		SetPlayShareConfig(v)
	}
	ScratchTicketConfigs = data.GetScratchTicketConfigList()
}

// InitActivity2 启动初始化
func InitActivity2() {
	ActivityMap = new(sync.Map)
	WeeklyCardMap = new(sync.Map)
	PlayShareMap = new(sync.Map)
	ScratchTicketConfigs = make([]*data.ScratchTicketConfig, 0)
}

// InitDailySign 启动初始化
func InitDailySign() {
	DailySignMap = new(sync.Map)
	l := data.GetDailySignList()
	for _, v := range l {
		SetDailySign(v)
	}
}

// InitDailySign2 启动初始化
func InitDailySign2() {
	DailySignMap = new(sync.Map)
}

// InitOnlineReward 启动初始化
func InitOnlineReward() {
	OnlineRewardMap = new(sync.Map)
	l := data.GetOnlineRewardList()
	for _, v := range l {
		SetOnlineReward(v)
	}
}

// InitOnlineReward2 启动初始化
func InitOnlineReward2() {
	OnlineRewardMap = new(sync.Map)
}

// GetActivitys2 同步数据时获取
func GetActivitys2() map[string]data.Activity {
	m := make(map[string]data.Activity)
	ActivityMap.Range(func(k, v interface{}) bool {
		m[k.(string)] = v.(data.Activity)
		return true
	})
	return m
}

// GetActivitys 客户端获取消息列表
func GetActivitys() []data.Activity {
	list := make([]data.Activity, 0)
	ActivityMap.Range(func(k, v interface{}) bool {
		if val, ok := v.(data.Activity); ok {
			if val.Del > 0 {
				return false
			}
			// if val.EndTime.Before(utils.BsonNow()) {
			// 	return false
			// }
			list = append(list, val)
		}
		return true
	})
	return list
}

// GetActivitys1 客户端获取消息列表
func GetActivitys1() []data.Activity {
	list := make([]data.Activity, 0)
	ActivityMap.Range(func(k, v interface{}) bool {
		if val, ok := v.(data.Activity); ok {
			if val.Del > 0 {
				return false
			}
			// if val.EndTime.Before(utils.BsonNow().AddDate(0, 0, 1)) {
			// 	return false //延长1天用于发货
			// }
			list = append(list, val)
		}
		return true
	})
	return list
}

// DelActivity 删除元素
func DelActivity(k interface{}) {
	ActivityMap.Delete(k)
}

// SetActivity 添加新的活动
func SetActivity(v data.Activity) {
	if v.Del > 0 {
		ActivityMap.Delete(v.Id)
	} else {
		ActivityMap.Store(v.Id, v)
	}
}

// GetActivity 获取指定活动
func GetActivity(id string) data.Activity {
	if v, ok := ActivityMap.Load(id); ok {
		//return v.(data.Activity)
		if val, ok := v.(data.Activity); ok {
			if val.Del > 0 {
				return data.Activity{}
			}
			// if val.EndTime.Before(utils.BsonNow()) {
			// 	return data.Activity{}
			// }
			return val
		}
	}
	return data.Activity{}
}

// GetActivity1 获取指定活动
func GetActivity1(id string) data.Activity {
	if v, ok := ActivityMap.Load(id); ok {
		//return v.(data.Activity)
		if val, ok := v.(data.Activity); ok {
			if val.Del > 0 {
				return data.Activity{}
			}
			// if val.EndTime.Before(utils.BsonNow().AddDate(0, 0, 1)) {
			// 	return data.Activity{}
			// }
			return val
		}
	}
	return data.Activity{}
}

// SetDailySign 添加新的签到数据
func SetDailySign(v data.DailySign) {
	DailySignMap.Store(v.Id, v)
}

// GetDailySign 获取指定签到数据
func GetDailySign(id int32) data.DailySign {
	if v, ok := DailySignMap.Load(id); ok {
		//return v.(data.Activity)
		if val, ok := v.(data.DailySign); ok {
			return val
		}
	}
	return data.DailySign{}
}

// DelDailySign 删除元素
func DelDailySign(k interface{}) {
	DailySignMap.Delete(k)
}

// GetDailySigns2 同步数据时获取
func GetDailySigns2() map[int32]data.DailySign {
	m := make(map[int32]data.DailySign)
	DailySignMap.Range(func(k, v interface{}) bool {
		m[k.(int32)] = v.(data.DailySign)
		return true
	})
	return m
}

// GetDailySigns3 同步数据时获取
func GetDailySigns3() []data.DailySign {
	m := make([]data.DailySign, 0)
	DailySignMap.Range(func(k, v interface{}) bool {
		if bean, ok := v.(data.DailySign); ok {
			m = append(m, bean)
		}
		return true
	})
	sort.Slice(m, func(i, j int) bool {
		return m[i].Id < m[j].Id
	})
	return m
}

// ----------------------在线奖励----------------------------------

// SetOnlineReward 添加新的签到数据
func SetOnlineReward(v data.OnlineReward) {
	OnlineRewardMap.Store(v.Id, v)
}

// GetOnlineReward 获取指定签到数据
func GetOnlineReward(id int32) data.OnlineReward {
	if v, ok := OnlineRewardMap.Load(id); ok {
		//return v.(data.Activity)
		if val, ok := v.(data.OnlineReward); ok {
			return val
		}
	}
	return data.OnlineReward{}
}

// DelOnlineReward 删除元素
func DelOnlineReward(k interface{}) {
	OnlineRewardMap.Delete(k)
}

// GetOnlineReward2 同步数据时获取
func GetOnlineReward2() map[int32]data.OnlineReward {
	m := make(map[int32]data.OnlineReward)
	OnlineRewardMap.Range(func(k, v interface{}) bool {
		m[k.(int32)] = v.(data.OnlineReward)
		return true
	})
	return m
}

// ----------------------限时礼包----------------------------------

// InitOnlineReward 启动初始化
func InitLimitedGift() {
	LimitedGiftMap = new(sync.Map)
	l := data.GetLimitedGiftList()
	for _, v := range l {
		SetLimitedGift(v)
	}
}

func InitLimitedGift2() {
	LimitedGiftMap = new(sync.Map)
}

// SetLimitedGift 添加新的签到数据
func SetLimitedGift(v data.LimitedGift) {
	LimitedGiftMap.Store(v.Id, v)
}

// LimitedGift 获取指定签到数据
// func GetLimitedGift(id int32) data.LimitedGift {
// 	if v, ok := LimitedGiftMap.Load(id); ok {
// 		//return v.(data.Activity)
// 		if val, ok := v.(data.LimitedGift); ok {
// 			return val
// 		}
// 	}
// 	return data.LimitedGift{}
// }

// DelOnlineReward 删除元素
func DelLimitedGift(k interface{}) {
	LimitedGiftMap.Delete(k)
}

// GetOnlineReward2 同步数据时获取
func GetLimitedGift2() map[int32]data.LimitedGift {
	m := make(map[int32]data.LimitedGift)
	LimitedGiftMap.Range(func(k, v interface{}) bool {
		m[k.(int32)] = v.(data.LimitedGift)
		return true
	})
	return m
}

// GetLimitedGiftList 同步数据时获取
func GetLimitedGiftList() []data.LimitedGift {
	m := make([]data.LimitedGift, 0)
	LimitedGiftMap.Range(func(k, v interface{}) bool {
		if g, ok := v.(data.LimitedGift); ok && g.Switch == 1 {
			m = append(m, g)
		}
		return true
	})
	sort.Slice(m, func(i, j int) bool {
		return m[i].Id < m[j].Id
	})
	return m
}

// ----------------------周卡----------------------------------
func SetWeeklyCard(v data.WeeklyCard) {
	WeeklyCardMap.Store(v.ID, v)
}

// func GetWeeklyCardMap() map[int32]data.WeeklyCard {
// 	m := make(map[int32]data.WeeklyCard)
// 	WeeklyCardMap.Range(func(k, v interface{}) bool {
// 		m[k.(int32)] = v.(data.WeeklyCard)
// 		return true
// 	})
// 	return m
// }

// 获取首冲活动
// func GetFirstRecharge(user *data.User) (res *pb.FirstRechargeRsp) {
// 	res = new(pb.FirstRechargeRsp)
// 	if user.FirstRecharge != "" {
// 		act := GetActivity(user.FirstRecharge)
// 		base := &pb.Activity{
// 			Id:        act.Id,
// 			Title:     act.Title,
// 			Content:   act.Content,
// 			StartTime: utils.Time2LocalStr(act.StartTime),
// 			// EndTime:   utils.Time2LocalStr(act.EndTime),
// 			Type: act.Type,
// 			Over: false,
// 		}
// 		data := &pb.FisrtRechargeData{
// 			Price:     int32(act.F_price),
// 			Number:    int32(act.F_number),
// 			FirstDay:  int32(act.F_firstDay),
// 			SecondDay: int32(act.F_secondDay),
// 			ThirdDay:  int32(act.F_thirdDay),
// 			ForthDay:  int32(act.F_forthDay),
// 			FifthDay:  int32(act.F_fifthDay),
// 		}
// 		userData := &pb.FisrtRecharge{
// 			Recharge: user.FirstRecharge != "",
// 			Receive:  user.Receive,
// 			Day:      user.ReceiveDay,
// 		}
// 		res.Base = base
// 		res.Data = data
// 		res.UserData = userData
// 		return
// 	}
// 	ActivityMap.Range(func(key, value any) bool {
// 		v := value.(data.Activity)
// 		if v.Del > 0 {
// 			// TODO
// 			return true
// 		}
// 		if v.Type == int32(pb.ACT_TYPE0) {
// 			base := &pb.Activity{
// 				Id:        v.Id,
// 				Title:     v.Title,
// 				Content:   v.Content,
// 				StartTime: utils.Time2LocalStr(v.StartTime),
// 				// EndTime:   utils.Time2LocalStr(v.EndTime),
// 				Type: v.Type,
// 				Over: false,
// 			}
// 			data := &pb.FisrtRechargeData{
// 				Price:     int32(v.F_price),
// 				Number:    int32(v.F_number),
// 				FirstDay:  int32(v.F_firstDay),
// 				SecondDay: int32(v.F_secondDay),
// 				ThirdDay:  int32(v.F_thirdDay),
// 				ForthDay:  int32(v.F_forthDay),
// 				FifthDay:  int32(v.F_fifthDay),
// 			}
// 			userData := &pb.FisrtRecharge{
// 				Recharge: user.FirstRecharge != "",
// 				Receive:  user.Receive,
// 				Day:      user.ReceiveDay,
// 			}
// 			res.Base = base
// 			res.Data = data
// 			res.UserData = userData
// 			res.Code = pb.OK
// 			return false
// 		}
// 		return true
// 	})
// 	return
// }

// 获取充值活动
func GetRechargeActivity(user *data.User) (res *pb.RechargeActivityRsp) {
	res = new(pb.RechargeActivityRsp)
	if user.NoviceGift != "" {
		// 重复参加
		res.Code = pb.ActRepeatJoin
		return
	}
	ActivityMap.Range(func(key, value any) bool {
		v := value.(data.Activity)
		if v.Del > 0 {
			// TODO
			return true
		}
		if v.Type == int32(pb.ACT_TYPE1) {
			base := &pb.Activity{
				Id:        v.Id,
				Title:     v.Title,
				Content:   v.Content,
				StartTime: utils.Time2LocalStr(v.StartTime),
				// EndTime:   utils.Time2LocalStr(v.EndTime),
				Type: v.Type,
				Over: false,
			}
			data := &pb.RechargeActivity{
				Id:             v.Id,
				GiveProportion: v.N_proportion,
				Number:         v.N_number,
				Price:          v.N_price,
			}
			res.Base = base
			res.Data = append(res.Data, data)
			res.Code = pb.OK
			return false
		}
		return true
	})
	return
}

// func GetWeeklyCardById(id int32) data.WeeklyCard {
// 	card := data.WeeklyCard{}
// 	if d, ok := WeeklyCardMap.Load(id); ok {
// 		if val, ok := d.(data.WeeklyCard); ok {
// 			return val
// 		}
// 	}
// 	return card
// }

// 获取周卡
func GetWeeklyCard(user *data.User, loc *time.Location) (res *pb.WeeklyCardRsp) {
	res = new(pb.WeeklyCardRsp)
	if user.WeeklyCardMap == nil {
		user.WeeklyCardMap = make(map[int32]*data.WeekCard)
	}
	tables := table.GetTables().GiftRechargeTable.GetDataList()
	for _, t := range tables {
		if t.GiftType != 3 {
			continue
		}
		data := BuildWeekCardData(t.Id, user, loc)
		res.Data = append(res.Data, data)
	}
	// WeeklyCardMap.Range(func(key, value any) bool {
	// 	if key == 0 {
	// 		return true
	// 	}
	// 	v := value.(data.WeeklyCard)
	// 	data := BuildWeekCardData(v, user)
	// 	res.Data = append(res.Data, data)
	// 	res.Code = pb.OK
	// 	return true
	// })
	return
}

// 获取每日签到数据
func GetDailSignData(user *data.User) (res *pb.DailySignDataRsp) {
	res = new(pb.DailySignDataRsp)
	if user.SignDay == 0 {
		user.SignDay = (1 << 0)
	}
	// lock := false
	list := GetDailySigns3()
	for _, bean := range list {
		data := &pb.DaiySignData{
			Day:    bean.Id,
			Ctype:  bean.Ctype,
			Number: bean.Number,
			Price:  bean.Price,
			State:  1,
		}
		// 判断能不能领
		if user.LoginTimes >= bean.Id {
			if (user.SignDay & (1 << bean.Id)) == 0 {
				// 还没领
				data.State = 2
			} else {
				data.State = 3
			}
		}
		// if user.LoginTimes >= bean.Id && !lock {
		// 	if (user.SignDay & (1 << bean.Id)) == 0 {
		// 		// 还没领
		// 		data.State = 2
		// 		if bean.Price > 0 {
		// 			// 这个没买后面的都不能领了
		// 			lock = true
		// 		} else {
		// 			lock = false
		// 		}
		// 	} else {
		// 		data.State = 3
		// 	}
		// }
		res.Data = append(res.Data, data)
	}
	// DailySignMap.Range(func(key, value any) bool {
	// 	v := value.(data.DailySign)
	// 	data := &pb.DaiySignData{
	// 		Day:    v.Id,
	// 		Ctype:  v.Ctype,
	// 		Number: v.Number,
	// 		Price:  v.Price,
	// 		State:  1,
	// 	}
	// 	// 判断能不能领
	// 	if user.LoginTimes >= v.Id {
	// 		if (user.SignDay & (1 << v.Id)) == 0 {
	// 			// 还没领
	// 			data.State = 2
	// 		} else {
	// 			data.State = 3
	// 		}
	// 	}
	// 	res.Data = append(res.Data, data)
	// 	res.Code = pb.OK
	// 	return true
	// })
	return
}

func BuildWeekCardData(id string, user *data.User, loc *time.Location) (bean *pb.WeeklyCardData) {
	idStr := id[len(id)-1:]
	cardId, _ := strconv.Atoi(idStr)
	t := table.GetTables().GiftRechargeTable.Get(id)
	var dailyGive, lastDayGive int32
	for _, v := range t.DailyGive[0].Nums {
		dailyGive += v
	}
	for _, v := range t.DailyGive[len(t.DailyGive)-1].Nums {
		lastDayGive += v
	}
	bean = &pb.WeeklyCardData{
		Id:            int32(cardId),
		Price:         t.Price,
		Number:        t.Price,
		DuringDay:     int32(len(t.DailyGive)),
		DailyReward:   dailyGive,
		LastDayReward: lastDayGive,
		State:         1,
	}
	if da, ok := user.WeeklyCardMap[int32(cardId)]; ok {
		if da.OverTime == -1 {
			return
		}
		if utils.LocalTime().Unix() >= da.OverTime {
			bean.State = 3
			// 下次可领取时间
			bean.NextTime = utils.TimestampTomorrow(loc)
			return
		}
		if da.Get {
			bean.State = 2
		} else {
			bean.State = 3
			// 下次可领取时间
			bean.NextTime = utils.TimestampTomorrow(loc)
		}
	}
	return
}

// 大富翁
func SetScratchTicketConfig(config []*data.ScratchTicketConfig) {
	ScratchTicketConfigs = config
}

func GetScratchTicketConfig() []data.ScratchTicketConfig {
	configs := make([]data.ScratchTicketConfig, 0)
	for _, s := range ScratchTicketConfigs {
		configs = append(configs, *s)
	}
	return configs
}

func GetScratchTicketConfigMap() map[int]data.ScratchTicketConfig {
	configs := make(map[int]data.ScratchTicketConfig)
	for _, s := range ScratchTicketConfigs {
		configs[s.Id] = *s
	}
	return configs
}

// 玩游戏分享抽奖
func SetPlayShareConfig(config data.PlayAndDrawActivity) {
	PlayShareMap.Store(config.Id, config)
}

func GetPlayShareConfig(id int) data.PlayAndDrawActivity {
	configs := data.PlayAndDrawActivity{}
	if v, ok := PlayShareMap.Load(id); ok {
		if val, ok := v.(data.PlayAndDrawActivity); ok {
			return val
		}
	}
	return configs
}

func GetPlayShareConfigMap() map[int]data.PlayAndDrawActivity {
	configs := make(map[int]data.PlayAndDrawActivity)
	PlayShareMap.Range(func(key, value any) bool {
		if val, ok := value.(data.PlayAndDrawActivity); ok {
			configs[val.Id] = val
		}
		return true
	})
	return configs
}
