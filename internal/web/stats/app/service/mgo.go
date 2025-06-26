package service

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"goserver/internal/web/stats/app/entity"
	"goserver/pkg/utils"

	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Init mgo and the common DAO

// 数据连接
var Session *mgo.Session
var Session1 *mgo.Session
var mgoCloseCh chan bool
var locationName string
var location *time.Location

// 各个表的Collection对象
var Actions *mgo.Collection
var Perms *mgo.Collection
var Roles *mgo.Collection
var RolePerms *mgo.Collection
var RoleChannels *mgo.Collection
var Users *mgo.Collection
var UserRoles *mgo.Collection
var GenIDs *mgo.Collection
var Channels *mgo.Collection
var PlayerUsers *mgo.Collection
var Pays *mgo.Collection
var Withdraws *mgo.Collection
var Details *mgo.Collection
var Stocks *mgo.Collection
var Games *mgo.Collection

// 日志库
var BloggerAccounts *mgo.Collection
var LogEventTracks *mgo.Collection
var LogGameTimes *mgo.Collection
var StatsLogs *mgo.Collection
var ScratchTicketss *mgo.Collection
var LogActivityMonitors *mgo.Collection
var LogPlayShareDraws *mgo.Collection

// 统计
var DataStatisticss *mgo.Collection
var ShareStatistcs *mgo.Collection
var UserRetaineds *mgo.Collection
var PayUserRetaineds *mgo.Collection
var Playtimes *mgo.Collection
var PointDatas *mgo.Collection
var GameNumberAnalysiss *mgo.Collection
var DataSummarys *mgo.Collection
var ChannelDatas *mgo.Collection
var RealTimes *mgo.Collection
var LogWaters *mgo.Collection
var BasicPays *mgo.Collection
var BasicFrees *mgo.Collection
var UserResources *mgo.Collection
var GlobalResources *mgo.Collection
var GameResources *mgo.Collection
var PayChannels *mgo.Collection
var ChannelRates *mgo.Collection
var ShareDatas *mgo.Collection
var LogBugFeedbacks *mgo.Collection
var RoomDatas *mgo.Collection
var BugDatas *mgo.Collection
var GoodsBuys *mgo.Collection
var AdReportDatas *mgo.Collection
var LogSmsRecords *mgo.Collection
var LogUserSmsRecords *mgo.Collection
var SmsDatas *mgo.Collection
var PaySources *mgo.Collection
var ButtonClicks *mgo.Collection
var LogButton2Clicks *mgo.Collection
var LotteryActivitys *mgo.Collection
var UserGameDatas *mgo.Collection
var UserGameStrategys *mgo.Collection
var PayChannelStatss *mgo.Collection
var PayChannelLogs *mgo.Collection
var ThirdPartyBalances *mgo.Collection

// 日志管理
var LoginLogs *mgo.Collection
var OutDiamonds *mgo.Collection
var Givediamonds *mgo.Collection
var IpRecords *mgo.Collection
var LogShareWaters *mgo.Collection
var BattleRooms *mgo.Collection
var BounsLogs *mgo.Collection
var PddStats *mgo.Collection
var CrashPlayerStats *mgo.Collection
var PlanePlayerStats *mgo.Collection
var LHDPlayerStats *mgo.Collection
var LHDXxscStats *mgo.Collection
var LHDQsbnStats *mgo.Collection
var LHDLkyhStats *mgo.Collection
var FirstCharges *mgo.Collection
var UPStrategies *mgo.Collection
var UpdownPlayerStats *mgo.Collection
var UpdownXxscStats *mgo.Collection
var UpdownQsbnStats *mgo.Collection
var UpdownLkyhStats *mgo.Collection

var ABStrategies *mgo.Collection
var ABPlayerStats *mgo.Collection
var ABAnqsStats *mgo.Collection
var ABArtyStats *mgo.Collection

var CPStrategies *mgo.Collection
var CPPlayerStats *mgo.Collection
var CPLwjyStats *mgo.Collection
var CPLyqnStats *mgo.Collection

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

var CrashStrategies *mgo.Collection
var PlaneStrategies *mgo.Collection
var ChargeClassifys *mgo.Collection
var TPPlayerStats *mgo.Collection
var TPRoomPlayerStats *mgo.Collection
var RankingLists *mgo.Collection
var LHDStrategies *mgo.Collection

// RB
var RBStrategies *mgo.Collection
var RBPlayerStats *mgo.Collection
var RBHydtStats *mgo.Collection
var RBJcfsStats *mgo.Collection

// 外接游戏
var NsqLogExternalBets *mgo.Collection
var NsqLogExternalRewards *mgo.Collection

var AdReports *mgo.Collection

