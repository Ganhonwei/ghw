package data

import (
	"goserver/pkg/glog"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// Init mgo and the common DAO

// 数据连接
var Session *mgo.Session
var mgoCloseCh chan bool

// 各个表的Collection对象
var TradeRecords *mgo.Collection
var PlayerUsers *mgo.Collection
var Stocks *mgo.Collection
var GameStocks *mgo.Collection
var NewbieStocks *mgo.Collection
var Details *mgo.Collection
var GameRecords *mgo.Collection
var UserRecords *mgo.Collection
var StatRecords *mgo.Collection
var UserStatRecords *mgo.Collection
var TPStoryChargeStocks *mgo.Collection

var IDGens *mgo.Collection
var WxLogins *mgo.Collection
var LogDiamonds *mgo.Collection
var LogCoins *mgo.Collection
var LogRegists *mgo.Collection
var LogLogins *mgo.Collection
var LogBuildAgencys *mgo.Collection
var LogOnlines *mgo.Collection
var LogCards *mgo.Collection
var LogChips *mgo.Collection
var LogExpects *mgo.Collection
var LogWaters *mgo.Collection
var LogOutDiamonds *mgo.Collection
var LogGiveDiamonds *mgo.Collection
var LogVBDiamonds *mgo.Collection
var LogShareWaters *mgo.Collection
var LogEventTracks *mgo.Collection
var LogGameTimes *mgo.Collection
var LogBugFeedbacks *mgo.Collection
var LogButtonClicks *mgo.Collection
var LogButton2Clicks *mgo.Collection
var LogADReports *mgo.Collection
var LogFBReports *mgo.Collection
var NsqLogExternalBets *mgo.Collection
var NsqLogExternalRewards *mgo.Collection
var NsqLogExternalCancels *mgo.Collection
var NsqLogEfiTransactions *mgo.Collection
var LogRechargeReports *mgo.Collection
var CustomerAdresses *mgo.Collection
var Cdks *mgo.Collection
var ShareWays *mgo.Collection
var LogSmsRecords *mgo.Collection
var LogUserSmsRecords *mgo.Collection
var CashFlows *mgo.Collection
var LogScratchTickets *mgo.Collection
var LogActivityMonitors *mgo.Collection
var LogPlayShareDraws *mgo.Collection
var LogBonuss *mgo.Collection
var LogShareRegists *mgo.Collection
var LogGameFlowWaters *mgo.Collection
var UploadUserHeads *mgo.Collection
var JHVFRecords *mgo.Collection
var ShareAgentIncomeRecordTackLogs *mgo.Collection
var GiftPackCodes *mgo.Collection
var GiftPackCodeTackLogs *mgo.Collection
var VolatilitySubsidyPools *mgo.Collection

var PayBlackLists *mgo.Collection

var Trends *mgo.Collection
var Pk10Records *mgo.Collection
var RoomRecords *mgo.Collection
var RoleRecords *mgo.Collection
var RoundRecords *mgo.Collection

var Agencys *mgo.Collection
var UserInfos *mgo.Collection

var Notices *mgo.Collection
var Shops *mgo.Collection
var Envs *mgo.Collection
var Games *mgo.Collection
var Vips *mgo.Collection
var Tasks *mgo.Collection
var PokerHandss *mgo.Collection
var Activitys *mgo.Collection
var DailySigns *mgo.Collection
var OnlineRewards *mgo.Collection
var LoginPrizes *mgo.Collection
var LogTasks *mgo.Collection
var LogProfits *mgo.Collection
var LogProfitsOrders *mgo.Collection
var LogBanks *mgo.Collection
var LogSysProfits *mgo.Collection
var Luckys *mgo.Collection
var LogDayProfits *mgo.Collection
var LogActivitys *mgo.Collection
var EquipmentBlacklists *mgo.Collection
var TPStoryChargeStockHistoryLogs *mgo.Collection

var SetWithdraws *mgo.Collection
var WithdrawRecords *mgo.Collection
var SystemSwitchs *mgo.Collection
var ChatLogs *mgo.Collection
var PayChannels *mgo.Collection
var Shares *mgo.Collection
var AdjustCallback *mgo.Collection
var ClientLogs *mgo.Collection
var Beginners *mgo.Collection
var RoomPeoples *mgo.Collection
var IPwhites *mgo.Collection
var WeeklyCards *mgo.Collection
var Emojis *mgo.Collection
var ShareAmounts *mgo.Collection
var Uids *mgo.Collection
var Channels *mgo.Collection
var ADJusts *mgo.Collection
var Banners *mgo.Collection
var BonusBanners *mgo.Collection
var ServerWhites *mgo.Collection
var LimitedGifts *mgo.Collection
var ShareAddrs *mgo.Collection
var PvpRooms *mgo.Collection
var FBReports *mgo.Collection
var RankWithdraws *mgo.Collection
var VipRobots *mgo.Collection
var LuckyDrawRewards *mgo.Collection
var LuckyDrawRobots *mgo.Collection
var LuckDrawNumbers *mgo.Collection
var LuckyDrawGives *mgo.Collection
var ScratchTicketss *mgo.Collection
var MarqueeWithdraws *mgo.Collection
var UserGameDatas *mgo.Collection
var PlayAndDrawActivitys *mgo.Collection
var ActivityBetRankPrizes *mgo.Collection
var CrashStrategies *mgo.Collection
var ChargeClassifys *mgo.Collection
var LHDStrategies *mgo.Collection
var SevenStrategies *mgo.Collection
var ABStrategies *mgo.Collection
var CPStrategies *mgo.Collection
var RBStrategies *mgo.Collection
var PlaneStrategies *mgo.Collection
var ActivityTurnTimes *mgo.Collection
var ActivityTurnDrawLogs *mgo.Collection
var ActivityTurnPrizeLogs *mgo.Collection
var ShareAgentIncomeRecords *mgo.Collection
var VolatilitySubsidys *mgo.Collection
var OnlineUsersLogs *mgo.Collection
var Utrs *mgo.Collection
var UserCoupons *mgo.Collection
var ManualCoupons *mgo.Collection
var CumRechargeWheelUnlockLogs *mgo.Collection
var CumRechargeWheelSpinLogs *mgo.Collection

// 初始化时连接数据库
func InitMgo(dbHost, dbPort, dbUser, dbPassword, dbName string) {
	// get db config from host, port, username, password
	usernameAndPassword := dbUser + ":" + dbPassword + "@"
	if dbUser == "" || dbPassword == "" {
		usernameAndPassword = ""
	}
	if dbPort == "" {
		dbPort = "27017"
	}
	authString := ""
	if usernameAndPassword != "" {
		authString = "?authSource=admin&authMechanism=SCRAM-SHA-1"
	}
	url := "mongodb://" + usernameAndPassword + dbHost + ":" + dbPort + "/" + dbName + authString

	// [mongodb://][user:pass@]host1[:port1][,host2[:port2],...][/database][?options]
	// mongodb://myuser:mypass@localhost:40001,otherhost:40001/mydb
	var err error
	Session, err = mgo.Dial(url)
	if err != nil {
		//return //TODO test
		glog.Error(err)
		panic(err)
	}
	go ticker()

	// Optional. Switch the session to a monotonic behavior.
	Session.SetMode(mgo.Monotonic, true)

	// trade_record
	TradeRecords = Session.DB(dbName).C("col_trade_record")
	// user
	PlayerUsers = Session.DB(dbName).C("col_user")
	IDGens = Session.DB(dbName).C("col_id_gen")
	// stock
	Stocks = Session.DB(dbName).C("col_stock")
	GameStocks = Session.DB(dbName).C("col_game_stock")
	NewbieStocks = Session.DB(dbName).C("col_newbie_stock")
	// detail
	// Details = Session.DB(dbName).C("col_detail")
	// record
	GameRecords = Session.DB(dbName).C("col_game_record")
	UserRecords = Session.DB(dbName).C("col_user_record")
	StatRecords = Session.DB(dbName).C("col_stat_record")
	UserStatRecords = Session.DB(dbName).C("col_user_stat_record")
	TPStoryChargeStocks = Session.DB(dbName).C("col_tp_story_charge_stock")

	WxLogins = Session.DB(dbName).C("col_wx_login")
	// LogDiamonds = Session.DB(dbName).C("col_log_diamond")
	// LogCoins = Session.DB(dbName).C("col_log_coin")
	// LogRegists = Session.DB(dbName).C("col_log_regist")
	// LogLogins = Session.DB(dbName).C("col_log_login")
	// LogBuildAgencys = Session.DB(dbName).C("col_log_build_agency")
	// LogOnlines = Session.DB(dbName).C("col_log_online")
	// LogCards = Session.DB(dbName).C("col_log_card")
	// LogChips = Session.DB(dbName).C("col_log_chip")
	// LogExpects = Session.DB(dbName).C("col_log_expect")
	// LogWaters = Session.DB(dbName).C("col_log_water")
	Trends = Session.DB(dbName).C("col_trends")
	Pk10Records = Session.DB(dbName).C("col_pkten")
	RoomRecords = Session.DB(dbName).C("col_room_record")
	RoleRecords = Session.DB(dbName).C("col_role_record")
	RoundRecords = Session.DB(dbName).C("col_round_record")
	RoomPeoples = Session.DB(dbName).C("col_room_people")
	//
	Agencys = Session.DB(dbName).C("t_user")
	UserInfos = Session.DB(dbName).C("col_user_info")
	//
	Notices = Session.DB(dbName).C("col_notice")
	Shops = Session.DB(dbName).C("col_shop")
	Envs = Session.DB(dbName).C("col_env")
	Vips = Session.DB(dbName).C("col_vip")
	Games = Session.DB(dbName).C("col_game")
	Tasks = Session.DB(dbName).C("col_task")
	Activitys = Session.DB(dbName).C("col_activity")
	DailySigns = Session.DB(dbName).C("col_dailysign")
	OnlineRewards = Session.DB(dbName).C("col_online_reward")
	LoginPrizes = Session.DB(dbName).C("col_login_prize")
	// LogTasks = Session.DB(dbName).C("col_log_task")
	// LogProfits = Session.DB(dbName).C("col_log_profit")
	// LogProfitsOrders = Session.DB(dbName).C("col_log_profit_order")
	// LogBanks = Session.DB(dbName).C("col_log_bank")
	// LogSysProfits = Session.DB(dbName).C("col_log_sys_profit")
	// Luckys = Session.DB(dbName).C("col_lucky")
	// LogDayProfits = Session.DB(dbName).C("col_log_day_profit")
	// LogActivitys = Session.DB(dbName).C("col_log_activity")
	SetWithdraws = Session.DB(dbName).C("col_set_withdraw")
	WithdrawRecords = Session.DB(dbName).C("col_withdraw_record")
	PokerHandss = Session.DB(dbName).C("col_pokerhands")
	SystemSwitchs = Session.DB(dbName).C("col_set_game_system")
	PayChannels = Session.DB(dbName).C("col_pay_channel")
	Shares = Session.DB(dbName).C("col_share")
	AdjustCallback = Session.DB(dbName).C("col_adjust_callback")
	Beginners = Session.DB(dbName).C("col_beginner")
	WeeklyCards = Session.DB(dbName).C("col_weekly_card")
	Emojis = Session.DB(dbName).C("col_emoji")
	ShareAmounts = Session.DB(dbName).C("col_share_amount")
	Uids = Session.DB(dbName).C("col_uid")
	Channels = Session.DB(dbName).C("col_channel")
	ADJusts = Session.DB(dbName).C("col_adjust")
	LimitedGifts = Session.DB(dbName).C("col_limited_gift")
	ShareAddrs = Session.DB(dbName).C("col_share_addr")
	PvpRooms = Session.DB(dbName).C("col_pvp_room")
	FBReports = Session.DB(dbName).C("col_fb_report")
	RankWithdraws = Session.DB(dbName).C("col_rank_withdraw")
	VipRobots = Session.DB(dbName).C("col_vip_robot")
	LuckyDrawRewards = Session.DB(dbName).C("col_lucky_draw_reward")
	LuckyDrawRobots = Session.DB(dbName).C("col_lucky_draw_robot")
	LuckDrawNumbers = Session.DB(dbName).C("col_lucky_draw_numbers")
	LuckyDrawGives = Session.DB(dbName).C("col_lucky_draw_gives")
	ScratchTicketss = Session.DB(dbName).C("col_scratch_tickets")
	MarqueeWithdraws = Session.DB(dbName).C("col_marquee_withdraws")
	PlayAndDrawActivitys = Session.DB(dbName).C("col_play_share")
	ActivityBetRankPrizes = Session.DB(dbName).C("col_activity_betrank_prize")
	CrashStrategies = Session.DB(dbName).C("col_crash_strategy")
	ChargeClassifys = Session.DB(dbName).C("col_charge_classify")
	LHDStrategies = Session.DB(dbName).C("col_lhd_strategy")
	SevenStrategies = Session.DB(dbName).C("col_seven_strategy")
	ABStrategies = Session.DB(dbName).C("col_ab_strategy")
	CPStrategies = Session.DB(dbName).C("col_cp_strategy")
	RBStrategies = Session.DB(dbName).C("col_rb_strategy")
	PlaneStrategies = Session.DB(dbName).C("col_plane_strategy")
	JHVFRecords = Session.DB(dbName).C("col_jh_vf_record")
	ShareAgentIncomeRecordTackLogs = Session.DB(dbName).C("col_activity_share_income_record_tack_log")
	GiftPackCodes = Session.DB(dbName).C("col_activity_gift_pack_code")
	GiftPackCodeTackLogs = Session.DB(dbName).C("col_activity_gift_pack_code_tack_log")
	VolatilitySubsidyPools = Session.DB(dbName).C("col_activity_volatility_subsidy_pool")
	UserCoupons = Session.DB(dbName).C("col_user_coupon")
	ManualCoupons = Session.DB(dbName).C("col_manual_coupon")
}

// 初始化时连接数据库
func InitLogMgo(dbHost, dbPort, dbUser, dbPassword, dbName string) {
	// get db config from host, port, username, password
	usernameAndPassword := dbUser + ":" + dbPassword + "@"
	if dbUser == "" || dbPassword == "" {
		usernameAndPassword = ""
	}
	if dbPort == "" {
		dbPort = "27017"
	}
	authString := ""
	if usernameAndPassword != "" {
		authString = "?authSource=admin&authMechanism=SCRAM-SHA-1"
	}
	url := "mongodb://" + usernameAndPassword + dbHost + ":" + dbPort + "/" + dbName + authString

	// [mongodb://][user:pass@]host1[:port1][,host2[:port2],...][/database][?options]
	// mongodb://myuser:mypass@localhost:40001,otherhost:40001/mydb
	var err error
	Session, err = mgo.Dial(url)
	if err != nil {
		//return //TODO test
		panic(err)
	}
	go ticker()

	// Optional. Switch the session to a monotonic behavior.
	Session.SetMode(mgo.Monotonic, true)

	// detail
	Details = Session.DB(dbName).C("col_detail")

	LogDiamonds = Session.DB(dbName).C("col_log_diamond")
	LogCoins = Session.DB(dbName).C("col_log_coin")
	LogRegists = Session.DB(dbName).C("col_log_regist")
	LogLogins = Session.DB(dbName).C("col_log_login")
	LogBuildAgencys = Session.DB(dbName).C("col_log_build_agency")
	LogOnlines = Session.DB(dbName).C("col_log_online")
	LogCards = Session.DB(dbName).C("col_log_card")
	LogChips = Session.DB(dbName).C("col_log_chip")
	LogExpects = Session.DB(dbName).C("col_log_expect")
	LogWaters = Session.DB(dbName).C("col_log_water")
	IPwhites = Session.DB(dbName).C("col_ip_white")
	PayBlackLists = Session.DB(dbName).C("col_pay_blacklist")

	//
	LogTasks = Session.DB(dbName).C("col_log_task")
	LogProfits = Session.DB(dbName).C("col_log_profit")
	LogProfitsOrders = Session.DB(dbName).C("col_log_profit_order")
	LogBanks = Session.DB(dbName).C("col_log_bank")
	LogSysProfits = Session.DB(dbName).C("col_log_sys_profit")
	Luckys = Session.DB(dbName).C("col_lucky")
	LogDayProfits = Session.DB(dbName).C("col_log_day_profit")
	LogActivitys = Session.DB(dbName).C("col_log_activity")
	LogOutDiamonds = Session.DB(dbName).C("col_log_outdiamond")
	LogGiveDiamonds = Session.DB(dbName).C("col_log_givediamond")
	LogVBDiamonds = Session.DB(dbName).C("col_log_vbdiamond")
	LogShareWaters = Session.DB(dbName).C("col_share_water")
	LogEventTracks = Session.DB(dbName).C("col_event_track")
	LogGameTimes = Session.DB(dbName).C("col_game_time")
	LogBugFeedbacks = Session.DB(dbName).C("col_bug_feedback")
	LogButtonClicks = Session.DB(dbName).C("col_button_click")
	LogButton2Clicks = Session.DB(dbName).C("col_button_click_full")
	LogADReports = Session.DB(dbName).C("col_ad_report")
	LogFBReports = Session.DB(dbName).C("col_fb_report")
	NsqLogExternalBets = Session.DB(dbName).C("col_nsq_log_external_bet")
	NsqLogExternalRewards = Session.DB(dbName).C("col_nsq_log_external_reward")
	NsqLogExternalCancels = Session.DB(dbName).C("col_nsq_log_external_cancel")
	NsqLogEfiTransactions = Session.DB(dbName).C("col_nsq_log_efi_transaction")
	LogRechargeReports = Session.DB(dbName).C("col_recharge_report")
	ChatLogs = Session.DB(dbName).C("col_chat_log")
	ServerWhites = Session.DB(dbName).C("col_server_white")
	CustomerAdresses = Session.DB(dbName).C("col_contact_way")
	Cdks = Session.DB(dbName).C("col_cdkey")
	ShareWays = Session.DB(dbName).C("col_share_way")
	LogSmsRecords = Session.DB(dbName).C("col_sms_record")
	LogUserSmsRecords = Session.DB(dbName).C("col_user_sms_record")
	CashFlows = Session.DB(dbName).C("col_cash_flow")
	EquipmentBlacklists = Session.DB(dbName).C("col_equipment_blacklist")
	LogScratchTickets = Session.DB(dbName).C("col_scratch_ticket")
	LogActivityMonitors = Session.DB(dbName).C("col_activity_monitor")
	LogPlayShareDraws = Session.DB(dbName).C("col_playshare_draw")
	LogBonuss = Session.DB(dbName).C("col_bonus_log")
	LogShareRegists = Session.DB(dbName).C("col_share_regist")
	LogGameFlowWaters = Session.DB(dbName).C("col_game_flow_water")
	UploadUserHeads = Session.DB(dbName).C("col_upload_user_head")
	UserGameDatas = Session.DB(dbName).C("s_user_game_data")
	TPStoryChargeStockHistoryLogs = Session.DB(dbName).C("col_log_tp_story_charge_stock_history")
	ActivityTurnTimes = Session.DB(dbName).C("col_activity_turn_times")
	ActivityTurnDrawLogs = Session.DB(dbName).C("col_activity_turn_draw_logs")
	ActivityTurnPrizeLogs = Session.DB(dbName).C("col_activity_turn_prize_logs")
	ShareAgentIncomeRecords = Session.DB(dbName).C("col_share_agent_income_records")
	VolatilitySubsidys = Session.DB(dbName).C("col_activity_volatility_subsidy")
	OnlineUsersLogs = Session.DB(dbName).C("col_log_online_users")
	Utrs = Session.DB(dbName).C("col_utr")
	CumRechargeWheelUnlockLogs = Session.DB(dbName).C("col_cum_recharge_wheel_unlock_log")
	CumRechargeWheelSpinLogs = Session.DB(dbName).C("col_cum_recharge_wheel_spin_log")
}

// 初始化login节点 db连接
func InitLoginMgo(dbHost, dbPort, dbUser, dbPassword, dbName string) {
	// get db config from host, port, username, password
	usernameAndPassword := dbUser + ":" + dbPassword + "@"
	if dbUser == "" || dbPassword == "" {
		usernameAndPassword = ""
	}
	if dbPort == "" {
		dbPort = "27017"
	}
	authString := ""
	if usernameAndPassword != "" {
		authString = "?authSource=admin&authMechanism=SCRAM-SHA-1"
	}
	url := "mongodb://" + usernameAndPassword + dbHost + ":" + dbPort + "/" + dbName + authString

	// [mongodb://][user:pass@]host1[:port1][,host2[:port2],...][/database][?options]
	// mongodb://myuser:mypass@localhost:40001,otherhost:40001/mydb
	var err error
	Session, err = mgo.Dial(url)
	if err != nil {
		//return //TODO test
		panic(err)
	}
	go ticker()

	// Optional. Switch the session to a monotonic behavior.
	Session.SetMode(mgo.Monotonic, true)

	AdjustCallback = Session.DB(dbName).C("col_adjust_callback")
	ClientLogs = Session.DB(dbName).C("col_client_log")
	PlayerUsers = Session.DB(dbName).C("col_user")
	Banners = Session.DB(dbName).C("col_banner")
	BonusBanners = Session.DB(dbName).C("col_bonus_banner")
}

// 初始化短信节点 db连接
func InitSmsMgo(dbHost, dbPort, dbUser, dbPassword, dbName string) {
	// get db config from host, port, username, password
	usernameAndPassword := dbUser + ":" + dbPassword + "@"
	if dbUser == "" || dbPassword == "" {
		usernameAndPassword = ""
	}
	if dbPort == "" {
		dbPort = "27017"
	}
	authString := ""
	if usernameAndPassword != "" {
		authString = "?authSource=admin&authMechanism=SCRAM-SHA-1"
	}
	url := "mongodb://" + usernameAndPassword + dbHost + ":" + dbPort + "/" + dbName + authString

	// [mongodb://][user:pass@]host1[:port1][,host2[:port2],...][/database][?options]
	// mongodb://myuser:mypass@localhost:40001,otherhost:40001/mydb
	var err error
	Session, err = mgo.Dial(url)
	if err != nil {
		//return //TODO test
		panic(err)
	}
	go ticker()

	// Optional. Switch the session to a monotonic behavior.
	Session.SetMode(mgo.Monotonic, true)

	LogSmsRecords = Session.DB(dbName).C("col_sms_record")
	LogUserSmsRecords = Session.DB(dbName).C("col_user_sms_record")
}

func Close() {
	Session.Close()
	if mgoCloseCh != nil {
		close(mgoCloseCh)
	}
}

// common DAO
// 公用方法

//----------------------

func Insert(collection *mgo.Collection, i interface{}) bool {
	err := collection.Insert(i)
	return Err(err)
}

//----------------------

// 适合一条记录全部更新
func Update(collection *mgo.Collection, query interface{}, i interface{}) bool {
	err := collection.Update(query, i)
	return Err(err)
}
func Upsert(collection *mgo.Collection, query interface{}, i interface{}) bool {
	_, err := collection.Upsert(query, i)
	return Err(err)
}
func UpdateAll(collection *mgo.Collection, query interface{}, i interface{}) bool {
	_, err := collection.UpdateAll(query, i)
	return Err(err)
}
func UpdateByIdAndUserId(collection *mgo.Collection, id, userId string, i interface{}) bool {
	err := collection.Update(GetIdAndUserIdQ(id, userId), i)
	return Err(err)
}

func UpdateByIdAndUserId2(collection *mgo.Collection, id, userId bson.ObjectId, i interface{}) bool {
	err := collection.Update(GetIdAndUserIdBsonQ(id, userId), i)
	return Err(err)
}
func UpdateByIdAndUserIdField(collection *mgo.Collection, id, userId, field string, value interface{}) bool {
	return UpdateByIdAndUserId(collection, id, userId, bson.M{"$set": bson.M{field: value}})
}
func UpdateByIdAndUserIdMap(collection *mgo.Collection, id, userId string, v bson.M) bool {
	return UpdateByIdAndUserId(collection, id, userId, bson.M{"$set": v})
}

func UpdateByIdAndUserIdField2(collection *mgo.Collection, id, userId bson.ObjectId, field string, value interface{}) bool {
	return UpdateByIdAndUserId2(collection, id, userId, bson.M{"$set": bson.M{field: value}})
}
func UpdateByIdAndUserIdMap2(collection *mgo.Collection, id, userId bson.ObjectId, v bson.M) bool {
	return UpdateByIdAndUserId2(collection, id, userId, bson.M{"$set": v})
}

func UpdateByQField(collection *mgo.Collection, q interface{}, field string, value interface{}) bool {
	_, err := collection.UpdateAll(q, bson.M{"$set": bson.M{field: value}})
	return Err(err)
}
func UpdateByQI(collection *mgo.Collection, q interface{}, v interface{}) bool {
	_, err := collection.UpdateAll(q, bson.M{"$set": v})
	return Err(err)
}

// 查询条件和值
func UpdateByQMap(collection *mgo.Collection, q interface{}, v interface{}) bool {
	_, err := collection.UpdateAll(q, bson.M{"$set": v})
	return Err(err)
}

//------------------------

// 删除一条
func Delete(collection *mgo.Collection, q interface{}) bool {
	err := collection.Remove(q)
	return Err(err)
}
func DeleteByIdAndUserId(collection *mgo.Collection, id, userId string) bool {
	err := collection.Remove(GetIdAndUserIdQ(id, userId))
	return Err(err)
}
func DeleteByIdAndUserId2(collection *mgo.Collection, id, userId bson.ObjectId) bool {
	err := collection.Remove(GetIdAndUserIdBsonQ(id, userId))
	return Err(err)
}

// 删除所有
func DeleteAllByIdAndUserId(collection *mgo.Collection, id, userId string) bool {
	_, err := collection.RemoveAll(GetIdAndUserIdQ(id, userId))
	return Err(err)
}
func DeleteAllByIdAndUserId2(collection *mgo.Collection, id, userId bson.ObjectId) bool {
	_, err := collection.RemoveAll(GetIdAndUserIdBsonQ(id, userId))
	return Err(err)
}

func DeleteAll(collection *mgo.Collection, q interface{}) bool {
	_, err := collection.RemoveAll(q)
	return Err(err)
}

//-------------------------

func Get(collection *mgo.Collection, id string, i interface{}) {
	collection.FindId(id).One(i)
}
func Get2(collection *mgo.Collection, id bson.ObjectId, i interface{}) {
	collection.FindId(id).One(i)
}
func Get3(collection *mgo.Collection, id string, i interface{}) {
	collection.FindId(bson.ObjectIdHex(id)).One(i)
}

func GetByQ(collection *mgo.Collection, q interface{}, i interface{}) {
	collection.Find(q).One(i)
}
func ListByQ(collection *mgo.Collection, q interface{}, i interface{}) {
	collection.Find(q).All(i)
}

func ListByQLimit(collection *mgo.Collection, q interface{}, i interface{}, limit int) {
	collection.Find(q).Limit(limit).All(i)
}

// 查询某些字段, q是查询条件, fields是字段名列表
func GetByQWithFields(collection *mgo.Collection, q bson.M, fields []string, i interface{}) {
	selector := make(bson.M, len(fields))
	for _, field := range fields {
		selector[field] = true
	}
	collection.Find(q).Select(selector).One(i)
}

// 查询某些字段, q是查询条件, fields是字段名列表
func ListByQWithFields(collection *mgo.Collection, q bson.M, fields []string, i interface{}) {
	selector := make(bson.M, len(fields))
	for _, field := range fields {
		selector[field] = true
	}
	collection.Find(q).Select(selector).All(i)
}
func GetByIdAndUserId(collection *mgo.Collection, id, userId string, i interface{}) {
	collection.Find(GetIdAndUserIdQ(id, userId)).One(i)
}
func GetByIdAndUserId2(collection *mgo.Collection, id, userId bson.ObjectId, i interface{}) {
	collection.Find(GetIdAndUserIdBsonQ(id, userId)).One(i)
}

// 按field去重
func Distinct(collection *mgo.Collection, q bson.M, field string, i interface{}) {
	collection.Find(q).Distinct(field, i)
}

//----------------------

func Count(collection *mgo.Collection, q interface{}) int {
	cnt, err := collection.Find(q).Count()
	if err != nil {
		Err(err)
	}
	return cnt
}

func Has(collection *mgo.Collection, q interface{}) bool {
	if Count(collection, q) > 0 {
		return true
	}
	return false
}

//-----------------

// 得到主键和userId的复合查询条件
func GetIdAndUserIdQ(id, userId string) bson.M {
	return bson.M{"_id": bson.ObjectIdHex(id), "UserId": bson.ObjectIdHex(userId)}
}
func GetIdAndUserIdBsonQ(id, userId bson.ObjectId) bson.M {
	return bson.M{"_id": id, "UserId": userId}
}

// DB处理错误
func Err(err error) bool {
	if err != nil {
		//fmt.Println(err)
		// 删除时, 查找
		if err.Error() == "not found" {
			return true
		}
		return false
	}
	return true
}

// 检查mognodb是否lost connection
// 每个请求之前都要检查!!
func CheckMongoSessionLost() {
	// fmt.Println("检查CheckMongoSessionLostErr")
	err := Session.Ping()
	if err != nil {
		//Log("Lost connection to db!")
		Session.Refresh()
		err = Session.Ping()
		if err == nil {
			//Log("Reconnect to db successful.")
		} else {
			//Log("重连失败!!!! 警告")
		}
	}
}

// 计时器
func ticker() {
	mgoCloseCh = make(chan bool, 1)
	tick := time.Tick(time.Minute)
	for {
		select {
		case <-tick:
			CheckMongoSessionLost()
		case <-mgoCloseCh:
			return
		}
	}
}

// 分页, 排序处理
func parsePageAndSort(pageNumber, pageSize int, sortField string, isAsc bool) (skipNum int, sortFieldR string) {
	skipNum = (pageNumber - 1) * pageSize
	if sortField == "" {
		sortField = "UpdatedTime"
	}
	if !isAsc {
		sortFieldR = "-" + sortField
	} else {
		sortFieldR = sortField
	}
	return
}

// FindFieldsByQ 根据条件查询指定字段
func FindFieldsByQ(collection *mgo.Collection, q interface{}, i interface{}, fields ...string) {
	project := bson.M{}
	for _, field := range fields {
		project[field] = "$" + field
	}
	err := collection.Pipe([]bson.M{
		{"$match": q},
		{"$project": project},
	}).All(i)
	if err != nil {
		glog.Error(err)
	}
}
