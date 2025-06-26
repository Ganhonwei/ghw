package dbms

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/config"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func (a *RoomActor) handlerWeb(arg *pb.WebRequest, rsp *pb.WebResponse, ctx actor.Context) {
	switch arg.Code {
	case pb.WebModifyStock:
		msg := new(pb.ModifyStock)
		err := msg.Unmarshal(arg.Data)
		if err != nil {
			rsp.ErrMsg = fmt.Sprintf("msg error: %v", err)
			return
		}
		ret := a.modifyStock(msg)
		result, err := ret.Marshal()
		if err != nil {
			rsp.ErrMsg = fmt.Sprintf("msg error: %v", err)
			return
		}
		rsp.Result = result
	case pb.WebNewbieStock:
		msg := new(data.ModifyNewbieStock)
		err := json.Unmarshal(arg.Data, &msg)
		if err != nil {
			rsp.ErrMsg = fmt.Sprintf("msg error: %v", err)
			return
		}
		a.modifyNewbieStock(msg)
	case pb.WebModifyTpStoryStock:
		msg := new(pb.ModifyTpStoryStock)
		err := msg.Unmarshal(arg.Data)
		if err != nil {
			rsp.ErrMsg = fmt.Sprintf("msg error: %v", err)
			return
		}
		ret := a.modifyTpStoryStock(msg)
		result, err := ret.Marshal()
		if err != nil {
			rsp.ErrMsg = fmt.Sprintf("msg error: %v", err)
			return
		}
		rsp.Result = result
	case pb.WebGameStock:
		msg := new(data.ModifyGameStock)
		err := json.Unmarshal(arg.Data, &msg)
		if err != nil {
			rsp.ErrMsg = fmt.Sprintf("msg error: %v", err)
			return
		}
		a.modifyGameStock(msg)
	}
}

func (a *RoomActor) modifyStock(arg *pb.ModifyStock) *pb.ModifyedStock {
	ret := &pb.ModifyedStock{}
	if v, ok := a.stocks[arg.GameId]; ok {
		//库存
		// if arg.CashStock != 0 {
		v.CashStock = arg.CashStock
		// }
		// if arg.BonusStock != 0 {
		v.BonusStock = arg.BonusStock
		// }
		//明税
		// if arg.CashMingTax != 0 {
		v.CashMingTax = arg.CashMingTax
		// }
		// if arg.BonusMingTax != 0 {
		v.BonusMingTax = arg.BonusMingTax
		// }
		//暗税
		// if arg.CashAnTax != 0 {
		v.CashAnTax = arg.CashAnTax
		// }
		// if arg.BonusAnTax != 0 {
		v.BonusAnTax = arg.BonusAnTax
		// }

		//响应数据
		ret.GameId = arg.GameId
		ret.CashStock = v.CashStock
		ret.BonusStock = v.BonusStock
		ret.CashMingTax = v.CashMingTax
		ret.BonusMingTax = v.BonusMingTax
		ret.CashAnTax = v.CashAnTax
		ret.BonusAnTax = v.BonusAnTax
		//刷新房间系数
		game := config.GetGame(arg.GameId)
		v.RefreshFactor(game.GetStockExpect())
	}
	return ret
}

func (a *RoomActor) modifyNewbieStock(arg *data.ModifyNewbieStock) {
	if v, ok := a.newbieStocks[arg.Gtype]; ok {
		//库存
		// if arg.CashStock != 0 {
		v.CashStock = int64(arg.Stock)
		//刷新房间系数
	}
}

func (a *RoomActor) modifyGameStock(arg *data.ModifyGameStock) {
	if v, ok := a.gameStocks[arg.Gtype]; ok {
		//库存
		// if arg.CashStock != 0 {
		v.CashStock = arg.Stock
	}
}

func (a *RoomActor) modifyTpStoryStock(arg *pb.ModifyTpStoryStock) *pb.ModifyTpStoryStock {
	ret := &pb.ModifyTpStoryStock{GameId: arg.GameId}
	stock, ok := a.tpStoryChargeStock[arg.GameId]
	if !ok {
		stock = &data.TPStoryChargeStock{
			Id:    arg.GameId,
			Stock: 0,
		}
		a.tpStoryChargeStock[stock.Id] = stock
	}
	stock.Stock = arg.Stock
	ret.Stock = stock.Stock
	stock.SaveStock()
	return ret
}
