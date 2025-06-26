package handler

import (
	"fmt"
	"math"
	"strconv"

	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/config"
	"goserver/pkg/glog"
	"goserver/pkg/table"
	"goserver/pkg/utils"

	"github.com/globalsign/mgo/bson"
	jsoniter "github.com/json-iterator/go"
)

// Desk2Data 打包桌子数据
func Desk2Data(deskData *data.DeskData) []byte {
	result, err := jsoniter.Marshal(deskData)
	if err != nil {
		glog.Errorf("Desk2Data Marshal err %v", err)
		return []byte{}
	}
	return result
}

// Data2Desk 解析桌子数据
func Data2Desk(deskDataStr []byte) *data.DeskData {
	deskData := new(data.DeskData)
	err := jsoniter.Unmarshal(deskDataStr, deskData)
	if err != nil {
		glog.Errorf("Data2Desk Unmarshal err %v", err)
		return nil
	}
	return deskData
}

// NewDeskData 转换为桌子数据
func NewDeskData(d *data.Game) *data.DeskData {
	return &data.DeskData{
		Unique:   d.Unique,
		DeskType: d.DeskType,
		Gtype:    d.Gtype,
		Rtype:    d.Rtype,
		Dtype:    d.Dtype,
		Ltype:    d.Ltype,
		Rname:    d.Name,
		Count:    d.Count,
		Ante:     d.Ante,
		Cost:     d.Cost,
		Vip:      d.Vip,
		Chip:     d.Chip,
		Deal:     d.Deal,
		Carry:    d.Carry,
		Down:     d.Down,
		Top:      d.Top,
		Sit:      d.Sit,
		Minimum:  d.Minimum,
		Maximum:  d.Maximum,
		Pub:      d.Pub,
		Mode:     d.Mode,
		Multiple: d.Multiple,
		Game:     *d,
		//Ctime:    uint32(utils.Time2Stamp(d.Ctime)),
		Ctime: uint32(utils.Timestamp()), //显示创建房间时间
	}
}

// NewFreeGameData 百人房间桌子数据
func NewFreeGameData(node, id string, gtype int32, dtype int32) *data.Game {
	return &data.Game{
		Id:     id,
		Unique: bson.NewObjectId().String(),
		Gtype:  gtype,
		Rtype:  int32(pb.ROOM_TYPE2),
		Dtype:  dtype,
		Status: 1,
		Count:  100,
		Node:   node,
		Ctime:  bson.Now(),
		Pub:    true,
		Mode:   1,
	}
}

func NewHuaCoinGameData(gameid string, desk_type int32) *data.Game {
	game := config.GetGame(gameid)
	game.Unique = data.ObjectIdString(bson.NewObjectId())
	game.DeskType = desk_type
	//检查其他字段
	return &game
}

// NewCoinGameData 自由房间桌子数据
func NewCoinGameData(node string, gtype, dtype, ltype int32) *data.Game {
	g := &data.Game{
		//Id:     bson.NewObjectId().String(),
		Id:     data.ObjectIdString(bson.NewObjectId()),
		Gtype:  gtype,
		Rtype:  int32(pb.ROOM_TYPE0),
		Dtype:  dtype,
		Ltype:  ltype,
		Status: 1,
		Count:  5,
		Ante:   100,
		Deal:   true,
		Chip:   20000,
		Sit:    20000,
		Node:   node,
		Ctime:  bson.Now(),
		Pub:    true,
		Mode:   1,
	}
	switch ltype {
	case int32(pb.ROOM_LEVEL0):
		g.Ante = 10
		g.Chip = 1000
		g.Sit = 1000
		g.Minimum = 500
		g.Maximum = 1000
	case int32(pb.ROOM_LEVEL1):
		g.Ante = 100
		g.Chip = 10000
		g.Sit = 10000
		g.Minimum = 5000
		g.Maximum = 10000
	case int32(pb.ROOM_LEVEL2):
		g.Ante = 200
		g.Chip = 20000
		g.Sit = 20000
		g.Minimum = 10000
		g.Maximum = 20000
	case int32(pb.ROOM_LEVEL3):
		g.Ante = 500
		g.Chip = 50000
		g.Sit = 50000
		g.Minimum = 25000
		g.Maximum = 50000
	case int32(pb.ROOM_LEVEL4):
		g.Ante = 600
		g.Chip = 120000
		g.Sit = 120000
		g.Minimum = 60000
		g.Maximum = 120000
	}
	return g
}