// 初始化时连接数据库
func InitMgo() {
	// get db config from host, port, username, password
	dbHost := beego.AppConfig.String("mdb.host")
	dbPort := beego.AppConfig.String("mdb.port")
	dbUser := beego.AppConfig.String("mdb.user")
	dbPassword := beego.AppConfig.String("mdb.password")
	dbName := beego.AppConfig.String("mdb.name")
	// sshON := beego.AppConfig.DefaultBool("mdb.ssh", false)
	// sshUser := beego.AppConfig.String("mdb.sshUser")
	// sshAddr := beego.AppConfig.String("mdb.sshAddr")
	// sshKey := beego.AppConfig.String("mdb.sshKey")
	locationName = beego.AppConfig.String("timezone")
	location, err = time.LoadLocation(locationName)
	if err != nil {
		panic(err)
	}

	var err error
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
	Session, err = mgo.Dial(url)
	if err != nil {
		panic(err)
	}

	// dialInfo := &mgo.DialInfo{
	// 	Addrs:    []string{fmt.Sprintf("%s:%s", dbHost, dbPort)},
	// 	Database: dbName,
	// 	Username: dbUser,
	// 	Password: dbPassword,
	// }
	// if dbUser != "" && dbPassword != "" {
	// 	dialInfo.Source = "admin"
	// 	dialInfo.Mechanism = "SCRAM-SHA-1"
	// }
	// // mongo ssh
	// if sshON {
	// 	// SSH 配置
	// 	k, err := os.ReadFile(sshKey)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	signer, err := ssh.ParsePrivateKey(k)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	sshConfig := &ssh.ClientConfig{
	// 		User: sshUser,
	// 		Auth: []ssh.AuthMethod{
	// 			// ssh.Password("ssh_password"), // SSH 密码认证
	// 			ssh.PublicKeysCallback(func() ([]ssh.Signer, error) { // SSH 私钥认证
	// 				return []ssh.Signer{signer}, nil
	// 			}),
	// 		},
	// 		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 忽略主机密钥检查（仅供开发使用）
	// 		Timeout:         5 * time.Second,             // SSH 连接超时
	// 	}

	// 	// 连接 SSH 服务器
	// 	sshClient, err := ssh.Dial("tcp", sshAddr, sshConfig)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	// 创建本地监听器 (本地端口)
	// 	localListener, err := sshClient.Dial("tcp", fmt.Sprintf("%s:%s", dbHost, dbPort)) // 远程 ClickHouse 地址及端口
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
	// 		return localListener, nil
	// 	}
	// }
	// Session, err = mgo.DialWithInfo(dialInfo)
	// if err != nil {
	// 	panic(err)
	// }

	// Optional. Switch the session to a monotonic behavior.
	Session.SetMode(mgo.Monotonic, true)

	// 日志记录数据库链接
	dbHost_log := beego.AppConfig.String("mgodb.host")
	dbPort_log := beego.AppConfig.String("mgodb.port")
	dbUser_log := beego.AppConfig.String("mgodb.user")
	dbPassword_log := beego.AppConfig.String("mgodb.password")
	dbName_log := beego.AppConfig.String("mgodb.name")
	// sshON_log := beego.AppConfig.DefaultBool("mgodb.ssh", false)
	// sshUser_log := beego.AppConfig.String("mgodb.sshUser")
	// sshAddr_log := beego.AppConfig.String("mgodb.sshAddr")
	// sshKey_log := beego.AppConfig.String("mgodb.sshKey")

	usernameAndPassword_log := dbUser_log + ":" + dbPassword_log + "@"
	if dbUser_log == "" || dbPassword_log == "" {
		usernameAndPassword_log = ""
	}
	if dbPort_log == "" {
		dbPort_log = "27019"
	}
	url_log := "mongodb://" + usernameAndPassword_log + dbHost_log + ":" + dbPort_log + "/" + dbName_log + authString

	// [mongodb://][user:pass@]host1[:port1][,host2[:port2],...][/database][?options]
	// mongodb://myuser:mypass@localhost:40001,otherhost:40001/mydb
	var err1 error
	Session1, err1 = mgo.Dial(url_log)
	if err1 != nil {
		panic(err1)
	}

	// dialInfo_log := &mgo.DialInfo{
	// 	Addrs:    []string{fmt.Sprintf("%s:%s", dbHost_log, dbPort_log)},
	// 	Database: dbName_log,
	// 	Username: dbUser_log,
	// 	Password: dbPassword_log,
	// }
	// if dbUser_log != "" && dbPassword_log != "" {
	// 	dialInfo_log.Source = "admin"
	// 	dialInfo_log.Mechanism = "SCRAM-SHA-1"
	// }
	// // mongo ssh
	// if sshON_log {
	// 	// SSH 配置
	// 	k, err := os.ReadFile(sshKey_log)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	signer, err := ssh.ParsePrivateKey(k)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	sshConfig := &ssh.ClientConfig{
	// 		User: sshUser_log,
	// 		Auth: []ssh.AuthMethod{
	// 			// ssh.Password("ssh_password"), // SSH 密码认证
	// 			ssh.PublicKeysCallback(func() ([]ssh.Signer, error) { // SSH 私钥认证
	// 				return []ssh.Signer{signer}, nil
	// 			}),
	// 		},
	// 		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 忽略主机密钥检查（仅供开发使用）
	// 		Timeout:         5 * time.Second,             // SSH 连接超时
	// 	}

	// 	// 连接 SSH 服务器
	// 	sshClient, err := ssh.Dial("tcp", sshAddr_log, sshConfig)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	// 创建本地监听器 (本地端口)
	// 	localListener, err := sshClient.Dial("tcp", fmt.Sprintf("%s:%s", dbHost_log, dbPort_log)) // 远程 ClickHouse 地址及端口
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	dialInfo_log.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
	// 		return localListener, nil
	// 	}
	// }
	// Session1, err = mgo.DialWithInfo(dialInfo_log)
	// if err != nil {
	// 	panic(err)
	// }
	// Optional. Switch the session to a monotonic behavior.
	Session1.SetMode(mgo.Monotonic, true)

	// 定时器
	go ticker()

	go tickerExecuteDay()

	// go DataAcquisition()

	// niuadmin
	Actions = Session.DB(dbName).C("t_action")
	Perms = Session.DB(dbName).C("t_perm")
	Roles = Session.DB(dbName).C("t_role")
	RolePerms = Session.DB(dbName).C("t_role_perm")
	Users = Session.DB(dbName).C("t_user")
	UserRoles = Session.DB(dbName).C("t_user_role")
	GenIDs = Session.DB(dbName).C("t_last_id")
	Channels = Session.DB(dbName).C("col_channel")
	PlayerUsers = Session.DB(dbName).C("col_user")
	Pays = Session.DB(dbName).C("col_trade_record")
	Withdraws = Session.DB(dbName).C("col_withdraw_record")
	Details = Session1.DB(dbName_log).C("col_detail")
	Stocks = Session.DB(dbName).C("col_stock")
	Games = Session.DB(dbName).C("col_game")
	// 日志库
	RoleChannels = Session1.DB(dbName_log).C("col_blogger_channel")
	BloggerAccounts = Session1.DB(dbName_log).C("col_blogger_account")
	LoginLogs = Session1.DB(dbName_log).C("col_log_login")
	LogEventTracks = Session1.DB(dbName_log).C("col_event_track")
	LogGameTimes = Session1.DB(dbName_log).C("col_game_time")
	LogWaters = Session1.DB(dbName_log).C("col_log_water")
	LogShareWaters = Session1.DB(dbName_log).C("col_share_water")
	BattleRooms = Session1.DB(dbName_log).C("s_battle_room_data")
	BounsLogs = Session1.DB(dbName_log).C("col_bonus_log")
	PddStats = Session1.DB(dbName_log).C("s_pdd_stats")
	CrashPlayerStats = Session1.DB(dbName_log).C("s_crash_player_stats")
	PlanePlayerStats = Session1.DB(dbName_log).C("s_plane_player_stats")
	LHDPlayerStats = Session1.DB(dbName_log).C("s_lhd_player_stats")
	LHDXxscStats = Session1.DB(dbName_log).C("s_lhd_xxsc_stats")
	LHDQsbnStats = Session1.DB(dbName_log).C("s_lhd_qsbn_stats")
	LHDLkyhStats = Session1.DB(dbName_log).C("s_lhd_lkyh_stats")

	TPPlayerStats = Session1.DB(dbName_log).C("s_tp_player_stats")
	TPRoomPlayerStats = Session1.DB(dbName_log).C("s_tp_room_stats")
	IpRecords = Session1.DB(dbName_log).C("col_user_ip_records")
	StatsLogs = Session1.DB(dbName_log).C("col_stats_logs")
	ScratchTicketss = Session1.DB(dbName_log).C("col_scratch_ticket")
	LogActivityMonitors = Session1.DB(dbName_log).C("col_activity_monitor")
	LogPlayShareDraws = Session1.DB(dbName_log).C("col_playshare_draw")
	ThirdPartyBalances = Session1.DB(dbName_log).C("col_third_party_balance")
	// crash
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
	CrashStrategies = Session.DB(dbName).C("col_crash_strategy")
	PlaneStrategies = Session.DB(dbName).C("col_plane_strategy")
	ChargeClassifys = Session.DB(dbName).C("col_charge_classify")
	LHDStrategies = Session.DB(dbName).C("col_lhd_strategy")

	// 7Updown
	UPStrategies = Session.DB(dbName).C("col_seven_strategy")
	UpdownPlayerStats = Session1.DB(dbName_log).C("s_updonw_player_stats")
	UpdownXxscStats = Session1.DB(dbName_log).C("s_updown_xxsc_stats")
	UpdownQsbnStats = Session1.DB(dbName_log).C("s_updown_qsbn_stats")
	UpdownLkyhStats = Session1.DB(dbName_log).C("s_updown_lkyh_stats")

	// AB
	ABStrategies = Session.DB(dbName).C("col_ab_strategy")
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

	// 统计
	DataSummarys = Session1.DB(dbName_log).C("s_data_summary")
	DataStatisticss = Session1.DB(dbName_log).C("s_data_statistics")
	ShareStatistcs = Session1.DB(dbName_log).C("s_share_data")
	UserRetaineds = Session1.DB(dbName_log).C("s_user_retained")
	PayUserRetaineds = Session1.DB(dbName_log).C("s_pay_user_retained")
	Playtimes = Session1.DB(dbName_log).C("s_playtime_analysis")
	PointDatas = Session1.DB(dbName_log).C("s_point_data")
	GameNumberAnalysiss = Session1.DB(dbName_log).C("s_game_number_analysis")
	ChannelDatas = Session1.DB(dbName_log).C("s_channel_data")
	RealTimes = Session1.DB(dbName_log).C("s_real_time_data")
	UserResources = Session1.DB(dbName_log).C("s_user_resource")
	GlobalResources = Session1.DB(dbName_log).C("s_global_resource")
	GameResources = Session1.DB(dbName_log).C("s_game_resource")
	ChannelRates = Session1.DB(dbName_log).C("s_channel_success_rate")
	ShareDatas = Session1.DB(dbName_log).C("s_share_activity_data")
	LogBugFeedbacks = Session1.DB(dbName_log).C("col_bug_feedback")
	RoomDatas = Session1.DB(dbName_log).C("s_room_data")
	BugDatas = Session1.DB(dbName_log).C("s_bug_statistics")
	GoodsBuys = Session1.DB(dbName_log).C("s_goods_buy_data")
	AdReports = Session1.DB(dbName_log).C("col_ad_report")
	AdReportDatas = Session1.DB(dbName_log).C("s_ad_report_data")
	LogSmsRecords = Session1.DB(dbName_log).C("col_sms_record")
	LogUserSmsRecords = Session1.DB(dbName_log).C("col_user_sms_record")
	SmsDatas = Session1.DB(dbName_log).C("s_sms_data")
	PaySources = Session1.DB(dbName_log).C("s_pay_source")
	ButtonClicks = Session1.DB(dbName_log).C("col_button_click")
	UserGameDatas = Session1.DB(dbName_log).C("s_user_game_data")
	UserGameStrategys = Session1.DB(dbName_log).C("s_user_game_strategy")
	LogButton2Clicks = Session1.DB(dbName).C("col_button_click_full")
	PayChannels = Session.DB(dbName).C("col_pay_channel")
	LotteryActivitys = Session1.DB(dbName_log).C("s_lottery_activity")
	PayChannelStatss = Session1.DB(dbName_log).C("s_pay_channel_stats")
	PayChannelLogs = Session1.DB(dbName_log).C("col_pay_channel_log")
	RankingLists = Session1.DB(dbName_log).C("s_ranking_list")
	FirstCharges = Session1.DB(dbName_log).C("s_first_charge_analysis")
	// 活动
	BasicPays = Session.DB(dbName).C("col_basic_pay")
	BasicFrees = Session.DB(dbName).C("col_basic_free")

	// 外接游戏
	NsqLogExternalBets = Session1.DB(dbName_log).C("col_nsq_log_external_bet")
	NsqLogExternalRewards = Session1.DB(dbName_log).C("col_nsq_log_external_reward")
	//init
	initService()

	//创建文件目录
	os.MkdirAll(GetFilePath(), 0755)
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
		beego.Error(err)
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

// 计时器
func ticker() {
	// 启动执行一次
	// go func() {
	// 	defer func() {
	// 		if err := recover(); err != nil {
	// 			beego.Error("error crash stats:", err)
	// 			debug.PrintStack()
	// 		}
	// 	}()

	// 	c := time.After(5 * time.Second)
	// 	<-c
	// 	now := time.Now()
	// 	for i := 0; i < 2; i++ {
	// 		startTime := time.Date(2025, 3, 20, 0, 0, 0, 0, now.Location()).AddDate(0, 0, i)
	// 		beego.Info("CrashStatsCK start: ", startTime)
	// 		PlainStatsCK(utils.Time2Stamp(startTime))
	// 		// PlaneStrategyStatsCK(utils.Time2Stamp(startTime))
	// 		CrashStatsCK(utils.Time2Stamp(startTime))
	// 		// CrashStrategyStatsCK(utils.Time2Stamp(startTime))
	// 		beego.Info("CrashStatsCK finish")
	// 	}
	// }()

	// 初始化定时统计任务
	initTimedStatistics()
	for {
		// now := time.Now()
		// // startTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		// nextRun := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 59, 59, 0, now.Location())
		// if now.After(nextRun) {
		// 	nextRun = nextRun.Add(time.Hour)
		// }
		// timeUntilNextRun := nextRun.Sub(now)
		// timer := time.NewTimer(timeUntilNextRun)

		// 15分钟检查一次
		timer := time.NewTimer(10 * time.Minute)
		<-timer.C // 等待定时器触发

		// 检查任务并执行
		curTime := time.Now().Unix()
		for i, taskName := range tickerTasks {
			running := tickerTasksRunning[i]
			// 执行中
			if running == -1 {
				// 任务执行时间过长???
				beego.Warn("task %s duplicate run", taskName)
				continue
			}
			// 未到执行时间
			if curTime < running+int64(tickerTasksDuration[i].Seconds()) {
				continue
			}
			tickerTasksRunning[i] = -1

			task, ok := tickerTasksMap[taskName]
			if !ok {
				beego.Error("task %s not exists", taskName)
				continue
			}

			go func(index int, name string, f func(timestamp int64)) {
				// beego.Infof("task wait %s", name)
				if !tickerPriorityTasks[name] { // 非优先任务排队
					tickerTasksCh <- struct{}{}
				}
				defer func() {
					if err := recover(); err != nil {
						beego.Error(fmt.Sprintf("task %s exec panic %v", name, err))
						debug.PrintStack() // 打印堆栈
					}
					tickerTasksRunning[index] = time.Now().Unix()
					if !tickerPriorityTasks[name] {
						<-tickerTasksCh
					}
				}()

				// 当前执行时间
				now := time.Now()
				sec := now.Unix()
				startTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

				// 检查昨天的最后统计是否执行
				loc, err := time.LoadLocation(locationName)
				if err != nil {
					loc = time.UTC
					beego.Error(fmt.Sprintf("load india location error: %v", err))
				}
				indiaTime := now.In(loc)
				yesterdayDate := indiaTime.AddDate(0, 0, -1).Format("2006-01-02")
				yesterdayLog := StatsLogService.GetByLogDate(name, yesterdayDate)
				if yesterdayLog == nil {
					startTime = startTime.AddDate(0, 0, -1)
					beego.Debug(fmt.Sprintf("task exec yesterday %s, %v", name, startTime))
				}

				beego.Debug(fmt.Sprintf("task exec start %s, %v", name, startTime))
				// 执行定时统计查询
				f(utils.Time2Stamp(startTime))

				// 保存昨天的最后统计执行记录
				if yesterdayLog == nil {
					StatsLogService.SaveStatsLog(&entity.StatsLog{
						StatsName: name,
						Date:      yesterdayDate,
						Stime:     now,
						Ctime:     time.Now(),
					})
				}

				beego.Debug(fmt.Sprintf("task exec finish  %s use %ds, %v", name, time.Now().Unix()-sec, startTime))
			}(i, taskName, task)
		}
	}
}

var (
	tickerTasksConcurrent = 5                                          // 同时执行任务并发数
	tickerTasksCh         = make(chan struct{}, tickerTasksConcurrent) // 并发限制
	tickerTasks           []string                                     // 任务列表
	tickerTasksRunning    []int64                                      // 任务运行标识：-1运行中,0未执行,>0上次执行时间秒
	tickerTasksDuration   []time.Duration                              // 任务运行延迟时间
	tickerTasksMap        = make(map[string]func(timestamp int64))     // map<taskName task>
	tickerPriorityTasks   = map[string]bool{                           // 优先任务，不排队
		"UserGameDataCK":     true, // 用户局数统计
		"CrashStats":         true, // 游戏统计
		"CrashStrategyStats": true,
		"PlainStats":         true,
		"PlaneStrategyStats": true,
		"TPStats":            true,
		"TPRoomStats":        true,
		"LHDStatsCK":         true,
		"UpdownStatsCK":      true,
		"ABStatsCK":          true,
		"CPStatsCK":          true,
		"RBStatsCK":          true,
	}
)

func addTask(taskName string, dur time.Duration, task func(timestamp int64)) {
	tickerTasks = append(tickerTasks, taskName)
	tickerTasksRunning = append(tickerTasksRunning, time.Now().Unix())
	tickerTasksDuration = append(tickerTasksDuration, dur)
	tickerTasksMap[taskName] = task
}

func addHourTask(taskName string, task func(timestamp int64)) {
	addTask(taskName, time.Hour, task)
}

func initTimedStatistics() {
	// 数据汇总统计
	addHourTask("DataStatistics4Web", DataStatistics4Web)
	// // 数据汇总(旧)
	// addHourTask("DataSummary", DataSummary)
	// // 渠道数据(旧)
	// addHourTask("ChannelData", ChannelData)
	addHourTask("RealTimeData", RealTimeData)
	addHourTask("BasicPayActivity", BasicPayActivityCK)
	addHourTask("BasicFreeActivity", BasicFreeActivity)
	// addHourTask("UserRetained", UserRetained)
	// addHourTask("UserRetainedByChannel", UserRetainedByChannel)
	// addHourTask("UserRetainedNew", UserRetainedCK)
	addHourTask("PayUserRetainedNew", PayUserRetainedCK)
	// addHourTask("UserResource", UserResource)
	// addHourTask("GlobalResource", GlobalResource)
	// addHourTask("GameResource", GameResource)
	// 支付渠道成功率
	addHourTask("PayChannelRateCK", PayChannelRateCK)
	addHourTask("ShareActivityData", ShareActivityData)
	addHourTask("PointData", PointData)
	addHourTask("GameNumberAnalysisData", GameNumberAnalysisData)
	addHourTask("BugCommitData", BugCommitData)
	addHourTask("PlaytimeAnalysisData", PlaytimeAnalysisData)
	// 房间数据
	addHourTask("RoomDataCK", RoomDataCK)
	addHourTask("CheckIPDuplication", CheckIPDuplication)
	// addHourTask("PayUserRetained", PayUserRetained)
	// addHourTask("PayUserRetainedByChannel", PayUserRetainedByChannel)
	// 商品购买统计
	addHourTask("GoodsBuyData", GoodsBuyData)
	addHourTask("AdReportStatistics", AdReportStatistics)
	// ShareStatistics(timestamp)
	// 短信统计
	addHourTask("SMSData", SMSData)
	// 充值来源
	addHourTask("PaySourceData", PaySourceData)
	// // 对战房埋点统计
	// addHourTask("BattleRoomData", BattleRoomData)
	// 彩票活动统计
	addHourTask("LotteryActivityData", LotteryActivityData)
	// 用户游戏局数统计
	// addTask("UserGameData", time.Minute*15, UserGameData)
	addTask("UserGameDataCK", time.Minute*15, UserGameDataCK)
	// 定时删除前端按钮全量埋点数据
	addTask("LogButton2ClickClean", time.Hour*3, LogButton2ClickClean)
	// 拼多多活动定时统计
	addTask("PDDStats", time.Minute*30, PDDStats)
	// 支付渠道统计
	addHourTask("PayChannelStats", PayChannelStats)
	// crash 游戏统计
	addTask("CrashStats", time.Minute*15, CrashStatsCK)
	addTask("CrashStrategyStats", time.Minute*15, CrashStrategyStatsCK)
	// plain 游戏统计
	addTask("PlainStats", time.Minute*15, PlainStatsCK)
	addTask("PlaneStrategyStats", time.Minute*15, PlaneStrategyStatsCK)
	// // crash 扶摇直上
	// addTask("CrashFyzsStats", time.Minute*30, CrashFyzsStats)
	// // crash 欲薅无门
	// addTask("CrashYhwmStats", time.Minute*30, CrashYhwmStats)
	// tp 游戏统计
	addTask("TPStats", time.Minute*15, TPStatsCK)
	// tp房间数据
	addTask("TPRoomStats", time.Minute*15, TPRoomStatsCK)
	// // crash 起死回生、奖池风控
	// addTask("CrashQshsOrJcfkStats", time.Minute*30, CrashQshsOrJcfkStats)
	// 龙虎游戏统计
	addTask("LHDStatsCK", time.Minute*15, LHDStatsCK)

	// 7updown游戏统计
	addTask("UpdownStatsCK", time.Minute*15, UpdownStatsCK)

	// AB游戏统计
	addTask("ABStatsCK", time.Minute*15, ABStatsCK)
	// CP游戏统计
	addTask("CPStatsCK", time.Minute*15, CPStatsCK)
	// 红黑游戏统计
	addTask("RBStatsCK", time.Minute*15, RBStatsCK)

	// 首充分析
	addHourTask("FirstChargeAnalysis", FirstChargeAnalysis)
}

// func TimedStatistics(timestamp int64, endstamp int64) {
// 	go func() {
// 		// 数据汇总统计
// 		DataStatistics4Web(timestamp)

// 		// 数据汇总(旧)
// 		DataSummary(timestamp)

// 		// 渠道数据(旧)
// 		ChannelData(timestamp)

// 		RealTimeData(endstamp)

// 		BasicPayActivity(timestamp)

// 		BasicFreeActivity(timestamp)

// 		UserRetained(timestamp)

// 		UserRetainedByChannel(timestamp)

// 		UserResource(timestamp)

// 		GlobalResource(timestamp)

// 		GameResource(timestamp)

// 		// 支付渠道成功率
// 		PayChannelRate(timestamp)

// 		ShareActivityData(timestamp)

// 		PointData(timestamp)

// 		GameNumberAnalysisData(timestamp)

// 		BugCommitData(timestamp)

// 		PlaytimeAnalysisData(timestamp)

// 		// 房间数据
// 		RoomData(timestamp)

// 		CheckIPDuplication(timestamp)

// 		PayUserRetained(timestamp)

// 		PayUserRetainedByChannel(timestamp)

// 		// 商品购买统计
// 		GoodsBuyData(timestamp)

// 		AdReportStatistics(timestamp)

// 		// 分享数据统计
// 		// ShareStatistics(timestamp)

// 		// 短信统计
// 		SMSData(timestamp)

// 		// 充值来源
// 		PaySourceData(timestamp)
// 		// 对战房埋点统计
// 		BattleRoomData(timestamp)

//			glog.Info("统计后台执行定时统计任务,时间：", endstamp)
//		}()
//	}
//
// 计时器
func tickerExecuteDay() {
	// now := time.Now()
	// // 计算距离下一个0点的时间间隔
	// nextMidnight := now.Add(time.Hour * 24)
	// nextMidnight = time.Date(nextMidnight.Year(), nextMidnight.Month(), nextMidnight.Day(), 0, 0, 0, 0, nextMidnight.Location())
	// // 计算等待时间
	// waitDuration := nextMidnight.Sub(now)
	// // 创建定时器
	// timer := time.NewTimer(waitDuration)
	// // 启动goroutine等待定时器的触发
	// go func() {
	// 	<-timer.C
	// 	// 定时器触发后执行任务
	// 	fmt.Println("开始执行每天0点的任务...")
	// 	// 获取渠道0点余额
	// 	GetPayChannelBalance()

	// 	RankingList(utils.TimestampYesterday())
	// 	UserGameData(utils.TimestampYesterday())
	// 	//重新设置定时器，以便明天的0点再次触发
	// 	timer.Reset(calculateDurationUntilMidnight())
	// }()

	// for {
	// 	now := time.Now()
	// 	// Calculate the start of the next day
	// 	startOfNextDay := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	// 	// Calculate the time until next midnight
	// 	timeUntilNextMidnight := startOfNextDay.Sub(now)

	// 	timer := time.NewTimer(timeUntilNextMidnight)
	// 	<-timer.C // Wait for the timer to trigger at midnight
	// 	ExecuteDayStatistics()
	// }

	// 获取 IST 时区
	// ist, err := time.LoadLocation(timeSetting)
	// if err != nil {
	// 	fmt.Println("Error loading IST location:", err)
	// 	return
	// }

	// for {
	// 	now := time.Now().In(ist)                                                                     // 获取当前 IST 时间
	// 	nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC).In(ist) // 计算下一个0点0分的时间

	// 	// 计算到下一个午夜的时间间隔
	// 	timeUntilNextMidnight := nextMidnight.Sub(now)

	// 	timer := time.NewTimer(timeUntilNextMidnight)
	// 	<-timer.C              // 等待计时器触发
	// 	ExecuteDayStatistics() // 执行每日统计任务
	// }

	// 补数据
	ExecuteDayStatistics() // 执行每日统计任务
}

func ExecuteDayStatistics() {
	// beego.Info(fmt.Sprintf("ExecuteDayStatistics开始执行，执行日期：%d", utils.TimestampYesterday()))
	beego.Info("ExecuteDayStatistics开始执行")
	// GetPayChannelBalance()
	// RankingList(utils.TimestampYesterday())
	// PayChannelRateCK(utils.TimestampYesterday())
	// DataStatistics4Web(utils.TimestampYesterday(location))
	// // UserRetainedCK(utils.TimestampYesterday())
	// PayUserRetainedCK(utils.TimestampYesterday())
	// PayChannelStats(utils.TimestampYesterday())
	// RoomDataCK(utils.TimestampYesterday())

	// 数据汇总补数据
	// UpdownStatsCK(GetDayTimestamp("2025-05-12"))

	// DataStatistics4Web(GetDayTimestamp("2025-05-01"))
	// var dayTimes = []int64{
	// 	GetDayTimestamp("2025-04-25"),
	// 	GetDayTimestamp("2025-04-26"),
	// 	GetDayTimestamp("2025-04-27"),
	// }
	// for _, dayTime := range dayTimes {
	// 	RealTimeData(dayTime)
	// 	BasicPayActivityCK(dayTime)
	// 	BasicFreeActivity(dayTime)
	// 	PayUserRetainedCK(dayTime)
	// 	PayChannelRateCK(dayTime)
	// 	ShareActivityData(dayTime)
	// 	PointData(dayTime)
	// 	GameNumberAnalysisData(dayTime)
	// 	BugCommitData(dayTime)
	// 	PlaytimeAnalysisData(dayTime)
	// 	RoomDataCK(dayTime)
	// 	CheckIPDuplication(dayTime)
	// 	GoodsBuyData(dayTime)
	// 	AdReportStatistics(dayTime)
	// 	SMSData(dayTime)
	// 	PaySourceData(dayTime)
	// 	LotteryActivityData(dayTime)
	// 	UserGameDataCK(dayTime)
	// 	LogButton2ClickClean(dayTime)
	// 	PDDStats(dayTime)
	// 	PayChannelStats(dayTime)
	// 	CrashStatsCK(dayTime)
	// 	CrashStrategyStatsCK(dayTime)
	// 	PlainStatsCK(dayTime)
	// 	PlaneStrategyStatsCK(dayTime)
	// 	TPStatsCK(dayTime)
	// 	TPRoomStatsCK(dayTime)
	// 	LHDStatsCK(dayTime)
	// 	UpdownStatsCK(dayTime)
	// 	ABStatsCK(dayTime)
	// 	CPStatsCK(dayTime)
	// 	RBStatsCK(dayTime)
	// 	FirstChargeAnalysis(dayTime)
	// }

	beego.Info("ExecuteDayStatistics执行完成!")
}

func GetDayTimestamp(day string) int64 {
	s := fmt.Sprintf("%s 00:00:00", day)
	ts := utils.Str2TimeZone(s, locationName)
	return ts.Unix()
}

// 获取渠道记录余额
func GetPayChannelBalance() {
	now := time.Now()
	b := make(map[string]string, 0)
	b["channelId"] = "-1"
	response, err := PayGetRequest(b, "/querybalance")
	if err == nil {
		if response.Code == 200 {
			if response.Body != nil {
				resp := make([]entity.WebQueryBalence, 0)
				err := json.Unmarshal(response.Body, &resp)
				if err != nil {
					beego.Error("GetPayChannelBalance Parsing json err :", err)
				}
				if len(resp) > 0 {
					// 获取我方渠道余额
					var list []entity.PayChannel
					PayChannels.Find(bson.M{}).All(&list)
					// 将获取到的支付渠道余额记录下来
					// addlist := make([]entity.ThirdPartyBalance, 0)
					for _, item := range resp {
						addinfo := new(entity.ThirdPartyBalance)
						addinfo.PayChannel = item.ChannelId
						addinfo.Balance = item.Balence
						str := fmt.Sprintf("%d", item.ChannelId)
						diff := int64(0)
						for _, v := range list {
							if v.Id == str {
								addinfo.OwnBalance = int64(v.Balance)
								diff = addinfo.Balance - addinfo.OwnBalance
								if diff != 0 {
									v.Balance = int(addinfo.Balance)
									err := PayService.UpdatePayChannelByBalance(&v)
									if err != nil {
										beego.Error("GetPayChannelBalance Add err :", err)
									}
								}
								break
							}
						}

						addinfo.Difference = diff
						addinfo.Ctime = now.Unix()
						err := PayService.AddThirdPartyBalance(addinfo)
						if err != nil {
							beego.Error("GetPayChannelBalance Add err :", err)
						}
					}
				}
			}

		}
	}
}

// // 计算距离下一个0点的时间间隔
// func calculateDurationUntilMidnight() time.Duration {
// 	now := time.Now()
// 	nextMidnight := now.Add(24 * time.Hour)
// 	nextMidnight = time.Date(nextMidnight.Year(), nextMidnight.Month(), nextMidnight.Day(), 0, 0, 0, 0, nextMidnight.Location())
// 	return nextMidnight.Sub(now)
// }

var isfirst = false

func DataAcquisition() {
	beego.Info("DataAcquisition开始执行!")
	startDate := time.Date(2024, time.September, 21, 0, 0, 0, 0, time.Local)
	nowTime := time.Now().AddDate(0, 0, 1)
	endDate := time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), 0, 0, 0, 0, time.Local)
	diff := endDate.Sub(startDate)
	daysDiff := int(diff.Hours() / 24)

	fmt.Printf("两个日期相差 %d 天\n", daysDiff)
	if daysDiff > 1 {
		// 删除往期数据
		// UserGameDatas.RemoveAll(bson.M{})
		// RankingLists.RemoveAll(bson.M{})
	}
	for i := 0; i < daysDiff; i++ {
		timeStamp := utils.Time2Stamp(startDate.AddDate(0, 0, i))
		// UserRetainedCK(timeStamp)
		PayUserRetainedCK(timeStamp)
		// UserGameDataCK(timeStamp)
		// RankingList(timeStamp)
		// RankingList(timeStamp)
		// PayUserRetainedCK(timeStamp)
		beego.Info("timeStamp：", timeStamp)
	}
	beego.Info("DataAcquisition执行完成!")
	// CrashStatsCK(utils.TimestampYesterday())
	// UserGameDataFullCK()
	// beego.Info("DataAcquisition执行完成!")
	// ABStatsCK(utils.TimestampYesterday())
	// BasicPayActivityCK(utils.TimestampYesterday())

	// UpdownStatsCK(utils.TimestampToday())
	// beego.Info(fmt.Sprintf("ExecuteDayStatistics开始执行，执行日期：%d", utils.TimestampYesterday()))

	// beego.Info("DataAcquisition开始执行!")
	// startDate := time.Date(2024, time.August, 15, 0, 0, 0, 0, time.Local)
	// nowTime := time.Now().AddDate(0, 0, 1)
	// endDate := time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), 0, 0, 0, 0, time.Local)
	// diff := endDate.Sub(startDate)
	// daysDiff := int(diff.Hours() / 24)

	// fmt.Printf("两个日期相差 %d 天\n", daysDiff)
	// if daysDiff > 1 {
	// 	// 删除往期数据
	// 	// UserGameDatas.RemoveAll(bson.M{})
	// 	// RankingLists.RemoveAll(bson.M{})
	// }
	// for i := 0; i < daysDiff; i++ {
	// 	timeStamp := utils.Time2Stamp(startDate.AddDate(0, 0, i))
	// 	// UserGameDataCK(timeStamp)
	// 	// RankingList(timeStamp)
	// 	RoomDataCK(timeStamp)
	// 	beego.Info("timeStamp：", timeStamp)
	// }
	// beego.Info("DataAcquisition执行完成!")

	// FirstChargeAnalysis(utils.TimestampYesterday())
	// FirstChargeAnalysis(utils.TimestampToday())
	// // CrashQshsOrJcfkStats(utils.TimestampToday())
	// beego.Info("CrashQshsOrJcfkStats执行完成！！！")

	// UserGameDataFull()
	// dataArr := []int64{
	// 	1717093800, 1717180200, 1717266600, 1717353000,
	// }
	// beego.Info("开始补充数据汇总数据")
	// for _, r := range dataArr {
	// 	beego.Info("执行日期：", r)
	// 	DataStatistics4Web(r)

	// }
	// beego.Info("数据汇总数据补充完成！！！")

	// PayUserRetained(utils.TimestampToday())
	// PayUserRetainedByChannel(utils.TimestampToday())

	// CrashYhwmStats(utils.TimestampToday())
	// 创建1分钟定时器
	// fiveMinuteTicker := time.NewTicker(1 * time.Minute)
	// go func() {
	// 	ranges := [][2]string{
	// 		{"2024-05-11", "2024-05-11"},
	// 		{"2024-05-12", "2024-05-12"},
	// 		{"2024-05-13", "2024-05-13"},
	// 		{"2024-05-14", "2024-05-14"},
	// 		{"2024-05-15", "2024-05-15"},
	// 		{"2024-05-16", "2024-05-16"},
	// 		{"2024-05-17", "2024-05-17"},
	// 		{"2024-05-18", "2024-05-18"},
	// 		{"2024-05-19", "2024-05-19"},
	// 		{"2024-05-20", "2024-05-20"},
	// 	}
	// 	for range fiveMinuteTicker.C {
	// 		if !isfirst {
	// 			// 只允许执行一次
	// 			isfirst = true
	// 			for _, r := range ranges {
	// 				UserGameDataByA1_5(r[0], r[1])
	// 				beego.Info("执行完成 UserGameDataByA1_5")
	// 				UserGameDataByB1_5(r[0], r[1])
	// 				beego.Info("执行完成 UserGameDataByB1_5")

	// 				// beego.Info("开始执行DataAcquisition", r[0], r[1])
	// 				// UserGameInningsByA(r[0], r[1])

	// 				// UserGameInningsByB(r[0], r[1])
	// 				// beego.Info("执行完成DataAcquisition", r[0], r[1])
	// 			}
	// 		}
	// 	}

	// }()
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

