package alert

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	telegramSwitch bool
	// alertorNotify  *notify.Notify
	tgBot *tgbotapi.BotAPI
)

var AlertChatID int64 = -4108600082

// 支付渠道ChatID
var PayChannelChatID map[int32]int64 = map[int32]int64{
	3039: -4679678305,    // DDpay
	3022: -4146426041,    // WEPay
	3018: -4078072162,    // unwinpay
	3030: -4705249946,    // aipay
	3044: -1002639028790, // cxpay
	3028: -4629335564,    // kkPluspay
	3041: -4676274326,    // netpay
	3038: -4768189089,    // mepay
	3029: -4664765258,    // dragonpay
	3010: -765000273,     // xdpay
	3045: -1002629147945, // csmpay
	3015: -4080176728,    // xfpay
	3033: -4667467947,    // 66pay
	3042: -4764826918,    // wyppay
	3040: -1002559894724, // atpay
	3013: -4080336907,    // letspay
	3032: -4608193374,    // gentlepay
	3011: -811411801,     // mlpay
	3034: -1002684671482, // ai2pay
	3035: -1002359774139, // cowpay
	3025: -4535432806,    // blizzardpay
	3026: -4720875306,    // metagopay
	3037: -4718311212,    // paywook
	3043: -1002539904918, // yunpay
	3046: -4843409432,    // wepay2
	3047: -4890329491,    // leopay
	// 孟加拉

	// 巴基斯坦
}

func initTelegram() {
	if env == "dev" {
		telegramSwitch = false
		return
	}
	// telegramService, err1 := telegram.New("6542067431:AAHZsTjUz2VHqm-VRYkaqazI3rVMV5KL7Jg")
	// if err1 != nil {
	// 	panic("new telegram service fail")
	// }
	// telegramService.AddReceivers(-4108600082)
	// alertorNotify = notify.New()
	// alertorNotify.UseServices(telegramService)
	telegramSwitch = cfg.Section("env").Key("alertor").MustBool(false)
	bot, err := tgbotapi.NewBotAPI("6542067431:AAHZsTjUz2VHqm-VRYkaqazI3rVMV5KL7Jg")
	if err != nil {
		log.Fatalf("初始化Bot失败: %s", err)
	}
	tgBot = bot
}
