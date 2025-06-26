package service

import (
	"errors"
	"fmt"

	"goserver/internal/web/admin/app/entity"
	"goserver/pkg/data/mq"

	"github.com/astaxie/beego"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type gameService struct{}

// 获取房间列表
func (this *gameService) GetGameList(page, pageSize int, m bson.M) ([]entity.Game, error) {
	var list []entity.Game
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "ctime", false)
	err := Games.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)

	//转换为分展示
	list = this.chipList2(list)
	return list, err
}

// 获取房间列表总数
func (this *gameService) GetGameListTotal(m bson.M) (int64, error) {
	return int64(Count(Games, m)), nil
}

// 根据ID获取房间信息
func (this *gameService) GetByGame(ids []string) ([]entity.Game, error) {
	var list []entity.Game
	q := bson.M{"_id": bson.M{"$in": ids}}
	err := Games.
		Find(q).All(&list)
	list = this.chipList2(list)
	return list, err
}

// 转换为分展示
func (this *gameService) chipList2(list []entity.Game) []entity.Game {
	for k, v := range list {
		v.FMin_Access = Chip2Float(int64(v.Min_Access))
		v.FMax_Access = Chip2Float(int64(v.Max_Access))
		v.FStock_Expect = Chip2Float(int64(v.Stock_Expect))
		v.FStock_Alarm = Chip2Float(int64(v.Stock_Alarm))
		c, _ := ConvertToIndiaTime(v.Ctime.Unix())
		v.Ctime = c
		list[k] = v
	}
	return list
}

// 添加房间列表信息
func (this *gameService) AddGame(shop *entity.Game) error {
	//shop.Id = utils.String(utils.Timestamp())
	id_str, err := this.GenerateID("last_game_id")
	if err != nil {
		return err
	}
	shop.Id = id_str
	shop.Ctime = bson.Now()
	//shop.Chip *= 100 //转换为分
	if !Insert(Games, shop) {
		return errors.New("写入失败:" + shop.Id)
	}
	return nil
}

// 根据ID获取房间列表信息
func (this *gameService) GetGame(id string) (entity.Game, error) {
	shop := entity.Game{}
	Get(Games, id, &shop)
	if shop.Id == "" {
		return shop, errors.New("不存在")
	}
	return shop, nil
}

// 移除房间列表
func (this *gameService) DelGame(id string) error {
	if Delete(Games, bson.M{"_id": id}) {
		return nil
	}
	return errors.New("移除失败")
}

// 更新房间列表
func (this *gameService) UpdateGame(shop *entity.Game) error {
	m := bson.M{
		"name":         shop.Name,
		"gtype":        shop.Gtype,
		"status":       shop.Status,
		"ai_status":    shop.Ai_Status,
		"min_access":   shop.Min_Access,
		"max_access":   shop.Max_Access,
		"stock_expect": shop.Stock_Expect,
		"stock_alarm":  shop.Stock_Alarm,
		"sort_id":      shop.SortId,
		"count":        shop.Count,
		"bottom":       shop.Bottom,
		"rounds":       shop.Rounds,
		"than_rounds":  shop.Than_Rounds,
		"pool_limit":   shop.Pool_Limit,
		"otime":        shop.Otime,
		"tcountdown":   shop.Tcountdown,
		"scountdown":   shop.Scountdown,
		"match_time":   shop.Match_Time,
		"single_robot": shop.Single_Robot,
		"robot_join":   shop.Robot_Join,
		"robot_leave":  shop.Robot_Leave,
		"prevent_time": shop.Prevent_Time,
		"prevent_num":  shop.Prevent_Num,
		"prevent_thaw": shop.Prevent_Thaw,
		"msg_score":    shop.Msg_Score,
		"ctime":        bson.Now(),
	}
	if Update(Games, bson.M{"_id": shop.Id}, bson.M{"$set": m}) {
		return nil
	}
	return errors.New("更新失败")
}

// 获取游戏系统设置列表
func (this *gameService) GetSystemList(page, pageSize int, m bson.M) ([]entity.GameSystem, error) {
	var list []entity.GameSystem
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "sort_id", true)
	err := GameSystems.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)

	//转换为分展示
	list = this.chipList4(list)
	return list, err
}

// 转换为分展示
func (this *gameService) chipList4(list []entity.GameSystem) []entity.GameSystem {
	for k, v := range list {
		c, _ := ConvertToIndiaTime(v.Ctime.Unix())
		v.Ctime = c
		list[k] = v
	}
	return list
}

// 获取游戏系统设置列表总数
func (this *gameService) GetSystemListTotal(m bson.M) (int64, error) {
	return int64(Count(GameSystems, m)), nil
}

// 根据ID获取房间列表信息
func (this *gameService) GetGameSystem(id int) ([]entity.GameSystem, error) {
	var list []entity.GameSystem
	ListByQ(GameSystems, bson.M{"stype": id}, &list)
	return list, nil
}

func (this *gameService) GetGameSystemByM(m bson.M) ([]entity.GameSystem, error) {
	var list []entity.GameSystem
	ListByQ(GameSystems, m, &list)
	return list, nil
}