// MatchLevel 匹配等级
func MatchLevel(coin int64) int32 {
	// if coin >= 120000 {
	// 	return int32(pb.ROOM_LEVEL4)
	// } else if coin >= 50000 {
	// 	return int32(pb.ROOM_LEVEL3)
	// } else if coin >= 20000 {
	// 	return int32(pb.ROOM_LEVEL2)
	// } else if coin >= 10000 {
	// 	return int32(pb.ROOM_LEVEL1)
	// } else if coin >= 1000 {
	// 	return int32(pb.ROOM_LEVEL0)
	// }

	if coin >= 0 {
		return int32(pb.ROOM_LEVEL0)
	}

	return int32(-1)
}

// NewPrivGameData 私人房间桌子数据
func NewPrivGameData(arg *pb.CreateDesk) *data.DeskData {
	return &data.DeskData{
		Unique: data.ObjectIdString(bson.NewObjectId()),
		Gtype:  arg.Gtype,
		Rtype:  arg.Rtype,
		Cid:    arg.Cid,
		Ctime:  uint32(utils.Timestamp()),
		Expire: utils.Timestamp() + 86400, // todo x秒后未开局强制解散
		Deal:   true,
		Gmode:  arg.Gmode,
	}
}

// SetNiuCoinGame 初始化房间配置
// func SetNiuCoinGame() {
// 	for _, v := range pb.DeskType_value {
// 		if v == 3 {
// 			continue
// 		}
// 		for _, l := range pb.RoomLevel_value {
// 			gameData := NewCoinGameData("game.niu1", int32(pb.NIU), v, l)
// 			config.SetGame(*gameData)
// 			gameData.Save()
// 		}
// 	}
// }

// // SetEBGCoinGame 初始化房间配置
// func SetEBGCoinGame() {
// 	for _, v := range pb.DeskType_value {
// 		for _, l := range pb.RoomLevel_value {
// 			gameData := NewCoinGameData("game.ebg1", int32(pb.EBG), v, l)
// 			config.SetGame(*gameData)
// 			gameData.Save()
// 		}
// 	}
// }