// 筹码转换为分展示
func Chip2Float(chip int64) float64 {
	return (float64(chip) / 100)
}

// 昨日凌晨4点30分
func TimestampYesterday4(loc *time.Location) int64 {
	return TimestampToday4(loc) - 86400
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
		return utils.Stamp2Time(TimestampYesterday4(location))
	}
	//今日凌晨4点30分
	return utils.Stamp2Time(t4)
}

// ConvertToIndiaTime 将给定的时间戳转换为印度时区的时间，并返回该时间
func ConvertToIndiaTime(timestamp int64) (time.Time, error) {
	location, err := time.LoadLocation(locationName)
	if err != nil {
		// return time.Time{}, err
		return time.Unix(timestamp, 0), err
	}
	return time.Unix(timestamp, 0).In(location), nil
}

// 毫秒时间戳转成印度时间
func ConvertToIndiaTime1(timestamp int64) (time.Time, error) {
	indiaTimeZone, err := time.LoadLocation(locationName)
	if err != nil {
		fmt.Println("无法加载印度时区:", err)
		return time.Time{}, err
	}
	// 毫秒时间戳
	millisecondTimestamp := int64(timestamp)
	// 将毫秒时间戳转换为秒级时间戳
	secondTimestamp := millisecondTimestamp / 1000
	// 根据秒级时间戳创建时间对象
	return time.Unix(secondTimestamp, 0).In(indiaTimeZone), nil
}