// 添加或修改
func (this *gameService) AddOrUpdateGameSystem(info *entity.GameSystem) (err error) {
	defer func() {
		if err == nil {
			if info.Stype == 4 && (info.Rtype == 6 || info.Rtype == 7) {
				err = mq.NatsPublish(mq.TopicWithdrawAutoTransfer, mq.NatsNilData)
				if err != nil {
					beego.Error(fmt.Sprintf("%s publish err: ", mq.TopicWithdrawAutoTransfer), err)
				}
			}
		}
	}()

	if info.Id == "" {
		info.Id = bson.NewObjectId().Hex()
		info.Ctime = bson.Now()
		if !Insert(GameSystems, info) {
			return errors.New("写入失败:" + info.Id)
		}
		return nil
	} else {
		m := bson.M{
			"name":       info.Name,
			"gtype":      info.Gtype,
			"rtype":      info.Rtype,
			"stype":      info.Stype,
			"status":     info.Status,
			"pay_status": info.PayStatus,
			"sort_id":    info.SortId,
			"ctime":      bson.Now(),
			"value":      info.Value,
			"tab":        info.Tab,
		}
		if Update(GameSystems, bson.M{"_id": info.Id}, bson.M{"$set": m}) {
			return nil
		}
		return errors.New("更新失败")
	}
}

// 获取支付渠道设置列表
func (this *gameService) GetPayChannelList(page, pageSize int, m bson.M) ([]entity.PayChannel, error) {
	var list []entity.PayChannel
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "sort_id", true)
	err := PayChannels.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)
	list = this.chipList5(list)
	return list, err
}

// 转换为分展示
func (this *gameService) chipList5(list []entity.PayChannel) []entity.PayChannel {
	for k, v := range list {
		c, _ := ConvertToIndiaTime(v.Ctime.Unix())
		v.Ctime = c
		v.FBalance = fmt.Sprintf("%.2f", Chip2Float(int64(v.Balance)))
		// v.FWithdrawFee = ComputeFloat(v.WithdrawFee, 100)
		list[k] = v
	}
	return list
}

// 获取支付渠道设置列表总数
func (this *gameService) GetPayChannelTotal(m bson.M) (int64, error) {
	return int64(Count(PayChannels, m)), nil
}

// 根据ID获取支付渠道信息
func (this *gameService) GetPayChannel() ([]entity.PayChannel, error) {
	var list []entity.PayChannel
	ListByQ(PayChannels, nil, &list)
	return list, nil
}

// 获取所有支付渠道信息Map
func (this *gameService) GetPayChannelsMap() (payChannelsMap map[string]entity.PayChannel, err error) {
	payChannelsMap = make(map[string]entity.PayChannel)
	var payChannels []entity.PayChannel
	err = PayChannels.
		Find(bson.M{}).
		All(&payChannels)
	if err != nil {
		return
	}
	for _, c := range payChannels {
		payChannelMapping(&c)
		payChannelsMap[c.Id] = c
	}
	return
}

func (this *gameService) PayChannelAll() ([]*entity.PackageInfo, error) {
	var userlist []entity.PayChannel
	list := make([]*entity.PackageInfo, 0)
	m := bson.M{"robot": false, "bundle_id": bson.M{"$ne": ""}}
	err := PayChannels.
		Find(m).
		All(&userlist)
	if len(userlist) > 0 {
		result := make(map[string]bool)
		for _, v := range userlist {
			id := v.Id
			result[id] = true
		}

		var ids []string
		for key := range result {
			ids = append(ids, key)
		}
		for _, item := range ids {
			temp := new(entity.PackageInfo)
			temp.Key = item
			temp.Name = item
			list = append(list, temp)
		}
		newSlice := []*entity.PackageInfo{
			{Key: "0", Name: "全部"},
		}
		newSlice = append(newSlice, list...)
		info := new(entity.PackageInfo)
		info.Key = "-"
		info.Name = "其他"
		list = append(newSlice, info)
	}
	return list, err
}

// 添加或修改支付渠道
func (this *gameService) AddOrUpdatePayChannel(res *entity.PayChannel) error {
	info := new(entity.PayChannel)
	GetByQ(PayChannels, bson.M{"_id": bson.M{"$eq": res.Id}}, info)
	if info.Id == "" {
		res.Ctime = bson.Now()
		if !Insert(PayChannels, res) {
			return errors.New("写入失败:" + res.Id)
		}
		return nil
	} else {
		m := bson.M{
			"name":          res.Name,
			"status":        res.Status,
			"wstatus":       res.Wstatus,
			"sort_id":       res.SortId,
			"ptype":         res.Ptype,
			"pay_rate":      res.PayRate,
			"withdraw_rate": res.WithdrawRate,
			"withdraw_fee":  res.WithdrawFee,
			"balance":       res.Balance,
			"ctime":         bson.Now(),
		}
		if Update(PayChannels, bson.M{"_id": res.Id}, bson.M{"$set": m}) {
			return nil
		}
		return errors.New("更新失败")
	}
}