/*
//SetGameList 游戏房间配置，测试数据
func SetGameList() {
	g1 := data.Game{
		Id:     bson.NewObjectId().String(),
		Gtype:  data.GAME_NIU,
		Rtype:  data.ROOM_TYPE0,
		Ltype:  data.GAME_BJPK10,
		Name:   "牛牛1区",
		Status: 1,
		Count:  100,
		Ante:   1,
		Cost:   5,
		Vip:    0,
		Chip:   0,
		Deal:   true,
		Carry:  20000,
		Down:   10000,
		Top:    60000,
		Sit:    20000,
		Del:    1,
		Ctime:  time.Now(),
	}
	g2 := data.Game{
		Id:     bson.NewObjectId().String(),
		Gtype:  data.GAME_SAN,
		Rtype:  data.ROOM_TYPE0,
		Ltype:  data.GAME_BJPK10,
		Name:   "三公1区",
		Status: 1,
		Count:  100,
		Ante:   1,
		Cost:   5,
		Vip:    0,
		Chip:   0,
		Deal:   true,
		Carry:  20000,
		Down:   10000,
		Top:    60000,
		Sit:    20000,
		Del:    1,
		Ctime:  time.Now(),
	}
	g3 := data.Game{
		Id:     bson.NewObjectId().String(),
		Gtype:  data.GAME_JIU,
		Rtype:  data.ROOM_TYPE0, //免佣房间
		Ltype:  data.GAME_BJPK10,
		Name:   "北京赛车1区",
		Status: 1,
		Count:  100,
		Ante:   1,
		Cost:   5,
		Vip:    0,
		Chip:   0,
		Deal:   false, //无庄
		Carry:  20000,
		Down:   10000,
		Top:    60000,
		Sit:    20000,
		Del:    0,
		Node:   "game.huiyin1",
		Ctime:  time.Now(),
	}
	g4 := data.Game{
		Id:     bson.NewObjectId().String(),
		Gtype:  data.GAME_JIU,
		Rtype:  data.ROOM_TYPE1, //抽佣房间
		Ltype:  data.GAME_BJPK10,
		Name:   "北京赛车2区",
		Status: 1,
		Count:  100,
		Ante:   1,
		Cost:   5,
		Vip:    0,
		Chip:   0,
		Deal:   false, //无庄
		Carry:  20000,
		Down:   10000,
		Top:    60000,
		Sit:    20000,
		Del:    0,
		Node:   "game.huiyin1",
		Ctime:  time.Now(),
	}
	g5 := data.Game{
		Id:     bson.NewObjectId().String(),
		Gtype:  data.GAME_JIU,
		Rtype:  data.ROOM_TYPE0, //免佣房间
		Ltype:  data.GAME_MLAFT,
		Name:   "幸运飞艇1区",
		Status: 1,
		Count:  100,
		Ante:   1,
		Cost:   5,
		Vip:    0,
		Chip:   0,
		Deal:   false, //有庄
		Carry:  20000,
		Down:   10000,
		Top:    60000,
		Sit:    20000,
		Del:    0,
		Node:   "game.huiyin2",
		Ctime:  time.Now(),
	}
	g6 := data.Game{
		Id:     bson.NewObjectId().String(),
		Gtype:  data.GAME_JIU,
		Rtype:  data.ROOM_TYPE1, //抽佣房间
		Ltype:  data.GAME_MLAFT,
		Name:   "幸运飞艇2区",
		Status: 1,
		Count:  100,
		Ante:   1,
		Cost:   5,
		Vip:    0,
		Chip:   0,
		Deal:   false, //有庄
		Carry:  20000,
		Down:   10000,
		Top:    60000,
		Sit:    20000,
		Del:    0,
		Node:   "game.huiyin2",
		Ctime:  time.Now(),
	}
	config.SetGame(g1)
	config.SetGame(g2)
	config.SetGame(g3)
	config.SetGame(g4)
	config.SetGame(g5)
	config.SetGame(g6)
}

//SetShopList 商城配置
func SetShopList() {
	s1 := data.Shop{
		Id:     "111",
		Status: 1,
		Propid: 1,
		Payway: 1,
		Number: 10,
		Price:  1000,
		Name:   "筹码",
		Info:   "筹码",
		Del:    0,
		Etime:  time.Now().AddDate(0, 0, 5),
		Ctime:  time.Now(),
	}
	s2 := data.Shop{
		Id:     "112",
		Status: 1,
		Propid: 1,
		Payway: 1,
		Number: 10,
		Price:  1000,
		Name:   "筹码",
		Info:   "筹码",
		Del:    0,
		Etime:  time.Now().AddDate(0, 0, 5),
		Ctime:  time.Now(),
	}
	s3 := data.Shop{
		Id:     "113",
		Status: 1,
		Propid: 1,
		Payway: 1,
		Number: 10,
		Price:  1000,
		Name:   "筹码",
		Info:   "筹码",
		Del:    0,
		Etime:  time.Now().AddDate(0, 0, 5),
		Ctime:  time.Now(),
	}
	config.SetShop(s1)
	config.SetShop(s2)
	config.SetShop(s3)
}
*/

// SetShopList 添加商城物品
func SetShopList() {
	//NewShop("1", 1, 2, 2, 10000, 100, "金币", "金币10000")
	//NewShop("2", 1, 2, 2, 20000, 200, "金币", "金币20000")
	//NewShop("3", 1, 2, 2, 50000, 450, "金币", "金币50000")
	//NewShop("4", 1, 1, 1, 100, 10, "钻石", "钻石100")
	//NewShop("5", 1, 1, 1, 200, 20, "钻石", "钻石200")
	//NewShop("6", 1, 1, 1, 500, 45, "钻石", "钻石500")
	NewShop("1", 1, 1, 1, 200, 200, 2, "Bouns", "Bouns202")
	NewShop("2", 1, 1, 1, 400, 400, 12, "Bouns", "Bouns412")
	NewShop("3", 1, 1, 1, 500, 500, 15, "Bouns", "Bouns515")
	NewShop("4", 1, 1, 1, 1000, 1000, 40, "Bouns", "Bouns1040")
	NewShop("5", 1, 1, 1, 1500, 1500, 60, "Bouns", "Bouns1560")
	NewShop("6", 1, 1, 1, 2000, 2000, 100, "Bouns", "Bouns2100")
	NewShop("7", 1, 1, 1, 5000, 5000, 300, "Bouns", "Bouns5300")
	NewShop("8", 1, 1, 1, 10000, 10000, 700, "Bouns", "Bouns10700")
}

