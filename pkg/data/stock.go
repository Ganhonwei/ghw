package data

import (
	"fmt"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"strconv"

	"gopkg.in/mgo.v2/bson"
)

type StockHistory struct {
	Timestamp    int64 `bson:"timestamp" json:"timestamp"`           //时间戳
	CashStock    int64 `bson:"cash_stock" json:"cash_stock"`         //彩金库存
	BonusStock   int64 `bson:"bonus_stock" json:"bonus_stock"`       //奖励金库存
	CashMingTax  int64 `bson:"cash_ming_tax" json:"cash_ming_tax"`   //彩金明税
	BonusMingTax int64 `bson:"bonus_ming_tax" json:"bonus_ming_tax"` //奖励金明税
	CashAnTax    int64 `bson:"cash_an_tax" json:"cash_an_tax"`       //彩金暗税
	BonusAnTax   int64 `bson:"bonus_an_tax" json:"bonus_an_tax"`     //奖励金暗税
}

type StockHistorySlice []StockHistory

func (s StockHistorySlice) Len() int           { return len(s) }
func (s StockHistorySlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s StockHistorySlice) Less(i, j int) bool { return s[i].Timestamp < s[j].Timestamp }

type Stock struct {
	Id           string         `bson:"_id" json:"id"`                        //房间ID
	Type         int32          `bson:"type" json:"type"`                     //游戏分类 1:tp类 2:rm类 3：百人类
	CashStock    int64          `bson:"cash_stock" json:"cash_stock"`         //彩金库存
	BonusStock   int64          `bson:"bonus_stock" json:"bonus_stock"`       //奖励金库存
	CashMingTax  int64          `bson:"cash_ming_tax" json:"cash_ming_tax"`   //彩金明税
	BonusMingTax int64          `bson:"bonus_ming_tax" json:"bonus_ming_tax"` //奖励金明税
	CashAnTax    int64          `bson:"cash_an_tax" json:"cash_an_tax"`       //彩金暗税
	BonusAnTax   int64          `bson:"bonus_an_tax" json:"bonus_an_tax"`     //奖励金暗税
	Factor       float64        `bson:"factor" json:"factor"`                 //当前房间系数
	History      []StockHistory `bson:"history" json:"history"`               //库存曲线
}

func (s *Stock) GetById(id string) {
	GetByQ(Stocks, bson.M{"_id": id}, s)
}

func (s *Stock) SaveStock() bool {
	set := bson.M{
		"cash_stock":     s.CashStock,
		"bonus_stock":    s.BonusStock,
		"cash_ming_tax":  s.CashMingTax,
		"bonus_ming_tax": s.BonusMingTax,
		"cash_an_tax":    s.CashAnTax,
		"bonus_an_tax":   s.BonusAnTax,
		"factor":         s.Factor,
	}

	return Upsert(Stocks, bson.M{"_id": s.Id}, bson.M{"$set": set})
}

func (s *Stock) SaveHistory() bool {
	set := bson.M{
		"history": s.History,
	}
	return Upsert(Stocks, bson.M{"_id": s.Id}, bson.M{"$set": set})
}

func (s *Stock) SaveGameStock() bool {
	set := bson.M{
		"cash_stock":     s.CashStock,
		"bonus_stock":    s.BonusStock,
		"cash_ming_tax":  s.CashMingTax,
		"bonus_ming_tax": s.BonusMingTax,
		"cash_an_tax":    s.CashAnTax,
		"bonus_an_tax":   s.BonusAnTax,
		"factor":         s.Factor,
	}

	return Upsert(GameStocks, bson.M{"_id": s.Id}, bson.M{"$set": set})
}
func (s *Stock) SaveGameHistory() bool {
	set := bson.M{
		"history": s.History,
	}
	return Upsert(GameStocks, bson.M{"_id": s.Id}, bson.M{"$set": set})
}

func (s *Stock) Init() bool {
	return Insert(Stocks, s)
}

