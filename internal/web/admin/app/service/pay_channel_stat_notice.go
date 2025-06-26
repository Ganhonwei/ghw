package service

import (
	"context"
	"fmt"
	"goserver/pkg/data/mq"
	"goserver/pkg/utils"
	"runtime/debug"
	"time"

	"github.com/astaxie/beego"
	"github.com/globalsign/mgo/bson"
	"github.com/nikoksr/notify"
	"github.com/nikoksr/notify/service/telegram"
	"github.com/scylladb/termtables"
)

var (
	alertorNotify *notify.Notify
)

func initTelegramNotify() {
	// https://api.telegram.org/bot<YourBOTToken>/getUpdates
	telegramService, err1 := telegram.New("7858344970:AAG-gl0Y8IR4fT0rMS0qIk3UP8F2fFXSde4")
	if err1 != nil {
		panic("new telegram service fail")
	}
	telegramService.AddReceivers(-1002610083923)
	alertorNotify = notify.New()
	alertorNotify.UseServices(telegramService)
}

// 监听订单提单失败消息
func subscribePayOrderError() {
	mq.NatsCreateConsumer("web", mq.StreamGame, mq.TopicPayOrderError, func() *mq.PublishPayOrderError { return new(mq.PublishPayOrderError) },
		func(arg *mq.PublishPayOrderError) (err error) {
			msg := fmt.Sprintf("支付订单提单失败:通道【%d-%s】,订单号【%s】,三方订单号【%s】,订单金额【%.2f】,用户id【%s】, 【失败原因】:%s, 【通道响应】:%s",
				arg.ChannelId, arg.ChannelName, arg.MerOrderId, arg.OutTradeNo, float64(arg.Amount)/100, arg.Userid, arg.Err, arg.Response)

			err = alertorNotify.Send(context.Background(), "", msg)
			if err != nil {
				beego.Error("payChannelStat tg notice error: ", err)
			}
			return
		})
}