// NewShop 添加商品
func NewShop(id string, status, propid, payway int,
	number, price, give uint32, name, info string) {
	// t := data.Shop{
	// 	Id:     id,
	// 	Status: status,
	// 	Propid: propid,
	// 	Payway: payway,
	// 	Number: number,
	// 	Give:   give,
	// 	Price:  price,
	// 	Name:   name,
	// 	Info:   info,
	// 	// Etime:  time.Now().AddDate(0, 0, 100),
	// 	Ctime: time.Now(),
	// }
	// config.SetShop(t)
	// t.Save()
}

// 移除玩家
func RemoveRealUser(userid string, base *data.DeskBase) {
	removeIndex := -1
	for i, v := range base.RealUser {
		if userid == v {
			removeIndex = i
			break
		}
	}
	if removeIndex >= 0 {
		base.RealUser = append(base.RealUser[:removeIndex], base.RealUser[removeIndex+1:]...)
	}
}

// 创建假位置
func CreateFakeSeats(robotNum int, seats map[uint32]*data.DeskSeat, maxNum uint32) map[uint32]string {
	seatmap := make(map[uint32]string)
	// 计算本轮需要几个假位置
	choices := make([]utils.Choice, 0)
	choices = append(choices, utils.Choice{Weight: 35, Item: 0})
	choices = append(choices, utils.Choice{Weight: 25, Item: 2})
	choices = append(choices, utils.Choice{Weight: 20, Item: 3})
	choices = append(choices, utils.Choice{Weight: 15, Item: 4})
	choices = append(choices, utils.Choice{Weight: 5, Item: 5})
	choice, err := utils.WeightedChoice(choices)
	if err != nil {
		return seatmap
	}
	seatNum := choice.Item.(int)
	for k := range seats {
		seatmap[k] = ""
	}
	if len(seatmap) >= seatNum {
		return seatmap
	}
	var seat uint32 = uint32(utils.RandInt32N(int32(maxNum)) + 1)
	for i := 0; i < int(maxNum); i++ {
		if _, ok := seatmap[seat]; ok {
			seat++
			if seat > maxNum {
				seat = 1
			}
			continue
		} else {
			seatmap[seat] = ""
		}
		if len(seatmap) >= seatNum {
			return seatmap
		}
	}
	glog.Debugf("fakeseat:%v", seatmap)
	return seatmap
}

// 防伙牌分配位置
func GankPartner(robot bool, count uint32, seats map[uint32]string) uint32 {
	double := false
	if robot {
		double = true
	}
	if double {
		i := 0
		for k := range seats {
			if k%2 == 0 {
				i++
			}
		}
		if count/2 > uint32(i) {
			// 还有双号座位
			var j uint32
			for j = 2; j <= count; j += 2 {
				if _, ok := seats[j]; !ok {
					return j
				}
			}
		}
	} else {
		i := 0
		for k := range seats {
			if k%2 == 1 {
				i++
			}
		}
		max := int(math.Ceil(float64(count) / 2))
		if max > i {
			// 还有单号座位
			var j uint32
			for j = 1; j <= count; j += 2 {
				if _, ok := seats[j]; !ok {
					return j
				}
			}
		}
	}
	// 单双号都坐满了，直接随机
	index := uint32(utils.RandInt32N(int32(count)) + 1)
	for i := 1; i <= int(count); i++ {
		if _, ok := seats[index]; !ok {
			return index
		}
		index++
		if index > count {
			index = 1
		}
	}
	return 0
}

// 匹配互斥
func MatchMutex(userid string, log map[string]map[string]string, realUsers []string) bool {
	if _, ok := log[userid]; !ok || len(realUsers) <= 0 {
		return false
	}
	match := log[userid]
	for _, v := range realUsers {
		if _, ok := match[v]; ok {
			return true
		}
	}
	return false
}

