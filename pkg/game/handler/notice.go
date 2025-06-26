package handler

import (
	"fmt"

	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/utils"
)

////GetNotice 公告列表
//func GetNotice(atype int32) (stoc *pb.NoticeRsp) {
//	stoc = new(pb.NoticeRsp)
//	list := config.GetNotices(atype)
//	for _, v := range list {
//		body := &pb.Notice{
//			Rtype:   int32(v.Rtype),
//			Acttype: int32(v.Acttype),
//			Content: v.Content,
//		}
//		stoc.List = append(stoc.List, body)
//	}
//	return
//}

// PackNotice 打包公告消息
// func PackNotice(msg *pb.NoticeRsp) {
// 	list := config.GetNotices(data.NOTICE_TYPE1)
// 	for _, v := range list {
// 		body := packNoticeMsg(&v)
// 		msg.List = append(msg.List, body)
// 	}
// }

func packNoticeMsg(v *data.Notice) (msg *pb.Notice) {
	msg = &pb.Notice{
		Content:    v.Content,
		Time:       utils.Time2LocalStr(v.Ctime),
		ExpireTime: utils.Time2LocalStr(v.Etime),
	}
	return
}

// PackUserNotice 打包玩家消息
// func PackUserNotice(arg *pb.NoticeReq) (msg *pb.NoticeRsp) {
// 	msg = new(pb.NoticeRsp)
// 	list, err := data.GetLogNotices(arg.Userid, int(arg.Page))
// 	if err != nil {
// 		glog.Errorf("get notice err : %v, arg %#v", err, arg)
// 		return
// 	}
// 	for _, v := range list {
// 		body := packNoticeMsg(v)
// 		msg.List = append(msg.List, body)
// 	}
// 	return
// }

// SaveNotice 保存消息记录
// func SaveNotice(arg *pb.LogNotice) {
// 	record := &data.Notice{
// 		Userid:  arg.Userid,
// 		Rtype:   int(arg.Rtype),
// 		Acttype: int(arg.Acttype),
// 		Content: arg.Content,
// 	}
// 	record.Save()
// }

// NewNotice 新消息
func NewNotice(rtype, atype int32, userid,
	content string) (record *pb.LogNotice, msg *pb.PushNoticeNtf) {
	record = &pb.LogNotice{
		Userid:  userid,
		Rtype:   rtype,
		Acttype: atype,
		Content: content,
	}
	msg = new(pb.PushNoticeNtf)
	msg.Info = &pb.Notice{
		Time:    utils.Time2LocalStr(utils.LocalTime()),
		Rtype:   rtype,
		Acttype: atype,
		Content: content,
	}
	msg.Userid = userid
	return
}

// BuyNotice 兑换金币消息
func BuyNotice(coin int64, userid string) (record *pb.LogNotice,
	msg *pb.PushNoticeNtf) {
	if coin <= 0 {
		return
	}
	content := fmt.Sprintf("恭喜你成功充值%d金豆", coin)
	return NewNotice(0, 0, userid, content)
}

// BankNotice 赠送消息
func BankNotice(coin int64, userid, from string) (record *pb.LogNotice,
	msg *pb.PushNoticeNtf) {
	if coin <= 0 {
		return
	}
	content := fmt.Sprintf("%s赠送给你%d金豆", from, coin)
	return NewNotice(data.NOTICE_TYPE5, 0, userid, content)
}

// GiveNotice 赠送消息
func GiveNotice(coin int64, userid, to string) (record *pb.LogNotice,
	msg *pb.PushNoticeNtf) {
	if coin <= 0 {
		return
	}
	content := fmt.Sprintf("恭喜成功赠送%d金豆给%s", coin, to)
	return NewNotice(data.NOTICE_TYPE5, 0, userid, content)
}

// BuildNotice 赠送消息
func BuildNotice(coin int64, build uint32, userid string) (record *pb.LogNotice,
	msg *pb.PushNoticeNtf) {
	if coin <= 0 {
		return
	}
	content := fmt.Sprintf("恭喜你成功邀请%d人获得%d金豆", build, coin)
	return NewNotice(data.NOTICE_TYPE3, 0, userid, content)
}

// LuckyNotice lucky消息
func LuckyNotice(coin int64, name, userid string) (record *pb.LogNotice,
	msg *pb.PushNoticeNtf) {
	if coin <= 0 {
		return
	}
	content := fmt.Sprintf("恭喜你完成幸运星%s获得%d金豆", name, coin)
	return NewNotice(data.NOTICE_TYPE3, 0, userid, content)
}

// BankOpenNotice 银行开通消息
func BankOpenNotice(coin int64, userid string) (record *pb.LogNotice,
	msg *pb.PushNoticeNtf) {
	if coin <= 0 {
		return
	}
	content := fmt.Sprintf("首次开通银行赠送%d豆子", coin)
	return NewNotice(data.NOTICE_TYPE3, 0, userid, content)
}

// ActNotice 活动奖励消息
// func ActNotice(arg *pb.AgentActivityProfit) (record *pb.LogNotice,
// 	msg *pb.PushNoticeNtf) {
// 	if arg.GetProfit() <= 0 {
// 		return
// 	}
// 	content := fmt.Sprintf("完成%s活动获取%d金豆豆", arg.GetTitle(), arg.GetProfit())
// 	return NewNotice(data.NOTICE_TYPE3, 0, arg.GetUserid(), content)
// }

// ActivityNotice activity消息
func ActivityNotice(title, timeStr, userid string) (record *pb.LogNotice,
	msg *pb.PushNoticeNtf) {
	content := fmt.Sprintf("你报名的活动%s,已于%s结束", title, timeStr)
	return NewNotice(data.NOTICE_TYPE3, 0, userid, content)
}

