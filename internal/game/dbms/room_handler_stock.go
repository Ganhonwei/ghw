package dbms

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/config"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"sort"
	"strings"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"gopkg.in/mgo.v2/bson"
)

// 获取房间系数
func (a *RoomActor) GetRoomFactor(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.GetRoomFactor)
	glog.Infof("%v", arg)

	rsp := &pb.GetedRoomFactor{}
	if v, ok := a.stocks[arg.GameId]; ok {
		rsp.GameId = arg.GameId
		rsp.Factor = v.Factor
		rsp.CashStock = v.CashStock
	}
	ctx.Respond(rsp)
}

// 增量修改房间库存
func (a *RoomActor) ChangeStock(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.ChangeStock)
	glog.Info("%v", arg)

	rsp := &pb.ChangedStock{}
	if v, ok := a.stocks[arg.GameId]; ok {
		//库存
		if arg.CashStock != 0 {
			v.CashStock += arg.CashStock
		}
		if arg.BonusStock != 0 {
			v.BonusStock += arg.BonusStock
		}
		//明税
		if arg.CashMingTax > 0 {
			v.CashMingTax += arg.CashMingTax
		}
		if arg.BonusMingTax > 0 {
			v.BonusMingTax += arg.BonusMingTax
		}
		//暗税
		if arg.CashAnTax > 0 {
			v.CashAnTax += arg.CashAnTax
		}
		if arg.BonusAnTax > 0 {
			v.BonusAnTax += arg.BonusAnTax
		}

		//响应数据
		rsp.GameId = arg.GameId
		rsp.CashStock = v.CashStock
		rsp.BonusStock = v.BonusStock
		rsp.CashMingTax = v.CashMingTax
		rsp.BonusMingTax = v.BonusMingTax
		rsp.CashAnTax = v.CashAnTax
		rsp.BonusAnTax = v.BonusAnTax
		//刷新房间系数
		game := config.GetGame(arg.GameId)
		v.RefreshFactor(game.GetStockExpect())
	}
	// 输赢分
	ctx.Respond(rsp)

	// game := config.GetGame(arg.GameId)
	if !arg.Real {
		return
	}

	// 游戏库存
	if v, ok := a.gameStocks[arg.Gtype]; ok {
		//库存
		if arg.CashStock != 0 {
			v.CashStock += arg.CashStock
		}
		//明税
		if arg.CashMingTax > 0 {
			v.CashMingTax += arg.CashMingTax
		}
		//暗税
		if arg.CashAnTax > 0 {
			v.CashAnTax += arg.CashAnTax
		}
	} else if arg.Gtype >= 0 {
		a.gameStocks[arg.Gtype] = &data.Stock{
			Id:          utils.String(arg.Gtype),
			CashStock:   arg.CashStock,
			CashMingTax: arg.CashMingTax,
			CashAnTax:   arg.CashAnTax,
		}
	}

	flow := new(data.CashFlow)
	date := utils.DateLocalStr()
	t := fmt.Sprintf("%s_%d", date, arg.Gtype)
	if g, ok := a.gameScores[t]; ok {
		flow = g
	} else {
		flow = &data.CashFlow{
			Id:    t,
			Gtype: arg.Gtype,
			Date:  date,
			Ctime: utils.BsonNow().Unix(),
		}
		a.gameScores[t] = flow
	}

	if arg.CashStock > 0 {
		flow.WinScore += arg.CashStock
	} else {
		flow.LoseScore -= arg.CashStock
	}
}

// 终值修改房间库存
func (a *RoomActor) ModifyStock(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.ModifyStock)
	glog.Info("%v", arg)

	rsp := &pb.ModifyedStock{}
	if v, ok := a.stocks[arg.GameId]; ok {
		//库存
		if arg.CashStock != 0 {
			v.CashStock = arg.CashStock
		}
		if arg.BonusStock != 0 {
			v.BonusStock = arg.BonusStock
		}
		//明税
		if arg.CashMingTax != 0 {
			v.CashMingTax = arg.CashMingTax
		}
		if arg.BonusMingTax != 0 {
			v.BonusMingTax = arg.BonusMingTax
		}
		//暗税
		if arg.CashAnTax != 0 {
			v.CashAnTax = arg.CashAnTax
		}
		if arg.BonusAnTax != 0 {
			v.BonusAnTax = arg.BonusAnTax
		}

		//响应数据
		rsp.GameId = arg.GameId
		rsp.CashStock = v.CashStock
		rsp.BonusStock = v.BonusStock
		rsp.CashMingTax = v.CashMingTax
		rsp.BonusMingTax = v.BonusMingTax
		rsp.CashAnTax = v.CashAnTax
		rsp.BonusAnTax = v.BonusAnTax
		//刷新房间系数
		game := config.GetGame(arg.GameId)
		v.RefreshFactor(game.GetStockExpect())
	}
	ctx.Respond(rsp)
}