// 刷新房间系数
func (s *Stock) RefreshFactor(expect int64) {
	if expect != 0 {
		var err error
		s.Factor, err = strconv.ParseFloat(fmt.Sprintf("%.2f", float64(s.CashStock)/float64(expect)), 64)
		if err != nil {
			glog.Errorf("RefreshFactor error: %s", s.Id)
		}
		//房间系数0-2
		if s.Factor < 0 {
			s.Factor = 0
		} else if s.Factor > 2 {
			s.Factor = 2
		}

	}
}

func GetStockList() []*Stock {
	var list []*Stock
	ListByQ(Stocks, bson.M{}, &list)
	return list
}

func GetGameStockList() []*Stock {
	var list []*Stock
	ListByQ(GameStocks, bson.M{}, &list)
	return list
}

// 输赢分
type CashFlow struct {
	Id        string `bson:"_id" json:"id"`               // id
	Gtype     int32  `json:"gtype" bson:"gtype"`          // 游戏类型
	WinScore  int64  `json:"winScore" bson:"win_score"`   // 赢分
	LoseScore int64  `json:"loseScore" bson:"lose_score"` // 输分
	Date      string `json:"date" bson:"date"`            // 日期
	Ctime     int64  `json:"ctime" bson:"ctime"`          // 时间
}

func (c *CashFlow) Save() {
	Upsert(CashFlows, bson.M{"_id": c.Id}, bson.M{"$set": c})
}

func GetCashFlowList() []*CashFlow {
	var list []*CashFlow
	date := utils.DateLocalStr()
	ListByQ(CashFlows, bson.M{"date": date}, &list)
	return list
}

type NewbieStock struct {
	Id          int32 `bson:"_id" json:"id"`                      //gtype
	CashStock   int64 `bson:"cash_stock" json:"cash_stock"`       //彩金库存
	CashMingTax int64 `bson:"cash_ming_tax" json:"cash_ming_tax"` //彩金明税
	CashAnTax   int64 `bson:"cash_an_tax" json:"cash_an_tax"`     //彩金暗税
	PartIn      int64 `json:"partIn" bson:"part_in"`              //参与人次
	PlayTimes   int64 `bson:"play_times" json:"play_times"`       //总局数
	GameTime    int64 `bson:"game_time" json:"game_time"`         //总时长
}

func (s *NewbieStock) SaveNewbieStock() bool {
	return Upsert(NewbieStocks, bson.M{"_id": s.Id}, bson.M{"$set": s})
}

func GetNewbieStockList() []*NewbieStock {
	var list []*NewbieStock
	ListByQ(NewbieStocks, bson.M{}, &list)
	return list
}

type TPStoryChargeStock struct {
	Id    string `bson:"_id" json:"id"`      //房间ID
	Stock int64  `bson:"stock" json:"stock"` //库存值
	// History []TPStoryChargeStockHistory `bson:"history" json:"history"` //库存曲线
}

func (s *TPStoryChargeStock) GetById(id string) {
	GetByQ(TPStoryChargeStocks, bson.M{"_id": id}, s)
}

func (s *TPStoryChargeStock) SaveStock() bool {
	set := bson.M{
		"stock": s.Stock,
	}
	return Upsert(TPStoryChargeStocks, bson.M{"_id": s.Id}, bson.M{"$set": set})
}

func GetTPStoryChargeStockList() []*TPStoryChargeStock {
	var list []*TPStoryChargeStock
	ListByQ(TPStoryChargeStocks, bson.M{}, &list)
	return list
}

type TPStoryChargeStockHistory struct {
	Id        string `bson:"_id" json:"id"`              // id
	Userid    string `bson:"userid" json:"userid"`       // 用户id
	GameId    string `bson:"gameid" json:"gameid"`       // 游戏id
	Change    int64  `bson:"change" json:"change"`       // 变化
	Before    int64  `bson:"before" json:"before"`       // 前值
	After     int64  `bson:"after" json:"after"`         // 账变后
	DetailId  string `bson:"detailid" json:"detailid"`   // 轮次
	Timestamp int64  `bson:"timestamp" json:"timestamp"` // 时间戳
}

func (s *TPStoryChargeStockHistory) Insert() bool {
	return Insert(TPStoryChargeStockHistoryLogs, s)
}