// 支付渠道成功率定时统计tg统计
func tickPayChannelStatNotice() {
	ticker := time.NewTicker(30 * time.Second)
	for {
		_, ok := <-ticker.C
		if !ok {
			break
		}

		func() {
			defer func() {
				if err := recover(); err != nil {
					beego.Error("payChannelStat error: ", err)
					debug.PrintStack()
				}
			}()

			now := NowTime()
			nowSec := now.Unix()
			cfg := PayService.GetPayChannelStatNoticeConfig()

			if (cfg.DayTicker > 0 || cfg.TimeTicker > 0) && // 开关打开
				(nowSec >= cfg.DayNextTime && nowSec >= cfg.TimeNextTime) { // 触发时间到
				timeBegin, timeEnd, prevTimeBegin, prevTimeEnd, prevDayTimeBegin, prevDayTimeEnd,
					dayBegin, dayEnd, prevDayBegin, prevDayEnd,
					withdrawTimeStats, payTimeStats, prevWithdrawTimeStats, prevPayTimeStats, prevDayWithdrawTimeStats, prevDayPayTimeStats,
					withdrawDayStats, payDayStats, prevDayWithdrawStats, prevDayPayStats, err := PayService.GetPayChannelStat()

				if err != nil {
					beego.Error("payChannelStat error: ", err)
					return
				}
				_, _, _, _ = prevTimeBegin, prevTimeEnd, prevDayTimeBegin, prevDayTimeEnd
				_, _ = prevDayBegin, prevDayEnd

				// 时段统计通知
				if cfg.TimeTicker > 0 && nowSec > cfg.TimeNextTime {
					cfg.TimeNextTime = nowSec + (cfg.TimeTicker * 60)

					// 时段代收表
					table := termtables.CreateTable()
					table.SetAlign(termtables.AlignCenter, 1, 2, 3, 4, 5)
					table.SetModeMarkdown()
					table.AddHeaders("渠道", "发起数", "成功数", "成功率", "报错数")
					for _, stat := range payTimeStats {
						table.AddRow(stat.ChannelName, stat.PayTotalTimes, stat.PaySuccessTimes, stat.PaySuccessRate, stat.PayErrorTimes)
					}
					t := fmt.Sprintf("%s ~ %s", timeBegin.Format(utils.FORMAT_TIME_Minute), timeEnd.Format(utils.FORMAT_TIME_Minute))
					msg := fmt.Sprintf("***代收通道-时段统计***\n%s\n%s", t, table.Render())
					fmt.Println(msg)
					err = alertorNotify.Send(context.Background(), "", msg)
					if err != nil {
						beego.Error("payChannelStat tg notice error: ", err)
					}

					// 时段代付表
					table = termtables.CreateTable()
					table.SetAlign(termtables.AlignCenter, 1, 2, 3, 4, 5)
					table.SetModeMarkdown()
					table.AddHeaders("渠道", "发起数", "成功数", "成功率", "报错数")
					for _, stat := range withdrawTimeStats {
						table.AddRow(stat.ChannelName, stat.WithdrawTotalTimes, stat.WithdrawSuccessTimes, stat.WithdrawSuccessRate, stat.WithdrawErrorTimes)
					}
					t = fmt.Sprintf("%s ~ %s", timeBegin.Format(utils.FORMAT_TIME_Minute), timeEnd.Format(utils.FORMAT_TIME_Minute))
					msg = fmt.Sprintf("***代付通道-时段统计***\n%s\n%s", t, table.Render())
					fmt.Println(msg)
					err = alertorNotify.Send(context.Background(), "", msg)
					if err != nil {
						beego.Error("payChannelStat tg notice error: ", err)
					}

					// ----代收统计评论
					var paySuccessRateF, payStatF, prevDay_paySuccessRateDiffF, prev_PaySuccessRateDiffF, day_paySuccessRateF, day_payStatF, dayprev_payStatDiffF string
					var payTotalTimes, paySuccessTimes int64
					var paySuccessRate float64
					for _, stat := range payTimeStats {
						payTotalTimes += stat.PayTotalTimes
						paySuccessTimes += stat.PaySuccessTimes
					}
					if payTotalTimes > 0 {
						paySuccessRate = float64(paySuccessTimes) / float64(payTotalTimes)
						paySuccessRateF = fmt.Sprintf("%.2f%%", paySuccessRate*100)
						payStatF = fmt.Sprintf("%d/%d", paySuccessTimes, payTotalTimes)
					}

					// 同比昨日
					var prevDay_payTotalTimes, prevDay_paySuccessTimes int64
					var prevDay_paySuccessRate float64
					for _, stat := range prevDayPayTimeStats {
						prevDay_payTotalTimes += stat.PayTotalTimes
						prevDay_paySuccessTimes += stat.PaySuccessTimes
					}
					if prevDay_payTotalTimes > 0 {
						prevDay_paySuccessRate = float64(prevDay_paySuccessTimes) / float64(prevDay_payTotalTimes)
					}
					prevDay_paySuccessRateDiffF = fmt.Sprintf("%.2f%%", (paySuccessRate-prevDay_paySuccessRate)*100)
					// 环比上一时段
					var prev_payTotalTimes, prev_paySuccessTimes int64
					var prev_paySuccessRate float64
					for _, stat := range prevPayTimeStats {
						prev_payTotalTimes += stat.PayTotalTimes
						prev_paySuccessTimes += stat.PaySuccessTimes
					}
					if prev_payTotalTimes > 0 {
						prev_paySuccessRate = float64(prev_paySuccessTimes) / float64(prev_payTotalTimes)
					}
					prev_PaySuccessRateDiffF = fmt.Sprintf("%.2f%%", (paySuccessRate-prev_paySuccessRate)*100)
					// 今日
					var day_payTotalTimes, day_paySuccessTimes int64
					var day_paySuccessRate float64
					for _, stat := range payDayStats {
						day_payTotalTimes += stat.PayTotalTimes
						day_paySuccessTimes += stat.PaySuccessTimes
					}
					if day_payTotalTimes > 0 {
						day_paySuccessRate = float64(day_paySuccessTimes) / float64(day_payTotalTimes)
						day_paySuccessRateF = fmt.Sprintf("%.2f%%", day_paySuccessRate*100)
						day_payStatF = fmt.Sprintf("%d/%d", day_paySuccessTimes, day_payTotalTimes)
					}
					// 今日环比昨日
					var dayprev_payTotalTimes, dayprev_paySuccessTimes int64
					var dayprev_paySuccessRate float64
					for _, stat := range prevDayPayStats {
						dayprev_payTotalTimes += stat.PayTotalTimes
						dayprev_paySuccessTimes += stat.PaySuccessTimes
					}
					if dayprev_payTotalTimes > 0 {
						dayprev_paySuccessRate = float64(dayprev_paySuccessTimes) / float64(dayprev_payTotalTimes)
					}
					dayprev_payStatDiffF = fmt.Sprintf("%.2f%%", (day_paySuccessRate-dayprev_paySuccessRate)*100)

					msg = fmt.Sprintf("【充值成功预警】订单成功率为【%s】，订单数【%s】，同比%s，环比%s；今日平均订单成功率为【%s】，订单数【%s】，环比昨日【%s】",
						paySuccessRateF, payStatF, prevDay_paySuccessRateDiffF, prev_PaySuccessRateDiffF, day_paySuccessRateF, day_payStatF, dayprev_payStatDiffF)
					err = alertorNotify.Send(context.Background(), "", msg)
					if err != nil {
						beego.Error("payChannelStat tg notice error: ", err)
					}

					// ----代付统计评论
					var withdrawSuccessRateF, withdrawStatF, prevDay_withdrawSuccessRateDiffF, prev_WithdrawSuccessRateDiffF, day_withdrawSuccessRateF, day_withdrawStatF, dayprev_withdrawStatDiffF string
					var withdrawTotalTimes, withdrawSuccessTimes int64
					var withdrawSuccessRate float64
					for _, stat := range withdrawTimeStats {
						withdrawTotalTimes += stat.WithdrawTotalTimes
						withdrawSuccessTimes += stat.WithdrawSuccessTimes
					}
					if withdrawTotalTimes > 0 {
						withdrawSuccessRate = float64(withdrawSuccessTimes) / float64(withdrawTotalTimes)
						withdrawSuccessRateF = fmt.Sprintf("%.2f%%", withdrawSuccessRate*100)
						withdrawStatF = fmt.Sprintf("%d/%d", withdrawSuccessTimes, withdrawTotalTimes)
					}

					// 同比昨日
					var prevDay_withdrawTotalTimes, prevDay_withdrawSuccessTimes int64
					var prevDay_withdrawSuccessRate float64
					for _, stat := range prevDayWithdrawTimeStats {
						prevDay_withdrawTotalTimes += stat.WithdrawTotalTimes
						prevDay_withdrawSuccessTimes += stat.WithdrawSuccessTimes
					}
					if prevDay_withdrawTotalTimes > 0 {
						prevDay_withdrawSuccessRate = float64(prevDay_withdrawSuccessTimes) / float64(prevDay_withdrawTotalTimes)
					}
					prevDay_withdrawSuccessRateDiffF = fmt.Sprintf("%.2f%%", (withdrawSuccessRate-prevDay_withdrawSuccessRate)*100)
					// 环比上一时段
					var prev_withdrawTotalTimes, prev_withdrawSuccessTimes int64
					var prev_withdrawSuccessRate float64
					for _, stat := range prevWithdrawTimeStats {
						prev_withdrawTotalTimes += stat.WithdrawTotalTimes
						prev_withdrawSuccessTimes += stat.WithdrawSuccessTimes
					}
					if prev_withdrawTotalTimes > 0 {
						prev_withdrawSuccessRate = float64(prev_withdrawSuccessTimes) / float64(prev_withdrawTotalTimes)
					}
					prev_WithdrawSuccessRateDiffF = fmt.Sprintf("%.2f%%", (withdrawSuccessRate-prev_withdrawSuccessRate)*100)
					// 今日
					var day_withdrawTotalTimes, day_withdrawSuccessTimes int64
					var day_withdrawSuccessRate float64
					for _, stat := range withdrawDayStats {
						day_withdrawTotalTimes += stat.WithdrawTotalTimes
						day_withdrawSuccessTimes += stat.WithdrawSuccessTimes
					}
					if day_withdrawTotalTimes > 0 {
						day_withdrawSuccessRate = float64(day_withdrawSuccessTimes) / float64(day_withdrawTotalTimes)
						day_withdrawSuccessRateF = fmt.Sprintf("%.2f%%", day_withdrawSuccessRate*100)
						day_withdrawStatF = fmt.Sprintf("%d/%d", day_withdrawSuccessTimes, day_withdrawTotalTimes)
					}
					// 今日环比昨日
					var dayprev_withdrawTotalTimes, dayprev_withdrawSuccessTimes int64
					var dayprev_withdrawSuccessRate float64
					for _, stat := range prevDayWithdrawStats {
						dayprev_withdrawTotalTimes += stat.WithdrawTotalTimes
						dayprev_withdrawSuccessTimes += stat.WithdrawSuccessTimes
					}
					if dayprev_withdrawTotalTimes > 0 {
						dayprev_withdrawSuccessRate = float64(dayprev_withdrawSuccessTimes) / float64(dayprev_withdrawTotalTimes)
					}
					dayprev_withdrawStatDiffF = fmt.Sprintf("%.2f%%", (day_withdrawSuccessRate-dayprev_withdrawSuccessRate)*100)

					msg = fmt.Sprintf("【提现成功预警】订单成功率为【%s】，订单数【%s】，同比%s，环比%s；今日平均订单成功率为【%s】，订单数【%s】，环比昨日【%s】",
						withdrawSuccessRateF, withdrawStatF, prevDay_withdrawSuccessRateDiffF, prev_WithdrawSuccessRateDiffF, day_withdrawSuccessRateF, day_withdrawStatF, dayprev_withdrawStatDiffF)
					err = alertorNotify.Send(context.Background(), "", msg)
					if err != nil {
						beego.Error("withdrawChannelStat tg notice error: ", err)
					}
				}

				// ------ 当天统计
				if cfg.DayTicker > 0 && nowSec > cfg.DayNextTime {
					cfg.DayNextTime = nowSec + (cfg.DayTicker * 60)

					// 当天代收表
					table := termtables.CreateTable()
					table.SetAlign(termtables.AlignCenter, 1, 2, 3, 4, 5)
					table.SetModeMarkdown()
					table.AddHeaders("渠道", "发起数", "成功数", "成功率", "报错数")
					for _, stat := range payDayStats {
						table.AddRow(stat.ChannelName, stat.PayTotalTimes, stat.PaySuccessTimes, stat.PaySuccessRate, stat.PayErrorTimes)
					}
					t := fmt.Sprintf("%s ~ %s", dayBegin.Format(utils.FORMAT_TIME_Minute), dayEnd.Format(utils.FORMAT_TIME_Minute))
					msg := fmt.Sprintf("***代收通道-当天统计***\n%s\n%s", t, table.Render())
					// fmt.Println(msg)
					err = alertorNotify.Send(context.Background(), "", msg)
					if err != nil {
						beego.Error("payChannelStat tg notice error: ", err)
					}

					// 当天代付表
					table = termtables.CreateTable()
					table.SetAlign(termtables.AlignCenter, 1, 2, 3, 4, 5)
					table.SetModeMarkdown()
					table.AddHeaders("渠道", "发起数", "成功数", "成功率", "报错数")
					for _, stat := range withdrawDayStats {
						table.AddRow(stat.ChannelName, stat.WithdrawTotalTimes, stat.WithdrawSuccessTimes, stat.WithdrawSuccessRate, stat.WithdrawErrorTimes)
					}
					t = fmt.Sprintf("%s ~ %s", dayBegin.Format(utils.FORMAT_TIME_Minute), dayEnd.Format(utils.FORMAT_TIME_Minute))
					msg = fmt.Sprintf("***代付通道-当天统计***\n%s\n%s", t, table.Render())
					// fmt.Println(msg)
					err = alertorNotify.Send(context.Background(), "", msg)
					if err != nil {
						beego.Error("payChannelStat tg notice error: ", err)
					}
				}

				Update(PayChannelStatNoticeConfigs, bson.M{"_id": "1"}, bson.M{"$set": bson.M{
					"day_next_time":  cfg.DayNextTime,
					"time_next_time": cfg.TimeNextTime,
				}})
			}

			// 提现提醒触发
			if cfg.WithdrawTicker > 0 && nowSec >= cfg.WithdrawNextTime {
				cfg.WithdrawNextTime = nowSec + (cfg.WithdrawTicker * 60)
				count, amountSum, err := PayService.WithdrawTaggedStat()
				if err != nil {
					beego.Error("WithdrawTaggedStat error: ", err)
				} else if count > 0 {
					msg := fmt.Sprintf("提现预警，有【%d】笔订单未处理, 总计金额【%.2f】", count, amountSum)
					err = alertorNotify.Send(context.Background(), "", msg)
					if err != nil {
						beego.Error("payChannelStat tg notice error: ", err)
					}
				}

				Update(PayChannelStatNoticeConfigs, bson.M{"_id": "1"}, bson.M{"$set": bson.M{
					"withdraw_next_time": cfg.WithdrawNextTime,
				}})
			}

			// 支付订单UTR补分通知
			if cfg.PayUtrTicker > 0 && nowSec >= cfg.PayUtrNextTime {
				cfg.PayUtrNextTime = nowSec + (cfg.PayUtrTicker * 60)
				count, amountSum, err := PayService.PayUtrStat()
				if err != nil {
					beego.Error("PayUtrStat error: ", err)
				} else if count > 0 {
					msg := fmt.Sprintf("支付订单UTR补分预警，有【%d】笔订单未处理, 总计金额【%.2f】", count, amountSum)
					err = alertorNotify.Send(context.Background(), "", msg)
					if err != nil {
						beego.Error("PayUtrStat tg notice error: ", err)
					}
				}

				Update(PayChannelStatNoticeConfigs, bson.M{"_id": "1"}, bson.M{"$set": bson.M{
					"pay_utr_next_time": cfg.PayUtrNextTime,
				}})
			}
		}()
	}
}