func (a *RoomActor) SaveHistory() {
	currentTime := time.Now()
	threeDaysAgo := currentTime.AddDate(0, 0, -3)
	timestamp := threeDaysAgo.Unix()

	for _, s := range a.stocks {
		var temp []data.StockHistory
		for i := len(s.History) - 1; i >= 0; i-- {
			if s.History[i].Timestamp > timestamp {
				temp = append(temp, s.History[i])
			}
		}
		sort.Sort(data.StockHistorySlice(temp))
		new := data.StockHistory{
			Timestamp:    time.Now().Unix(),
			CashStock:    s.CashStock,
			BonusStock:   s.BonusStock,
			CashMingTax:  s.CashMingTax,
			BonusMingTax: s.BonusMingTax,
			CashAnTax:    s.CashAnTax,
			BonusAnTax:   s.BonusAnTax,
		}
		temp = append(temp, new)
		s.History = temp
		s.SaveHistory()
	}

	for _, s := range a.gameStocks {
		var temp []data.StockHistory
		for i := len(s.History) - 1; i >= 0; i-- {
			if s.History[i].Timestamp > timestamp {
				temp = append(temp, s.History[i])
			}
		}
		sort.Sort(data.StockHistorySlice(temp))
		new := data.StockHistory{
			Timestamp:    time.Now().Unix(),
			CashStock:    s.CashStock,
			BonusStock:   s.BonusStock,
			CashMingTax:  s.CashMingTax,
			BonusMingTax: s.BonusMingTax,
			CashAnTax:    s.CashAnTax,
			BonusAnTax:   s.BonusAnTax,
		}
		temp = append(temp, new)
		s.History = temp
		s.SaveGameHistory()
	}
}

// 房间库存历史记录
func (a *RoomActor) SaveStock() {
	for _, s := range a.stocks {
		s.SaveStock()
		// sh := data.StockHistory{
		// 	Timestamp:    time.Now().Unix(),
		// 	CashStock:    s.CashStock,
		// 	BonusStock:   s.BonusStock,
		// 	CashMingTax:  s.CashMingTax,
		// 	BonusMingTax: s.BonusMingTax,
		// 	CashAnTax:    s.CashAnTax,
		// 	BonusAnTax:   s.BonusAnTax,
		// }
		// s.History = append(s.History, sh)
	}

	for _, s := range a.gameStocks {
		s.SaveGameStock()
	}

	for _, s := range a.newbieStocks {
		s.SaveNewbieStock()
	}

	// 保存房间资源
	del := []string{}
	date := utils.DateLocalStr()
	for k, v := range a.gameScores {
		v.Save()
		if !strings.Contains(k, date) {
			del = append(del, k)
		}
	}

	// 移除非当天的数据
	for _, v := range del {
		delete(a.gameScores, v)
	}

	// 保存tp剧情库存
	for _, stock := range a.tpStoryChargeStock {
		stock.SaveStock()
	}
}

func (a *RoomActor) WebRequest(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.WebRequest)
	glog.Debugf("WebRequest %#v", arg)
	rsp := new(pb.WebResponse)
	rsp.Code = arg.Code
	a.handlerWeb(arg, rsp, ctx)
	ctx.Respond(rsp)
}

func (a *RoomActor) NewbieStock(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.NewbieStock)
	glog.Debugf("NewbieStock %#v", arg)

	if s, ok := a.newbieStocks[arg.Gtype]; ok {
		s.CashStock += arg.CashStock
		s.CashMingTax += arg.CashMing
		s.CashAnTax += arg.CashAn
		s.PlayTimes += 1
		s.GameTime += arg.GameTime
		if arg.First {
			s.PartIn += 1
		}
	} else {
		stock := &data.NewbieStock{
			Id:          arg.Gtype,
			CashStock:   arg.CashStock,
			CashMingTax: arg.CashMing,
			CashAnTax:   arg.CashAn,
			GameTime:    arg.GameTime,
			PlayTimes:   1,
			PartIn:      1,
		}
		a.newbieStocks[arg.Gtype] = stock
	}
}

func (a *RoomActor) GetGameStock(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.GetGameStock)
	glog.Info("%v", arg)

	rsp := new(pb.GetGamedStock)

	defer ctx.Respond(rsp)

	if s, ok := a.gameStocks[arg.Gtype]; ok {
		rsp.Stock = s.CashStock
	}
}

// 获取tp剧情模式局内充值平台亏损额
func (a *RoomActor) GetTpStoryChargeLoss(ctx actor.Context) {
	arg := ctx.Message().(*pb.GetTpStoryChargeLoss)
	glog.Debugf("GetTpStoryChargeLoss %#v", arg)
	rsp := new(pb.GotTpStoryChargeLoss)
	// 库存
	stock, ok := a.tpStoryChargeStock[arg.GameId]
	if !ok {
		ctx.Respond(rsp)
		return
	}
	rsp.ChargeLoss = stock.Stock
	ctx.Respond(rsp)
}

// 更新tp剧情模式局内充值平台亏损额
func (a *RoomActor) ChangeTpStoryChargeLoss(ctx actor.Context) {
	arg := ctx.Message().(*pb.ChangeTpStoryChargeLoss)
	glog.Debugf("ChangeTpStoryChargeLoss %#v", arg)

	// 库存
	stock, ok := a.tpStoryChargeStock[arg.GameId]
	if !ok {
		stock = &data.TPStoryChargeStock{
			Id:    arg.GameId,
			Stock: 0,
		}
		a.tpStoryChargeStock[stock.Id] = stock
	}

	curStock := stock.Stock
	stock.Stock += arg.ChargeLoss
	// stock.SaveStock() // tick定时保存

	log := &data.TPStoryChargeStockHistory{
		Id:        bson.NewObjectId().String(),
		Userid:    arg.Userid,
		GameId:    arg.GameId,
		Change:    arg.ChargeLoss,
		Before:    curStock,
		After:     stock.Stock,
		DetailId:  arg.DetailId,
		Timestamp: bson.Now().Unix(),
	}
	// todo 优化
	go log.Insert()
}