// TaskNotice task消息
func TaskNotice(coin int64, name, userid string) (record *pb.LogNotice,
	msg *pb.PushNoticeNtf) {
	if coin <= 0 {
		return
	}
	content := fmt.Sprintf("恭喜你完成任务%s获得%d金豆", name, coin)
	return NewNotice(data.NOTICE_TYPE3, 0, userid, content)
}

// ProfitNotice 提取收益消息
func ProfitNotice(coin int64, userid string) (record *pb.LogNotice,
	msg *pb.PushNoticeNtf) {
	if coin <= 0 {
		return
	}
	content := fmt.Sprintf("你成功提取%d收益", coin)
	return NewNotice(data.NOTICE_TYPE3, 0, userid, content)
}

func BuildLanguageMsg(id string, ctype int, param ...string) *pb.LanguageNoticeNtf {
	return &pb.LanguageNoticeNtf{
		Id:    id,
		Ctype: int32(ctype),
		Param: param,
	}
}

// TP跑马灯
func BuildTPMarquee(name, roomName string, score int64) string {
	return fmt.Sprintf("%s won %.2f in %s room", name, float64(score)/100, roomName)
}

// LH跑马灯
func BuildLHMarquee(name string, score int64) string {
	return fmt.Sprintf("%s won %.2f in DRAGON TIGER", name, float64(score)/100)
}

// 7up跑马灯
func BuildUPMarquee(name string, score int64) string {
	return fmt.Sprintf("%s won %.2f in 7 UP DOWN", name, float64(score)/100)
}

// crash跑马灯
func BuildCRASHMarquee(name string, score int64) string {
	return fmt.Sprintf("%s won %.2f in CRASH", name, float64(score)/100)
}

// plane跑马灯
func BuildPLANEMarquee(name string, score int64) string {
	return fmt.Sprintf("%s won %.2f in AVIATOR", name, float64(score)/100)
}

// AB跑马灯
func BuildABMarquee(name string, score int64) string {
	return fmt.Sprintf("%s won %.2f in ANDAR BAHAR", name, float64(score)/100)
}

// 彩票跑马灯
func BuildCPMarquee(name string, score int64) string {
	return fmt.Sprintf("%s won %.2f in LUCKY3PATTI", name, float64(score)/100)
}

// rb跑马灯
func BuildRBMarquee(name string, score int64) string {
	return fmt.Sprintf("%s won %.2f in king VS Queen", name, float64(score)/100)
}

// 提现下单跑马灯
func BuildWithdrawApplyMarquee(name string, amount int64) string {
	return fmt.Sprintf("%s has successfully withdrawn %d！", name, amount)
}

// 提现下单
func BuildWithdrawApply(name string) string {
	return fmt.Sprintf("Dear %s:\nYour withdrawal application is under processing at the bank end. You will receive the notification email once the application is processed and you may check it under your bank card balance within 48 hours. If you do not receive the withdrawal after 48 hours, please contact our customer service.", name)
}

// 提现完成
func BuildWithdrawSuccess(name string) string {
	return fmt.Sprintf("Dear %s: \nYour withdrawal application is processed sucessfully, you may check it in your bank card.", name)
}

// 提现失败
func BuildWithdrawFail(name string) string {
	return fmt.Sprintf("Dear %s:\nYour withdrawal application failed and the chips have been returned to your game account. If you would like to withdraw again, please confirm that your bank card number matches the mobile phone number, email address, and account name and request again.", name)
}

// 提现回退
func BuildWithdrawBack(name string) string {
	return fmt.Sprintf("Dear %s:\nYour withdrawal application is rejected automatically due to network issue or some unknown reasons. Please try again later.", name)
}

// 充值完成
func BuildRechargeFinish(money int64, name string) string {
	return fmt.Sprintf("Dear %s:\nThanks for your purchase. The %.2f chips have been added to your game account, you may check it under your account balance.", name, float32(money)/100)
}

// 后台充值
func BuildGiveCash(money int64, name string) string {
	return fmt.Sprintf("Dear %s:\nThanks very much for your advice, here is %.2f chips as a gift of gratitude. Hope you enjoy.", name, float32(money)/100)
}

// 提现回退（卡号错误）
func BuildWithdrawBackBank(name string) string {
	return fmt.Sprintf("Dear %s:\nYou initiated a withdrawal application at %s. The bank processing failed. We have returned the chips to your game account. Please fill in the following information carefully, bank card number information, mobile phone number, email address, account name Submit the application again after waiting for the letter.", name, utils.LocalTime().Format("2006/01/02 15:04:05"))
}

// 抽小米手机活动
func BuildLuckyDraw(name, numbers string, recharge int64) string {
	return fmt.Sprintf("Dear %s:\nYou purchased %d chips at %s, and successfully participated in the activity of recharging to win a xiaomi14 mobile phone. The lottery number this time is %s.", name, recharge, utils.Time2LocalStr(utils.LocalTime()), numbers)
}

// 抽小米手机活动中奖
func BuildLuckyDrawWin(name string, level, round int, amount int64) string {
	var lv string
	switch level {
	case 1:
		lv = "first"
	case 2:
		lv = "second"
	case 3:
		lv = "third"
	case 4:
		lv = "fourth"
	case 5:
		lv = "fifth"
	case 6:
		lv = "sixth"
	case 7:
		lv = "seventh"
	case 8:
		lv = "eighth"
	case 9:
		lv = "ninth"
	default:
		lv = ""
	}

	return fmt.Sprintf("Dear %s:\nYou participated in the %d phase of the recharge to win xiaomi14 mobile phone at %s. You were lucky and won the %s prize with a bonus of %d rupees. Please claim the reward on the event page.",
		name, round, utils.Time2LocalStr(utils.LocalTime()), lv, amount)
}