// 生成ID
func (this *gameService) GenerateID(idName string) (string, error) {
	gen := new(entity.ShopIDGen)
	gen.Id = idName

	// 使用findAndModify原子操作
	change := mgo.Change{
		Update: bson.M{
			"$setOnInsert": bson.M{"last_user_id": "m000001"},
			"$inc":         bson.M{"last_user_id": 1},
		},
		Upsert:    true,
		ReturnNew: true,
	}

	_, err := GenIDs.Find(bson.M{"_id": idName}).Apply(change, gen)
	if err != nil {
		return "", err
	}

	// 格式化ID
	if gen.LastUserId == "m000001" {
		return gen.LastUserId, nil
	}

	// 提取数字部分并格式化
	numStr := fmt.Sprintf("%06d", gen.LastUserId)
	return "m" + numStr, nil
}

/*库存*/
// 获取库存列表
func (this *gameService) GetStockList(page, pageSize int, m bson.M) ([]entity.Stock, error) {
	var list []entity.Stock
	if pageSize == -1 {
		pageSize = 100000
	}
	skipNum, sortFieldR := parsePageAndSort(page, pageSize, "_id", true)
	err := Stocks.
		Find(m).
		Sort(sortFieldR).
		Skip(skipNum).
		Limit(pageSize).
		All(&list)

	//转换为分展示
	list = this.chipList1(list)
	return list, err
}

// 获取库存列表总数
func (this *gameService) GetStockTotal(m bson.M) (int64, error) {
	return int64(Count(Stocks, m)), nil
}

// 根据ID获取库存信息
func (this *gameService) GetStockInfo(id string) (entity.Stock, error) {
	info := entity.Stock{}
	Get(Stocks, id, &info)
	if info.Id == "" {
		return info, errors.New("不存在")
	}
	return this.chipList3(&info), nil
}
func (this *gameService) GetRoomDropList(m bson.M) ([]entity.Game, error) {
	var list []entity.Game
	var gameids []string
	err := Games.
		Find(m).Distinct("_id", &gameids)
	n := bson.M{}
	n["_id"] = bson.M{"$in": gameids}
	var s_ids []string
	err = Stocks.Find(n).Distinct("_id", &s_ids)
	m["_id"] = bson.M{"$in": s_ids}
	err = Games.Find(m).All(&list)
	return list, err
}

// 转换为分展示
func (this *gameService) chipList1(list []entity.Stock) []entity.Stock {
	for k, v := range list {
		v.FCashStock = Chip2Float(int64(v.CashStock))
		v.FBonusStock = Chip2Float(int64(v.BonusStock))
		v.FCashMingTax = Chip2Float(int64(v.CashMingTax))
		v.FBonusMingTax = Chip2Float(int64(v.BonusMingTax))
		v.FCashAnTax = Chip2Float(int64(v.CashAnTax))
		v.FBonusAnTax = Chip2Float(int64(v.BonusAnTax))
		if v.CashStock < 0 {
			v.IsCashStock = true
		} else {
			v.IsCashStock = false
		}
		list[k] = v
	}
	return list
}

// 转换为分展示
func (this *gameService) chipList3(info *entity.Stock) entity.Stock {
	info.FCashMingTax = Chip2Float(int64(info.CashMingTax))
	info.FBonusMingTax = Chip2Float(int64(info.BonusMingTax))
	info.FCashAnTax = Chip2Float(int64(info.CashAnTax))
	info.FBonusAnTax = Chip2Float(int64(info.BonusAnTax))
	info.FCashStock = Chip2Float(info.CashStock)
	info.FBonusStock = Chip2Float(info.BonusStock)
	if len(info.History) > 0 {
		for k, v := range info.History {
			// v.SDate = time.Unix(v.Timestamp, 0)
			v.FCashStock = Chip2Float(int64(v.CashStock))
			v.FBonusStock = Chip2Float(int64(v.BonusStock))
			info.History[k] = v
		}
	}

	return *info
}

// 根据ID获取库存信息
func (this *gameService) GetGameStockInfo(id string) (entity.Stock, error) {
	info := entity.Stock{}
	Get(GameStocks, id, &info)
	if info.Id == "" {
		return info, errors.New("不存在")
	}
	return this.chipList3(&info), nil
}

// 龙虎斗 新地址库存
func (this *gameService) GetGameStock(m bson.M) ([]entity.GameStock, error) {
	var list []entity.GameStock
	err := GameStocks.
		Find(m).
		All(&list)
	//转换为分展示
	list = this.chipList6(list)
	return list, err
}

func (this *gameService) chipList6(list []entity.GameStock) []entity.GameStock {
	for k, info := range list {
		info.FCashMingTax = Chip2Float(int64(info.CashMingTax))
		info.FBonusMingTax = Chip2Float(int64(info.BonusMingTax))
		info.FCashAnTax = Chip2Float(int64(info.CashAnTax))
		info.FBonusAnTax = Chip2Float(int64(info.BonusAnTax))
		info.FCashStock = Chip2Float(info.CashStock)
		info.FBonusStock = Chip2Float(info.BonusStock)
		if info.CashStock < 0 {
			info.IsCashStock = true
		} else {
			info.IsCashStock = false
		}
		list[k] = info
	}
	return list
}
