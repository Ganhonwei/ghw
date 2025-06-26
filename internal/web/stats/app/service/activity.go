package service

import (
	"errors"
	"goserver/internal/web/stats/app/entity"

	"gopkg.in/mgo.v2/bson"
)

// 活动管理
type activityService struct{}

// 新增基础付费活动
func (this *activityService) AddBasicPay(res *entity.BasicPayActivity) error {
	info := new(entity.BasicPayActivity)
	GetByQ(BasicPays, bson.M{"date": res.Date}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"shop_number":  res.ShopNumber,
			"shop_diamond": res.ShopDiamond,
			"shop_coin":    res.ShopCoin,
			"rm_number":    res.RmNumber,
			"rm_diamond":   res.RmDiamond,
			"rm_coin":      res.RmCoin,
			"jy_number":    res.JyNumber,
			"jy_diamond":   res.JyDiamond,
			"jy_coin":      res.JyCoin,
			"sc_number":    res.ScNumber,
			"sc_diamond":   res.ScDiamond,
			"sc_coin":      res.ScCoin,
		}
		if Update(BasicPays, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + res.Id)
	} else {
		// 新增
		res.Id = bson.NewObjectId().Hex()
		if !Insert(BasicPays, res) {
			return errors.New("写入失败:" + res.Id)
		}
		return nil
	}
}

// 新增基础免费活动
func (this *activityService) AddBasicFree(res *entity.BasicFreeActivity) error {
	info := new(entity.BasicFreeActivity)
	GetByQ(BasicFrees, bson.M{"date": res.Date}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"sign_in_number":       res.SignInNumber,
			"sign_in_diamond":      res.SignInDiamond,
			"sign_in_coin":         res.SignInCoin,
			"online_number":        res.OnlineNumber,
			"online_diamond":       res.OnlineDiamond,
			"online_coin":          res.OnlineCoin,
			"task_number":          res.TaskNumber,
			"task_diamond":         res.TaskDiamond,
			"task_coin":            res.TaskCoin,
			"trail_or_set_number":  res.TrailOrSetNumber,
			"trail_or_set_diamond": res.TrailOrSetDiamond,
			"trail_or_set_coin":    res.TrailOrSetCoin,
		}
		if Update(BasicFrees, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + res.Id)
	} else {
		// 新增
		res.Id = bson.NewObjectId().Hex()
		if !Insert(BasicFrees, res) {
			return errors.New("写入失败:" + res.Id)
		}
		return nil
	}
}

// 添加分享活动数据
func (this *activityService) AddShareData(rate *entity.ShareActivityData) error {
	info := new(entity.ShareActivityData)
	GetByQ(ShareDatas, bson.M{"ctime": rate.Ctime}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"register_number":     rate.RegisterNumber,
			"water_out_put":       rate.WaterOutPut,
			"first_order_out_put": rate.FirstOrderOutPut,
			"gross_output":        rate.GrossOutput,
		}
		if Update(ShareDatas, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + rate.Id)
	} else {
		// 新增
		rate.Id = bson.NewObjectId().Hex()
		if !Insert(ShareDatas, rate) {
			return errors.New("写入失败:" + rate.Id)
		}
		return nil
	}
}

// 添加彩票活动数据
func (this *activityService) AddOrUpdateLottery(user *entity.LotteryActivity) error {
	info := new(entity.LotteryActivity)
	GetByQ(LotteryActivitys, bson.M{"date": user.Date}, info)
	if info.Id != "" {
		m := bson.M{"_id": info.Id}
		n := bson.M{
			"click_number": user.ClickNumber,
			"gj_number":    user.GJNumber,
			"hj_number":    user.HJNumber,
			"first_prize":  user.FirstPrize,
			"second_prize": user.SecondPrize,
			"third_prize":  user.ThirdPrize,
			"fourth_prize": user.FourthPrize,
			"fifth_prize":  user.FifthPrize,
			"expend_bouns": user.ExpendBouns,
		}
		if Update(LotteryActivitys, m, bson.M{"$set": n}) {
			return nil
		}
		return errors.New("更新失败:" + user.Id)
	} else {
		// 新增
		user.Id = bson.NewObjectId().Hex()
		if !Insert(LotteryActivitys, user) {
			return errors.New("写入失败:" + user.Id)
		}
		return nil
	}
}
