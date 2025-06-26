package alert

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/data/mq"
	"goserver/pkg/glog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func initNatsConsumer() {
	consumerName := "alert"

	inst_AlertorMail := func() *pb.AlertorMail { return &pb.AlertorMail{} }
	mq.NatsCreateConsumer(consumerName, mq.StreamAlert, mq.TopicAlertEmail, inst_AlertorMail, func(data *pb.AlertorMail) (err error) {
		if !telegramSwitch {
			// 开发模式不发
			return nil
		}

		glog.Debugf("receive event: %v", data)

		content := fmt.Sprintf("%s\n%s", data.Subject, data.Message)

		msg := tgbotapi.NewMessage(AlertChatID, content)
		_, err = tgBot.Send(msg)
		if err != nil {
			glog.Infof("send %s telegram msg fail %v", AlertChatID, err)
			return
		}

		glog.Infof("send %s telegram msg to %s", data.Subject, AlertChatID)

		return nil
	})

	// UTR补单订阅
	mq.NatsCreateConsumer(consumerName, mq.StreamUtr, mq.TopicFillOrder, func() *pb.UtrFillOrder { return new(pb.UtrFillOrder) },
		func(f *pb.UtrFillOrder) (err error) {
			if !telegramSwitch {
				// 开发模式不发
				return nil
			}

			payOrder := data.TradeRecord{OrderID: f.OrderId}
			payOrder.Get()
			glog.Infof("receive event: %v", consumerName)

			content := fmt.Sprintf("\n%s", payOrder.OutTradeNo)
			chatId, ok := PayChannelChatID[f.ChannelId]
			if !ok {
				chatId = AlertChatID
			}
			msg := tgbotapi.NewPhotoShare(chatId, f.ImageUrl)
			msg.Caption = content
			_, err = tgBot.Send(msg)
			if err != nil {
				glog.Infof("send %d telegram msg fail %v", f.ChannelId, err)
				return
			}
			// @客服
			notify := tgbotapi.NewMessage(chatId, "@shreyaali @mayaya1128 @david4780 @EdwardNova0 @BigWinProgram")
			_, err = tgBot.Send(notify)
			if err != nil {
				glog.Infof("send %d telegram msg fail %v", f.ChannelId, err)
				return
			}
			return
		})
}