// 百人胜场
func GenNewbiewFreeWin(gtype int32) []data.FreeWin {
	wins := make([]data.FreeWin, 0)
	win := utils.RandIntN(16)
	for i := 0; i < win; i++ {
		wins = append(wins, data.FreeWin{Wtype: 1})
	}
	return wins
}

// 百人泡沫胜率
func GetFreeFrothWinning(bets int64) int32 {
	beginner := config.GetBeginner(1)
	if beginner.Id == 0 {
		glog.Errorf("no find beginner config")
		return 1000
	}
	interval := beginner.FrothBetInterval
	if len(interval) <= 0 {
		glog.Errorf("no find beginner config")
		return 1000
	}
	winning := beginner.FrothWonRate[len(interval)]
	for i, v := range interval {
		if bets < int64(v) {
			winning = beginner.FrothWonRate[i]
		}
	}
	return int32(winning)
}

// 玩家系数
func GetFactor(u *data.User) float64 {
	money := u.GetMoney()
	if money == 0 {
		if u.RegistArea == 0 {
			return 1
		}
		// b类玩家默认1分钱
		money = 1
	}

	seed := 0.993
	bean := config.GetBeginner(1)
	if bean.Id != 0 && bean.FactorSeed != 0 {
		seed = bean.FactorSeed
	}

	m := math.Pow(seed, float64(money)/10000)
	diamond := int64(math.Max(float64(u.GetDiamond()-u.PrivDiamond), 0))
	// B类玩家
	// if u.RegistArea == 1 {
	// 	// 玩家系数=(总彩金+总提现)/（总充值+起伏线）*[0.99^(充值金额/10000)]
	// 	line := FluctuateLine(u) // 起伏线
	// 	ret, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(diamond+int64(u.GetCashOut()))/(float64(money))*m), 64)
	// 	return math.Min(2, ret)
	// }

	// A类玩家
	// (总彩金+总提现)/总充值*[0.99^(充值金额/10000)]
	ret, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(diamond+int64(u.GetCashOut()))/float64(money)/m), 64)
	return math.Min(2, ret)
}

// 起伏线
func FluctuateLine(u *data.User) float64 {
	betAvg := u.BetAvg()
	// 充值均价
	avgCharge := u.ChargeAvg()
	// 提现均价
	avgWithdraw := u.WithdrawAvg()
	line := betAvg*10 + float64(u.FirstChargeVal) + avgCharge - 40000 + (avgWithdraw-20000)*0.5

	// 如果起付线算出负值，那么起伏线=0
	if line < 0 {
		line = 0
	}
	return math.Min(line, 1000)
}

// B类百人必赢策略
func FreeWinStrategy_B(u *data.User, bet int64, factor float64) bool {
	// B类正常或者泡沫
	if !u.IsB() || (u.State != 2 && u.State != 4) {
		return false
	}

	if FluctuateLine(u) < 500 {
		return false
	}

	if u.LoginTimes > 2 {
		return false
	}

	if u.FreeWelfare >= 1 {
		return false
	}

	if factor >= 0.5 && u.Diamond+bet >= 5000 {
		return false
	}

	return true
}

// 检测新手提现踢出标记
func CheckNewbieWithdrawFlag(u *data.User) (bool, int64) {
	if u.State != data.NoveiceState && u.State != data.ExceptionState {
		return true, 0
	}
	conf := table.GetTables().TransferDemoCashTable.Get()
	if conf.Open == 0 {
		return true, 0
	}
	level := len(conf.NewbieCarry)
	for i, v := range conf.NewbieCarry {
		if u.OutDiamond < int64(v) {
			level = i
			break
		}
	}

	if level > 0 && level != u.KickWithdrawFlag {
		return false, int64(conf.NewbieCarry[level-1])
	}
	return true, 0
}

func IsFreeGame(gtype int32) bool {
	switch gtype {
	case int32(pb.LHD), int32(pb.SEVEN), int32(pb.ABAR),
		int32(pb.REDBLACK), int32(pb.CRASH), int32(pb.PLANE), int32(pb.LOTTERY):
		return true
	}
	return false
}
