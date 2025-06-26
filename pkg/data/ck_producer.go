package data

// var (
// 	TopicUser                                 = "col_user"
// 	TopicUserFinance                          = "col_user_finance"
// 	TopicTradeRecord                          = "col_trade_record"
// 	TopicWithdrawRecord                       = "col_withdraw_record"
// 	TopicWithdrawRecord_transfer_order_detail = "col_withdraw_record_transfer_order_detail"
// 	TopicLogWater                             = "col_log_water"
// 	TopicDetail                               = "col_detail"
// 	TopicLogLogin                             = "col_log_login"
// 	TopicExternalBet                          = "col_nsq_log_external_bet"
// 	TopicExternalReward                       = "col_nsq_log_external_reward"
// 	TopicExternalCancel                       = "col_nsq_log_external_cancel"
// 	TopicActivityBetRankPrize                 = "col_activity_betrank_prize"
// 	TopicActivityTurnDrawLog                  = "col_activity_turn_draw_log"
// 	TopicActivityTurnPrizeLog                 = "col_activity_turn_prize_log"
// 	TopicLaunchInviteLog                      = "col_launch_invite_log"
// 	TopicShareAgentIncomeRecord               = "col_activity_share_income_record"
// 	TopicShareAgentIncomeRecordTackLog        = "col_activity_share_income_record_tack_log"
// )

var DataProducers DataProducer = new(DefaultDataProducer)

type DataProducer interface {
	PublishDatas(topic string, datas []interface{})
}

type DefaultDataProducer struct {
}

func (DefaultDataProducer) PublishDatas(topic string, datas []interface{}) {}
