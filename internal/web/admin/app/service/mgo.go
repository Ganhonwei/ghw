package service

import (
	"fmt"
	"net"
	"os"
	"time"

	"goserver/gen/pb"
	"goserver/internal/web/admin/app/entity"
	"goserver/pkg/data/ck"
	"goserver/pkg/data/mq"
	"goserver/pkg/utils"

	"github.com/astaxie/beego"
	"golang.org/x/crypto/ssh"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// Init mgo and the common DAO

// 数据连接
var Session *mgo.Session
var Session1 *mgo.Session
var mgoCloseCh chan bool

var locationName string
var location *time.Location
var versionsName string

// 各个表的Collection对象
var Actions *mgo.Collection
var Perms *mgo.Collection
var Roles *mgo.Collection
var RolePerms *mgo.Collection
var Users *mgo.Collection
var UserRoles *mgo.Collection
var MailTpls *mgo.Collection
var GenIDs *mgo.Collection

var TradeRecords *mgo.Collection
var PlayerUsers *mgo.Collection
var GameRecords *mgo.Collection
var UserRecords *mgo.Collection
var StatRecords *mgo.Collection
var UserStatRecords *mgo.Collection
var Pk10Records *mgo.Collection

var Agencys *mgo.Collection
var RegistLogs *mgo.Collection
var DiamondLogs *mgo.Collection
var CoinLogs *mgo.Collection
var ChipLogs *mgo.Collection
var LogBuildAgencys *mgo.Collection
var LogOnlines *mgo.Collection
var LogPayTodays *mgo.Collection
var LogRegistTodays *mgo.Collection
var ApplyCashs *mgo.Collection
var LogChipTodays *mgo.Collection
var AgentFees *mgo.Collection
var AgentFeeLogs *mgo.Collection
var AccountingLogs *mgo.Collection

var Notices *mgo.Collection
var Shops *mgo.Collection
var Envs *mgo.Collection
var Vips *mgo.Collection
var Games *mgo.Collection
var GameStocks *mgo.Collection
var ManualCoupons *mgo.Collection

// CC Add
var Pays *mgo.Collection
var Withdraws *mgo.Collection
var WithdrawSetting *mgo.Collection
var GoldGiftLogs *mgo.Collection
var Packages *mgo.Collection
var SetWithdraws *mgo.Collection
var GameSystems *mgo.Collection
var PayChannels *mgo.Collection
var CustomerMsgs *mgo.Collection
var LogWaters *mgo.Collection
var Stocks *mgo.Collection
var NewbieStocks *mgo.Collection
var Details *mgo.Collection
var Channels *mgo.Collection
var Uploads *mgo.Collection
var Banners *mgo.Collection
var BonusBanners *mgo.Collection
var ContactWays *mgo.Collection
var ChatConfigs *mgo.Collection
var ChatAssigns *mgo.Collection
var Cdkeys *mgo.Collection
var ShareWays *mgo.Collection
var AdReports *mgo.Collection
var CashFlows *mgo.Collection
var LogSmsRecords *mgo.Collection
var LogUserSmsRecords *mgo.Collection
var UserGameDatas *mgo.Collection
var UserGameStrategys *mgo.Collection
var LogGameFlowWaters *mgo.Collection
var PayChannelStatss *mgo.Collection
var PayChannelLogs *mgo.Collection
var ThirdPartyBalances *mgo.Collection
var UploadUserHeads *mgo.Collection
var DsfPayOrder *mgo.Collection
var DsfWithdrawOrder *mgo.Collection
var PayChannelStatRecords *mgo.Collection
var VolatilitySubsidyPools *mgo.Collection

// 统计
var DataSummarys *mgo.Collection
var UserRetaineds *mgo.Collection
var ChannelDatas *mgo.Collection
var UserResources *mgo.Collection
var GlobalResources *mgo.Collection
var GameResources *mgo.Collection
var BasicPays *mgo.Collection
var BasicFrees *mgo.Collection
var RealTimes *mgo.Collection
var ChannelRates *mgo.Collection
var ShareDatas *mgo.Collection
var PointDatas *mgo.Collection
var GameNumberAnalysiss *mgo.Collection
var BugDatas *mgo.Collection
var StrategyDatas *mgo.Collection
var LogEventTracks *mgo.Collection
var LogBugFeedbacks *mgo.Collection
var Playtimes *mgo.Collection
var LogGameTimes *mgo.Collection
var RoomDatas *mgo.Collection
var PayUserRetaineds *mgo.Collection
var GoodsBuys *mgo.Collection
var DataStatisticss *mgo.Collection
var AdReportDatas *mgo.Collection
var ShareStatistcs *mgo.Collection
var SmsDatas *mgo.Collection
var BloggerAccounts *mgo.Collection
var PaySources *mgo.Collection
var ButtonClicks *mgo.Collection
var BattleRooms *mgo.Collection
var WithDrawOptRecords *mgo.Collection
var UserCashRecods *mgo.Collection
var PddStats *mgo.Collection
var CrashPlayerStats *mgo.Collection
var PlanePlayerStats *mgo.Collection
var CrashFYZSStats *mgo.Collection
var CrashYHWMStats *mgo.Collection
var CrashQSHSStats *mgo.Collection
var CrashJCFKStats *mgo.Collection
var CrashMXJLStats *mgo.Collection
var CrashRKYSStats *mgo.Collection
var PlaneFYZSStats *mgo.Collection
var PlaneYHWMStats *mgo.Collection
var PlaneQSHSStats *mgo.Collection
var PlaneJCFKStats *mgo.Collection
var PlaneMXJLStats *mgo.Collection
var PlaneRKYSStats *mgo.Collection
var TPPlayerStats *mgo.Collection
var TPRoomPlayerStats *mgo.Collection
var LHDPlayerStats *mgo.Collection
var UpdownPlayerStats *mgo.Collection
var UpdownXxscStats *mgo.Collection
var UpdownQsbnStats *mgo.Collection
var UpdownLkyhStats *mgo.Collection
var ABPlayerStats *mgo.Collection
var ABAnqsStats *mgo.Collection
var ABArtyStats *mgo.Collection

// CP
var CPStrategies *mgo.Collection
var CPPlayerStats *mgo.Collection
var CPLwjyStats *mgo.Collection
var CPLyqnStats *mgo.Collection

// RB
var RBStrategies *mgo.Collection
var RBPlayerStats *mgo.Collection
var RBHydtStats *mgo.Collection
var RBJcfsStats *mgo.Collection

var RankingLists *mgo.Collection
var SubprojectIncomes *mgo.Collection
var SubprojectIncomeResetLogs *mgo.Collection
var LHDXxscStats *mgo.Collection
var LHDQsbnStats *mgo.Collection
var LHDLkyhStats *mgo.Collection
var FirstCharges *mgo.Collection

var TPStoryChargeStockHistoryLogs *mgo.Collection

// 日志管理
var LoginLogs *mgo.Collection
var OutDiamonds *mgo.Collection
var Givediamonds *mgo.Collection
var IpRecords *mgo.Collection
var CardBlacklists *mgo.Collection         // 银行卡号黑名单
var WithdrawUserBlacklists *mgo.Collection // 支付用户黑名单
var EquipmentBlacklists *mgo.Collection
var IPwhites *mgo.Collection
var ServerWhites *mgo.Collection
var PayBlackLists *mgo.Collection
var Shares *mgo.Collection
var LogShareWaters *mgo.Collection
var LuckyDrawGives *mgo.Collection
var LotteryActivitys *mgo.Collection
var LogVBDiamonds *mgo.Collection
var TPStoryChargeStocks *mgo.Collection
var LevelStockHistorys *mgo.Collection
var RecoverycycleDateConsumes *mgo.Collection
var PayChannelStatNoticeConfigs *mgo.Collection
var ActivityTurnPrizeLogs *mgo.Collection
var GiftPackCodes *mgo.Collection
var FinanceWalletStats *mgo.Collection
var CustomerChatSessions *mgo.Collection
var CustomerChatMessages *mgo.Collection
var CustomerSeats *mgo.Collection
var CustomerSeatLogs *mgo.Collection
var CustomerSettings *mgo.Collection
var CustomerChatPhrases *mgo.Collection

// 外接游戏
var NsqLogExternalBets *mgo.Collection
var NsqLogExternalRewards *mgo.Collection

// nsq
// var producer *nsq.Producer

// 初始化时连接数据库
func InitMgo() {
	var err error
	// get db config from host, port, username, password
	dbHost := beego.AppConfig.String("mdb.host")
	dbPort := beego.AppConfig.String("mdb.port")
	dbUser := beego.AppConfig.String("mdb.user")
	dbPassword := beego.AppConfig.String("mdb.password")
	dbName := beego.AppConfig.String("mdb.name")
	sshON := beego.AppConfig.DefaultBool("mdb.ssh", false)
	sshUser := beego.AppConfig.String("mdb.sshUser")
	sshAddr := beego.AppConfig.String("mdb.sshAddr")
	sshKey := beego.AppConfig.String("mdb.sshKey")

	locationName = beego.AppConfig.String("timezone")
	location, err = time.LoadLocation(locationName)
	if err != nil {
		panic(fmt.Errorf("load timezone %s error, %v", locationName, err))
	}

	versionsName = beego.AppConfig.String("versions")

	// usernameAndPassword := dbUser + ":" + dbPassword + "@"
	// if dbUser == "" || dbPassword == "" {
	// 	usernameAndPassword = ""
	// }
	// if dbPort == "" {
	// 	dbPort = "27017"
	// }
	// authString := ""
	// if usernameAndPassword != "" {
	// 	authString = "?authSource=admin&authMechanism=SCRAM-SHA-1"
	// }
	// url := "mongodb://" + usernameAndPassword + dbHost + ":" + dbPort + "/" + dbName + authString

	// [mongodb://][user:pass@]host1[:port1][,host2[:port2],...][/database][?options]
	// mongodb://myuser:mypass@localhost:40001,otherhost:40001/mydb
	// Session, err = mgo.Dial(url)

	dialInfo := &mgo.DialInfo{
		Addrs:    []string{fmt.Sprintf("%s:%s", dbHost, dbPort)},
		Database: dbName,
		Username: dbUser,
		Password: dbPassword,
	}
	if dbUser != "" && dbPassword != "" {
		dialInfo.Source = "admin"
		dialInfo.Mechanism = "SCRAM-SHA-1"
	}
	// mongo ssh
	if sshON {
		// SSH 配置
		k, err := os.ReadFile(sshKey)
		if err != nil {
			panic(err)
		}
		signer, err := ssh.ParsePrivateKey(k)
		if err != nil {
			panic(err)
		}

		sshConfig := &ssh.ClientConfig{
			User: sshUser,
			Auth: []ssh.AuthMethod{
				// ssh.Password("ssh_password"), // SSH 密码认证
				ssh.PublicKeysCallback(func() ([]ssh.Signer, error) { // SSH 私钥认证
					return []ssh.Signer{signer}, nil
				}),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 忽略主机密钥检查（仅供开发使用）
			Timeout:         5 * time.Second,             // SSH 连接超时
		}

		// 连接 SSH 服务器
		sshClient, err := ssh.Dial("tcp", sshAddr, sshConfig)
		if err != nil {
			panic(err)
		}
		// 创建本地监听器 (本地端口)
		localListener, err := sshClient.Dial("tcp", fmt.Sprintf("%s:%s", dbHost, dbPort)) // 远程 ClickHouse 地址及端口
		if err != nil {
			panic(err)
		}
		dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
			return localListener, nil
		}
	}
	Session, err = mgo.DialWithInfo(dialInfo)
	if err != nil {
		panic(err)
	}

	// Optional. Switch the session to a monotonic behavior.
	Session.SetMode(mgo.Monotonic, true)

	// 日志记录数据库链接
	dbHost_log := beego.AppConfig.String("mgodb.host")
	dbPort_log := beego.AppConfig.String("mgodb.port")
	dbUser_log := beego.AppConfig.String("mgodb.user")
	dbPassword_log := beego.AppConfig.String("mgodb.password")
	dbName_log := beego.AppConfig.String("mgodb.name")
	sshON_log := beego.AppConfig.DefaultBool("mgodb.ssh", false)
	sshUser_log := beego.AppConfig.String("mgodb.sshUser")
	sshAddr_log := beego.AppConfig.String("mgodb.sshAddr")
	sshKey_log := beego.AppConfig.String("mgodb.sshKey")
	// usernameAndPassword_log := dbUser_log + ":" + dbPassword_log + "@"
	// if dbUser_log == "" || dbPassword_log == "" {
	// 	usernameAndPassword_log = ""
	// }
	// if dbPort_log == "" {
	// 	dbPort_log = "27019"
	// }
	// url_log := "mongodb://" + usernameAndPassword_log + dbHost_log + ":" + dbPort_log + "/" + dbName_log + authString

	// [mongodb://][user:pass@]host1[:port1][,host2[:port2],...][/database][?options]
	// mongodb://myuser:mypass@localhost:40001,otherhost:40001/mydb
	// var err1 error
	// Session1, err1 = mgo.Dial(url_log)
	// if err1 != nil {
	// 	panic(err1)
	// }

	dialInfo_log := &mgo.DialInfo{
		Addrs:    []string{fmt.Sprintf("%s:%s", dbHost_log, dbPort_log)},
		Database: dbName_log,
		Username: dbUser_log,
		Password: dbPassword_log,
	}
	if dbUser_log != "" && dbPassword_log != "" {
		dialInfo_log.Source = "admin"
		dialInfo_log.Mechanism = "SCRAM-SHA-1"
	}
	// mongo ssh
	if sshON_log {
		// SSH 配置
		k, err := os.ReadFile(sshKey_log)
		if err != nil {
			panic(err)
		}
		signer, err := ssh.ParsePrivateKey(k)
		if err != nil {
			panic(err)
		}

		sshConfig := &ssh.ClientConfig{
			User: sshUser_log,
			Auth: []ssh.AuthMethod{
				// ssh.Password("ssh_password"), // SSH 密码认证
				ssh.PublicKeysCallback(func() ([]ssh.Signer, error) { // SSH 私钥认证
					return []ssh.Signer{signer}, nil
				}),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 忽略主机密钥检查（仅供开发使用）
			Timeout:         5 * time.Second,             // SSH 连接超时
		}

		// 连接 SSH 服务器
		sshClient, err := ssh.Dial("tcp", sshAddr_log, sshConfig)
		if err != nil {
			panic(err)
		}
		// 创建本地监听器 (本地端口)
		localListener, err := sshClient.Dial("tcp", fmt.Sprintf("%s:%s", dbHost_log, dbPort_log)) // 远程 ClickHouse 地址及端口
		if err != nil {
			panic(err)
		}
		dialInfo_log.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
			return localListener, nil
		}
	}
	Session1, err = mgo.DialWithInfo(dialInfo_log)
	if err != nil {
		panic(err)
	}
	// Optional. Switch the session to a monotonic behavior.
	Session1.SetMode(mgo.Monotonic, true)

	// 定时器
	// go ticker()
	go tickerTesting()
	// 开启定时统计任务
	InitTickerStats()

	// 支付渠道定时通知
	if RunMode != "dev" {
		initTelegramNotify()
		subscribePayOrderError()
		go tickPayChannelStatNotice()
	}

	// niuadmin
	Actions = Session.DB(dbName).C("t_action")
	Perms = Session.DB(dbName).C("t_perm")
	Roles = Session.DB(dbName).C("t_role")
	RolePerms = Session.DB(dbName).C("t_role_perm")
	Users = Session.DB(dbName).C("t_user")
	UserRoles = Session.DB(dbName).C("t_user_role")
	MailTpls = Session.DB(dbName).C("t_mail_tpl")
	GenIDs = Session.DB(dbName).C("t_last_id")

	// trade_record
	TradeRecords = Session.DB(dbName).C("col_trade_record")
	// user
	PlayerUsers = Session.DB(dbName).C("col_user")
	// record
	GameRecords = Session.DB(dbName).C("col_game_record")
	UserRecords = Session.DB(dbName).C("col_user_record")
	StatRecords = Session.DB(dbName).C("col_stat_record")
	UserStatRecords = Session.DB(dbName).C("col_user_stat_record")

	Pk10Records = Session.DB(dbName).C("col_pkten")

	//
	Agencys = Session.DB(dbName).C("col_agency")
	//
	RegistLogs = Session.DB(dbName).C("col_log_regist")
	DiamondLogs = Session.DB(dbName).C("col_log_diamond")
	CoinLogs = Session.DB(dbName).C("col_log_coin")
	ChipLogs = Session.DB(dbName).C("col_log_chip")
	LogBuildAgencys = Session.DB(dbName).C("col_log_build_agency")
	LogOnlines = Session.DB(dbName).C("col_log_online")
	LogPayTodays = Session.DB(dbName).C("col_log_pay_today")
	LogRegistTodays = Session.DB(dbName).C("col_log_regist_today")
	LogChipTodays = Session.DB(dbName).C("col_log_chip_today")
	AgentFees = Session.DB(dbName).C("col_log_agent_fee")
	AgentFeeLogs = Session.DB(dbName).C("col_log_agent_fee_log")
	AccountingLogs = Session.DB(dbName).C("col_log_accounting_log")
	//
	ApplyCashs = Session.DB(dbName).C("col_apply_cash")
	//
	Notices = Session.DB(dbName).C("col_notice")
	Shops = Session.DB(dbName).C("col_shop")
	Envs = Session.DB(dbName).C("col_env")
	Vips = Session.DB(dbName).C("col_vip")
	Games = Session.DB(dbName).C("col_game")
	GameStocks = Session.DB(dbName).C("col_game_stock")

	//CC Add
	Pays = Session.DB(dbName).C("col_trade_record")
	Withdraws = Session.DB(dbName).C("col_withdraw_record")
	WithdrawSetting = Session.DB(dbName).C("col_withdraw_setting")
	GoldGiftLogs = Session.DB(dbName).C("col_gold_gift_log")
	Packages = Session.DB(dbName).C("col_package")
	SetWithdraws = Session.DB(dbName).C("col_set_withdraw")
	GameSystems = Session.DB(dbName).C("col_set_game_system")
	PayChannels = Session.DB(dbName).C("col_pay_channel")
	Channels = Session.DB(dbName).C("col_channel")
	Stocks = Session.DB(dbName).C("col_stock")
	NewbieStocks = Session.DB(dbName).C("col_newbie_stock")
	// 活动
	BasicPays = Session.DB(dbName).C("col_basic_pay")
	BasicFrees = Session.DB(dbName).C("col_basic_free")
	Shares = Session.DB(dbName).C("col_share")
	Banners = Session.DB(dbName).C("col_banner")
	BonusBanners = Session.DB(dbName).C("col_bonus_banner")
	LuckyDrawGives = Session.DB(dbName).C("col_lucky_draw_gives")
	GiftPackCodes = Session.DB(dbName).C("col_activity_gift_pack_code")
	VolatilitySubsidyPools = Session.DB(dbName).C("col_activity_volatility_subsidy_pool")
	ManualCoupons = Session.DB(dbName).C("col_manual_coupon")

	// 27019 日志数据库
	LoginLogs = Session1.DB(dbName_log).C("col_log_login")
	LogWaters = Session1.DB(dbName_log).C("col_log_water")
	Details = Session1.DB(dbName_log).C("col_detail")
	OutDiamonds = Session1.DB(dbName_log).C("col_log_outdiamond")
	Givediamonds = Session1.DB(dbName_log).C("col_log_givediamond")
	LogVBDiamonds = Session1.DB(dbName_log).C("col_log_vbdiamond")
	CardBlacklists = Session1.DB(dbName_log).C("col_card_blacklist")
	WithdrawUserBlacklists = Session1.DB(dbName_log).C("col_user_blacklist")
	IPwhites = Session1.DB(dbName_log).C("col_ip_white")
	ServerWhites = Session1.DB(dbName_log).C("col_server_white")
	PayBlackLists = Session1.DB(dbName_log).C("col_pay_blacklist")
	EquipmentBlacklists = Session1.DB(dbName_log).C("col_equipment_blacklist")
	LogShareWaters = Session1.DB(dbName_log).C("col_share_water")
	Uploads = Session1.DB(dbName_log).C("col_upload_records")
	LogEventTracks = Session1.DB(dbName_log).C("col_event_track")
	LogBugFeedbacks = Session1.DB(dbName_log).C("col_bug_feedback")
	LogGameTimes = Session1.DB(dbName_log).C("col_game_time")
	IpRecords = Session1.DB(dbName_log).C("col_user_ip_records")
	CustomerMsgs = Session1.DB(dbName_log).C("col_chat_log")
	ContactWays = Session1.DB(dbName_log).C("col_contact_way")
	ChatConfigs = Session1.DB(dbName_log).C("col_chat_config")
	ChatAssigns = Session1.DB(dbName_log).C("col_chat_assign_log")
	Cdkeys = Session1.DB(dbName_log).C("col_cdkey")
	ShareWays = Session1.DB(dbName_log).C("col_share_way")
	AdReports = Session1.DB(dbName_log).C("col_ad_report")
	CashFlows = Session1.DB(dbName_log).C("col_cash_flow")
	LogSmsRecords = Session1.DB(dbName_log).C("col_sms_record")
	LogUserSmsRecords = Session1.DB(dbName_log).C("col_user_sms_record")
	BloggerAccounts = Session1.DB(dbName_log).C("col_blogger_account")
	ButtonClicks = Session1.DB(dbName_log).C("col_button_click")
	BattleRooms = Session1.DB(dbName_log).C("s_battle_room_data")
	UserGameDatas = Session1.DB(dbName_log).C("s_user_game_data")
	UserGameStrategys = Session1.DB(dbName_log).C("s_user_game_strategy")
	LogGameFlowWaters = Session1.DB(dbName_log).C("col_game_flow_water")
	WithDrawOptRecords = Session1.DB(dbName_log).C("col_withdraw_opt_record")
	PayChannelStatss = Session1.DB(dbName_log).C("s_pay_channel_stats")
	PayChannelLogs = Session1.DB(dbName_log).C("col_pay_channel_log")
	ThirdPartyBalances = Session1.DB(dbName_log).C("col_third_party_balance")
	UploadUserHeads = Session1.DB(dbName_log).C("col_upload_user_head")
	DsfPayOrder = Session1.DB(dbName_log).C("dsf_pay_order")
	DsfWithdrawOrder = Session1.DB(dbName_log).C("dsf_withdraw_order")
	PayChannelStatRecords = Session1.DB(dbName_log).C("col_pay_channel_stat_records")
	// 统计
	DataSummarys = Session1.DB(dbName_log).C("s_data_summary")
	UserRetaineds = Session1.DB(dbName_log).C("s_user_retained")
	ChannelDatas = Session1.DB(dbName_log).C("s_channel_data")
	UserResources = Session1.DB(dbName_log).C("s_user_resource")
	GlobalResources = Session1.DB(dbName_log).C("s_global_resource")
	GameResources = Session1.DB(dbName_log).C("s_game_resource")
	RealTimes = Session1.DB(dbName_log).C("s_real_time_data")
	ChannelRates = Session1.DB(dbName_log).C("s_channel_success_rate")
	ShareDatas = Session1.DB(dbName_log).C("s_share_activity_data")
	PointDatas = Session1.DB(dbName_log).C("s_point_data")
	GameNumberAnalysiss = Session1.DB(dbName_log).C("s_game_number_analysis")
	BugDatas = Session1.DB(dbName_log).C("s_bug_statistics")
	StrategyDatas = Session1.DB(dbName_log).C("s_strategy_data")
	Playtimes = Session1.DB(dbName_log).C("s_playtime_analysis")
	RoomDatas = Session1.DB(dbName_log).C("s_room_data")
	PayUserRetaineds = Session1.DB(dbName_log).C("s_pay_user_retained")
	GoodsBuys = Session1.DB(dbName_log).C("s_goods_buy_data")
	DataStatisticss = Session1.DB(dbName_log).C("s_data_statistics")
	AdReportDatas = Session1.DB(dbName_log).C("s_ad_report_data")
	ShareStatistcs = Session1.DB(dbName_log).C("s_share_data")
	SmsDatas = Session1.DB(dbName_log).C("s_sms_data")
	PaySources = Session1.DB(dbName_log).C("s_pay_source")
	UserCashRecods = Session1.DB(dbName_log).C("s_user_cash_recod")
	PddStats = Session1.DB(dbName_log).C("s_pdd_stats")
	LotteryActivitys = Session1.DB(dbName_log).C("s_lottery_activity")
	CrashPlayerStats = Session1.DB(dbName_log).C("s_crash_player_stats")
	PlanePlayerStats = Session1.DB(dbName_log).C("s_plane_player_stats")
	CrashFYZSStats = Session1.DB(dbName_log).C("s_crash_fyzs_stats")
	CrashYHWMStats = Session1.DB(dbName_log).C("s_crash_yhwm_stats")
	CrashQSHSStats = Session1.DB(dbName_log).C("s_crash_qshs_stats")
	CrashJCFKStats = Session1.DB(dbName_log).C("s_crash_jcfk_stats")
	CrashMXJLStats = Session1.DB(dbName_log).C("s_crash_mxjl_stats")
	CrashRKYSStats = Session1.DB(dbName_log).C("s_crash_rkys_stats")
	PlaneFYZSStats = Session1.DB(dbName_log).C("s_plane_fyzs_stats")
	PlaneYHWMStats = Session1.DB(dbName_log).C("s_plane_yhwm_stats")
	PlaneQSHSStats = Session1.DB(dbName_log).C("s_plane_qshs_stats")
	PlaneJCFKStats = Session1.DB(dbName_log).C("s_plane_jcfk_stats")
	PlaneMXJLStats = Session1.DB(dbName_log).C("s_plane_mxjl_stats")
	PlaneRKYSStats = Session1.DB(dbName_log).C("s_plane_rkys_stats")
	TPPlayerStats = Session1.DB(dbName_log).C("s_tp_player_stats")
	TPRoomPlayerStats = Session1.DB(dbName_log).C("s_tp_room_stats")
	LHDPlayerStats = Session1.DB(dbName_log).C("s_lhd_player_stats")
	UpdownPlayerStats = Session1.DB(dbName_log).C("s_updonw_player_stats")
	RankingLists = Session1.DB(dbName_log).C("s_ranking_list")
	SubprojectIncomes = Session1.DB(dbName_log).C("s_subproject_income")
	SubprojectIncomeResetLogs = Session1.DB(dbName_log).C("s_subproject_income_reset_log")
	LHDXxscStats = Session1.DB(dbName_log).C("s_lhd_xxsc_stats")
	LHDQsbnStats = Session1.DB(dbName_log).C("s_lhd_qsbn_stats")
	LHDLkyhStats = Session1.DB(dbName_log).C("s_lhd_lkyh_stats")

	UpdownXxscStats = Session1.DB(dbName_log).C("s_updown_xxsc_stats")
	UpdownQsbnStats = Session1.DB(dbName_log).C("s_updown_qsbn_stats")
	UpdownLkyhStats = Session1.DB(dbName_log).C("s_updown_lkyh_stats")

	ABPlayerStats = Session1.DB(dbName_log).C("s_ab_player_stats")
	ABAnqsStats = Session1.DB(dbName_log).C("s_ab_anqs_stats")
	ABArtyStats = Session1.DB(dbName_log).C("s_ab_arty_stats")

	// CP
	CPStrategies = Session.DB(dbName).C("col_cp_strategy")
	CPPlayerStats = Session1.DB(dbName_log).C("s_cp_player_stats")
	CPLwjyStats = Session1.DB(dbName_log).C("s_cp_lwjy_stats")
	CPLyqnStats = Session1.DB(dbName_log).C("s_cp_lyqn_stats")

	// RB
	RBStrategies = Session.DB(dbName).C("col_rb_strategy")
	RBPlayerStats = Session1.DB(dbName_log).C("s_rb_player_stats")
	RBHydtStats = Session1.DB(dbName_log).C("s_rb_hydt_stats")
	RBJcfsStats = Session1.DB(dbName_log).C("s_rb_jcfs_stats")

	FirstCharges = Session1.DB(dbName_log).C("s_first_charge_analysis")
	TPStoryChargeStocks = Session.DB(dbName).C("col_tp_story_charge_stock")
	TPStoryChargeStockHistoryLogs = Session1.DB(dbName_log).C("col_log_tp_story_charge_stock_history")
	LevelStockHistorys = Session1.DB(dbName_log).C("col_level_stock_history")
	RecoverycycleDateConsumes = Session1.DB(dbName_log).C("col_recoverycycle_date_consumes")
	PayChannelStatNoticeConfigs = Session1.DB(dbName_log).C("col_pay_channel_stat_notice_config")
	ActivityTurnPrizeLogs = Session1.DB(dbName_log).C("col_activity_turn_prize_logs")
	FinanceWalletStats = Session1.DB(dbName_log).C("col_finance_wallet_stats")
	CustomerChatSessions = Session1.DB(dbName_log).C("col_customer_chat_sessions")
	CustomerChatMessages = Session1.DB(dbName_log).C("col_customer_chat_messages")
	CustomerSeats = Session1.DB(dbName_log).C("col_customer_seats")
	CustomerSeatLogs = Session1.DB(dbName_log).C("col_customer_seat_logs")
	CustomerSettings = Session1.DB(dbName_log).C("col_customer_settings")
	CustomerChatPhrases = Session1.DB(dbName_log).C("col_customer_chat_phrases")
	// 外接游戏
	NsqLogExternalBets = Session1.DB(dbName_log).C("col_nsq_log_external_bet")
	NsqLogExternalRewards = Session1.DB(dbName_log).C("col_nsq_log_external_reward")

	//init
	initService()

	//创建文件目录
	os.MkdirAll(GetFilePath(), 0755)
	//TODO test
	//dayStamp := utils.Stamp2Time(utils.TimestampYesterday())
	//dayStamp := TimeYesterday4()
	//statProfit(dayStamp)
	//AgencyService.statChipToday(dayStamp) //定时统计更新
	//initParentAgent()
	//initParentAgent2()
	//num := LoggerService.getRobotProfitsYesterday()
	//beego.Trace("getRobotProfitsYesterday : ", num)
}

// func InitNsq() {
// 	nsqAddr := beego.AppConfig.String("nsq.addr")
// 	var err error
// 	producer, err = nsq.NewProducer(nsqAddr, nsq.NewConfig())
// 	if err != nil {
// 		panic(err)
// 	}
// }

func InitNats() {
	natsUrl := beego.AppConfig.String("nats.url")
	if err := mq.InitNats(natsUrl); err != nil {
		panic(err)
	}
	ck.InitNatsProducer()
}

// 获取后台版本 1：O服；2：M服；
func GetVersions() int {
	vers := 0
	switch versionsName {
	case "origin":
		vers = 1
	case "mengjiala":
		vers = 2
	case "PK":
		vers = 3
	}
	return vers
}

// 创建文件目录
func GetFilePath() string {
	return fmt.Sprintf(beego.AppConfig.String("files_dir"))
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
func GetByInt(collection *mgo.Collection, id int, i interface{}) {
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
		fmt.Println("Lost connection to db!")
		Session.Refresh()
		err = Session.Ping()
		if err == nil {
			fmt.Println("Reconnect to db successful.")
		} else {
			fmt.Println("重连失败!!!! 警告")
		}
	}
}

// 半小时执行一次定时器
func tickerTesting() {
	// 创建半小时定时器
	halfHourTicker := time.NewTicker(30 * time.Minute)
	go func() {
		for range halfHourTicker.C {
			// IP检测记录，从23.11.1-执行时间
			list1, _ := GameService.GetGameSystem(6)
			info := new(entity.GameSystem)
			isIP := false
			if len(list1) > 0 {
				for _, item := range list1 {
					if item.Name == "IP检测" && item.Status == 1 {
						isIP = true
					}
				}
			}
			if isIP {
				IPTesting(utils.Timestamp())
				info = &list1[0]
				info.Status = 0
				GameService.AddOrUpdateGameSystem(info)
			}
			// 检测短信状态
			checkSMS()
			fmt.Println("自动审核、IP检测定时任务")

			// 用户充值、提现金额汇总
			ComputePayAndWithdraw()
		}
	}()

	// 创建1分钟定时器
	fiveMinuteTicker := time.NewTicker(1 * time.Minute)
	go func() {
		for range fiveMinuteTicker.C {
			list, _ := GameService.GetGameSystem(4)
			isExecute := false
			isReturn := false
			isNotSpecial := false

			if len(list) > 0 {
				for _, item := range list {
					if item.Name == "提现自动审核" && (item.Status == 1 || item.Status == 2) {
						isExecute = true
						if item.Status == 2 {
							isNotSpecial = true
						}
					}
					if item.Name == "提现失败自动退回" && item.Status == 1 {
						isReturn = true
					}
				}
			}
			// 执行审核任务
			if isExecute {
				PayService.AutomationAudit(utils.Timestamp(), isNotSpecial)
			}

			//提现失败自动退回
			// 提现订单检测是否存在待核对订单
			_ = isReturn
			// PayService.CheckOrderStatus(isReturn)
			// fmt.Println("定时检查提现订单状态")
		}
	}()

	// 创建1分钟定时器
	oneMinuteTicker := time.NewTicker(1 * time.Minute)
	go func() {
		for range oneMinuteTicker.C {
			MessageService.ChatAssign()
			fmt.Println("定时检查客服消息")
		}
	}()
	select {}
}

// 发送邮件提醒
func SendMail(msg *pb.AlertorMail) {
	// msg := &pb.AlertorMail{Subject: "test", Message: "web test"}
	// body, err := msg.Marshal()
	// if err != nil {
	// 	beego.Error("SendMail fail err: ", err)
	// 	return
	// }

	// err = producer.Publish(data.TopicAlertor, body)
	err := mq.NatsPublish(mq.TopicAlertEmail, msg)
	if err != nil {
		beego.Error("SendMail publish error: ", err)
	}
}

// 分页, 排序处理
func parsePageAndSort(pageNumber, pageSize int, sortField string, isAsc bool) (skipNum int, sortFieldR string) {
	skipNum = (pageNumber - 1) * pageSize
	if skipNum < 0 {
		skipNum = 0
	}
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

func NowTime() time.Time {
	loc, _ := time.LoadLocation(locationName)
	return time.Now().In(loc)
}

func Location() *time.Location {
	loc, _ := time.LoadLocation(locationName)
	return loc
}

// 时间查询
func FindByDate(startDate, endDate, startField, endField string) bson.M {
	m := bson.M{}
	if startDate == "" && endDate == "" {
		return m
	}
	s := fmt.Sprintf("%s 00:00:00", startDate)
	startTime := utils.Str2TimeZone(s, locationName)
	e := fmt.Sprintf("%s 23:59:59", endDate)
	endTime := utils.Str2TimeZone(e, locationName)
	if !startTime.IsZero() && !endTime.IsZero() &&
		startDate != "" && endDate != "" &&
		startField == endField && startField != "" {
		m[startField] = bson.M{"$gte": startTime, "$lte": endTime}
	} else if !startTime.IsZero() && startField != "" &&
		startDate != "" {
		m[startField] = bson.M{"$gte": startTime}
	} else if !endTime.IsZero() && endField != "" &&
		endDate != "" {
		m[endField] = bson.M{"$lte": endTime}
	}
	return m
}
func FindByDateBy2(startDate, endDate, startDate1, endDate1, startField, startField1, endField, endField1 string) bson.M {
	m := bson.M{}
	if startDate == "" && endDate == "" {
		return m
	}
	s := fmt.Sprintf("%s 00:00:00", startDate)
	startTime := utils.Str2TimeZone(s, locationName)
	e := fmt.Sprintf("%s 23:59:59", endDate)
	endTime := utils.Str2TimeZone(e, locationName)
	if !startTime.IsZero() && !endTime.IsZero() &&
		startDate != "" && endDate != "" &&
		startField == endField && startField != "" {
		m[startField] = bson.M{"$gte": startTime, "$lte": endTime}
	} else if !startTime.IsZero() && startField != "" &&
		startDate != "" {
		m[startField] = bson.M{"$gte": startTime}
	} else if !endTime.IsZero() && endField != "" &&
		endDate != "" {
		m[endField] = bson.M{"$lte": endTime}
	}

	s1 := fmt.Sprintf("%s 00:00:00", startDate1)
	startTime1 := utils.Str2TimeZone(s1, locationName)
	e1 := fmt.Sprintf("%s 23:59:59", endDate1)
	endTime1 := utils.Str2TimeZone(e1, locationName)
	if !startTime1.IsZero() && !endTime1.IsZero() &&
		startDate1 != "" && endDate1 != "" &&
		startField1 == endField1 && startField1 != "" {
		m[startField1] = bson.M{"$gte": startTime1, "$lte": endTime1}
	} else if !startTime.IsZero() && startField != "" &&
		startDate1 != "" {
		m[startField] = bson.M{"$gte": startTime1}
	} else if !endTime1.IsZero() && endField1 != "" &&
		endDate1 != "" {
		m[endField1] = bson.M{"$lte": endTime1}
	}

	return m
}

func FindByDate1(startDate, endDate, startField, endField string) bson.M {
	m := bson.M{}
	if startDate == "" && endDate == "" {
		return m
	}
	s := fmt.Sprintf("%s 00:00:00", startDate)
	s1 := utils.Str2TimeZone(s, locationName)
	startTime := utils.Time2Stamp(s1)
	e := fmt.Sprintf("%s 23:59:59", endDate)
	s2 := utils.Str2TimeZone(e, locationName)
	endTime := utils.Time2Stamp(s2)

	if startDate != "" && endDate != "" && startField == endField && startField != "" {
		m[startField] = bson.M{"$gte": startTime, "$lte": endTime}
	} else if startField != "" && startDate != "" {
		m[startField] = bson.M{"$gte": startTime}
	} else if endField != "" && endDate != "" {
		m[endField] = bson.M{"$lte": endTime}
	}
	return m
}

// 使用时间转成毫秒级时间戳
func FindByDate3(startDate, endDate, startField, endField string) bson.M {
	m := bson.M{}
	if startDate == "" && endDate == "" {
		return m
	}
	s := fmt.Sprintf("%s 00:00:00", startDate)
	s1 := utils.Str2TimeZone(s, locationName)
	startTime := utils.Time2StampToMS(s1)
	e := fmt.Sprintf("%s 23:59:59", endDate)
	s2 := utils.Str2TimeZone(e, locationName)
	endTime := utils.Time2StampToMS(s2)

	if startDate != "" && endDate != "" && startField == endField && startField != "" {
		m[startField] = bson.M{"$gte": startTime, "$lte": endTime}
	} else if startField != "" && startDate != "" {
		m[startField] = bson.M{"$gte": startTime}
	} else if endField != "" && endDate != "" {
		m[endField] = bson.M{"$lte": endTime}
	}
	return m
}

// 使用时间转成毫秒级时间戳
func FindByDate4(startDate, endDate string) (start, end *time.Time) {
	if startDate == "" && endDate == "" {
		return
	}
	s := fmt.Sprintf("%s 00:00:00", startDate)
	startTime := utils.Str2TimeZone(s, locationName)
	e := fmt.Sprintf("%s 23:59:59", endDate)
	endTime := utils.Str2TimeZone(e, locationName)

	if startDate != "" && endDate != "" {
		start = &startTime
		end = &endTime
	} else if startDate != "" {
		start = &startTime
	} else if endDate != "" {
		end = &endTime
	}
	return
}

// 使用时间转成秒级时间戳
func FindByDate5(startDate, endDate, startField, endField string) bson.M {
	m := bson.M{}
	if startDate == "" && endDate == "" {
		return m
	}
	s := fmt.Sprintf("%s 00:00:00", startDate)
	s1 := utils.Str2TimeZone(s, locationName)
	startTime := utils.Time2Stamp(s1)
	e := fmt.Sprintf("%s 23:59:59", endDate)
	s2 := utils.Str2TimeZone(e, locationName)
	endTime := utils.Time2Stamp(s2)

	if startDate != "" && endDate != "" && startField == endField && startField != "" {
		m[startField] = bson.M{"$gte": startTime, "$lte": endTime}
	} else if startField != "" && startDate != "" {
		m[startField] = bson.M{"$gte": startTime}
	} else if endField != "" && endDate != "" {
		m[endField] = bson.M{"$lte": endTime}
	}
	return m
}

// ConvertToIndiaTime 将给定的时间戳转换为印度时区的时间，并返回该时间
func ConvertToIndiaTime(timestamp int64) (time.Time, error) {
	location, err := time.LoadLocation(locationName)
	if err != nil {
		return time.Unix(timestamp, 0), err
	}
	return time.Unix(timestamp, 0).In(location), nil
}

// 毫秒时间戳转成印度时间
func ConvertToIndiaTime1(timestamp int64) (time.Time, error) {
	indiaTimeZone, err := time.LoadLocation(locationName)
	if err != nil {
		fmt.Println("无法加载时区:", err)
		return time.Time{}, err
	}
	// 毫秒时间戳
	millisecondTimestamp := int64(timestamp)
	// 将毫秒时间戳转换为秒级时间戳
	secondTimestamp := millisecondTimestamp / 1000
	// 根据秒级时间戳创建时间对象
	return time.Unix(secondTimestamp, 0).In(indiaTimeZone), nil
}

func isSameDayIndia(time1, time2 time.Time) bool {
	// 将时间转换为印度时区
	loc, err := time.LoadLocation(locationName)
	if err != nil {
		fmt.Println("Failed to load Indian timezone:", err)
		return false
	}

	time1India := time1.In(loc)
	time2India := time2.In(loc)

	// 比较年、月和日部分是否相等
	return time1India.Year() == time2India.Year() &&
		time1India.Month() == time2India.Month() &&
		time1India.Day() == time2India.Day()
}

// 筹码转换为分展示
func Chip2Float[T int | int64 | int32 | uint32 | float64](chip T) float64 {
	return (float64(chip) / 100)
}

// 昨日凌晨4点30分
func TimestampYesterday4() int64 {
	return TimestampToday4(location) - 86400
}

// 今日凌晨4点30分
func TimestampToday4(loc *time.Location) int64 {
	return utils.TimestampTodayTime(loc).Unix() + 14400 + 1800
}

// 以凌晨4点算昨日开始
func TimeYesterday4() time.Time {
	return TimeToday4().AddDate(0, 0, -1)
}

// 以凌晨4点算今日开始
func TimeToday4() time.Time {
	now := utils.Timestamp()
	t4 := TimestampToday4(location)
	//0点到4点30分之间
	if now < t4 {
		//昨日凌晨4点30分
		return utils.Stamp2Time(TimestampYesterday4())
	}
	//今日凌晨4点30分
	return utils.Stamp2Time(t4)
}
