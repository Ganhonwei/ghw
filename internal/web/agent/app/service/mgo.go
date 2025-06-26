package service

import (
	"fmt"
	"os"
	"time"

	"goserver/pkg/utils"

	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	locationName string
	location     *time.Location
)

// Init mgo and the common DAO

// 数据连接
var Session *mgo.Session
var Session1 *mgo.Session
var mgoCloseCh chan bool

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

// 日志库
var BloggerAccounts *mgo.Collection
var LogEventTracks *mgo.Collection
var LogGameTimes *mgo.Collection

// 统计
var DataStatisticss *mgo.Collection
var ShareStatistcs *mgo.Collection
var UserRetaineds *mgo.Collection
var PayUserRetaineds *mgo.Collection
var Playtimes *mgo.Collection
var PointDatas *mgo.Collection
var GameNumberAnalysiss *mgo.Collection

// 日志管理
var LoginLogs *mgo.Collection
var OutDiamonds *mgo.Collection
var Givediamonds *mgo.Collection
var IpRecords *mgo.Collection

// 初始化时连接数据库
func InitMgo() {
	// get db config from host, port, username, password
	dbHost := beego.AppConfig.String("mdb.host")
	dbPort := beego.AppConfig.String("mdb.port")
	dbUser := beego.AppConfig.String("mdb.user")
	dbPassword := beego.AppConfig.String("mdb.password")
	dbName := beego.AppConfig.String("mdb.name")

	var err error
	locationName = beego.AppConfig.String("timezone")
	location, err = time.LoadLocation(locationName)
	if err != nil {
		panic(err)
	}

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

	//url := "mongodb://" + usernameAndPassword + dbHost + ":" + dbPort + "/" + dbName

	// [mongodb://][user:pass@]host1[:port1][,host2[:port2],...][/database][?options]
	// mongodb://myuser:mypass@localhost:40001,otherhost:40001/mydb
	Session, err = mgo.Dial(url)
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

	// Optional. Switch the session to a monotonic behavior.
	Session1.SetMode(mgo.Monotonic, true)

	// 定时器
	go ticker()

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
	// 日志库
	RoleChannels = Session1.DB(dbName_log).C("col_blogger_channel")
	BloggerAccounts = Session1.DB(dbName_log).C("col_blogger_account")
	LoginLogs = Session1.DB(dbName_log).C("col_log_login")
	LogEventTracks = Session1.DB(dbName_log).C("col_event_track")
	LogGameTimes = Session1.DB(dbName_log).C("col_game_time")
	// 统计
	DataStatisticss = Session1.DB(dbName_log).C("dl_data_statistics")
	ShareStatistcs = Session1.DB(dbName_log).C("dl_share_data")
	UserRetaineds = Session1.DB(dbName_log).C("dl_user_retained")
	PayUserRetaineds = Session1.DB(dbName_log).C("dl_pay_user_retained")
	Playtimes = Session1.DB(dbName_log).C("dl_playtime_analysis")
	PointDatas = Session1.DB(dbName_log).C("dl_point_data")
	GameNumberAnalysiss = Session1.DB(dbName_log).C("dl_game_number_analysis")
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

// 计时器
func ticker() {
	for {
		now := time.Now()
		startTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		nextRun := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 59, 59, 0, now.Location())
		if now.After(nextRun) {
			nextRun = nextRun.Add(time.Hour)
		}
		timeUntilNextRun := nextRun.Sub(now)
		timer := time.NewTimer(timeUntilNextRun)
		// timer := time.NewTicker(1 * time.Minute)
		<-timer.C // 等待定时器触发
		TimedStatistics(utils.Time2Stamp(startTime), utils.Time2Stamp(nextRun))
	}
}

func TimedStatistics(timestamp int64, endstamp int64) {
	go func() {
		// 数据汇总统计
		DataStatistics(timestamp, "")

		// 分享数据统计
		ShareStatistics(timestamp, "")

		// 用户留存统计
		UserRetained(timestamp, "")

		// 付费用户留存统计
		PayUserRetained(timestamp, "")

		// 行为分析
		PointData(timestamp, "")
		// 局数分析
		GameNumberAnalysisData(timestamp, "")
		// 时间分析
		PlaytimeAnalysisData(timestamp, "")

		fmt.Println("代理后台执行定时统计任务,时间：", endstamp)
	}()
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
	startTime := utils.Str2Time(s, location)
	e := fmt.Sprintf("%s 23:59:59", endDate)
	endTime := utils.Str2Time(e, location)
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
	startTime := utils.Str2Time(s, location)
	e := fmt.Sprintf("%s 23:59:59", endDate)
	endTime := utils.Str2Time(e, location)
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
	startTime1 := utils.Str2Time(s1, location)
	e1 := fmt.Sprintf("%s 23:59:59", endDate1)
	endTime1 := utils.Str2Time(e1, location)
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
	s1 := utils.Str2Time(s, location)
	startTime := utils.Time2Stamp(s1)
	e := fmt.Sprintf("%s 23:59:59", endDate)
	s2 := utils.Str2Time(e, location)
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
	s1 := utils.Str2Time(s, location)
	startTime := utils.Time2StampToMS(s1)
	e := fmt.Sprintf("%s 23:59:59", endDate)
	s2 := utils.Str2Time(e, location)
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

// ConvertToIndiaTime 将给定的时间戳转换为印度时区的时间，并返回该时间
func ConvertToIndiaTime(timestamp int64) (time.Time, error) {
	return time.Unix(timestamp, 0).In(location), nil
}

func Location() *time.Location {
	return location
}
