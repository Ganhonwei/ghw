package ck

import (
	"context"
	"database/sql"
	"fmt"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"goserver/pkg/zlog"
	"net"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"gopkg.in/ini.v1"
	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	clickhouse0 "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

var (
	dataParsers  *sync.Map
	db           *gorm.DB
	ckDB         *sql.DB
	conn         driver.Conn
	autoMigrates []any // 需要自动对齐的表
)

type CkEntity interface {
	New() CkEntity
	ParseData(interface{}) ([]interface{}, error)
}

func initRegistry() {
	dataParsers = &sync.Map{}
	// 注册类型转换
	RegistryData(new(data.User), new(User))                                             // 用户
	RegistryData(new(data.UserFinance), new(UserFinance))                               // 用户账户
	RegistryData(new(data.TradeRecord), new(TradeRecord))                               // 支付订单
	RegistryData(new(data.WithdrawRecord), new(WithdrawRecord))                         // 提现订单
	RegistryData(new(data.TransferOrderDetail), new(WithdrawRecordTransferOrderDetail)) // 提现订单转单详情
	RegistryData(new(data.LogWater), new(LogWater))                                     // 流水日志
	RegistryData(new(data.Detail), new(Detail))                                         // 对局详情表
	RegistryData(new(data.LogLogin), new(LogLogin))                                     // 登录日志
	RegistryData(new(data.NsqLogExternalBet), new(NsqLogExternalBet))
	RegistryData(new(data.NsqLogExternalReward), new(NsqLogExternalReward))
	RegistryData(new(data.NsqLogExternalCancel), new(NsqLogExternalCancel))
	RegistryData(new(data.ActivityBetRankPrize), new(ActivityBetRankPrize))                   // 打码排行榜活动
	RegistryData(new(data.ActivityTurnDrawLog), new(ActivityTurnDrawLog))                     // 转盘活动抽奖记录
	RegistryData(new(data.ActivityTurnPrizeLog), new(ActivityTurnPrizeLog))                   // 转盘活动领奖记录
	RegistryData(new(data.LaunchInviteLog), new(LaunchInviteLog))                             // 点击邀请按钮日志
	RegistryData(new(data.ShareAgentIncomeRecord), new(ShareAgentIncomeRecord))               // 代理打码/人头奖励记录表
	RegistryData(new(data.ShareAgentIncomeRecordTackLog), new(ShareAgentIncomeRecordTackLog)) // 代理打码/人头奖励领取记录表
	RegistryData(new(data.LogVbDiamond), new(LogVbDiamond))                                   // vbbank变化日志
	RegistryData(new(data.LogOutDiamond), new(LogOutDiamond))                                 // vbbank变化日志
	RegistryData(new(data.VolatilitySubsidy), new(VolatilitySubsidy))                         // 波动返水记录
	RegistryData(new(data.OnlineUsersLog), new(OnlineUsersLog))                               // 在线人数日志
	RegistryData(new(data.UserCoupon), new(UserCoupon))                                       // 满减券
	RegistryData(new(data.CumRechargeWheelUnlockLog), new(CumRechargeWheelUnlockLog))         // 累充转盘解锁记录
	RegistryData(new(data.CumRechargeWheelSpinLog), new(CumRechargeWheelSpinLog))             // 累充转盘抽奖记录

	autoMigrates = append(autoMigrates, new(CustomerChatSession)) // 客服聊天会话自动对齐
	autoMigrates = append(autoMigrates, new(CustomerChatMessage)) // 客服聊天消息自动对齐
	autoMigrates = append(autoMigrates, new(CustomerSeatLog))     // 客服坐席操作记录自动对齐

}

type ClickhouseConfig struct {
	Logsql      bool
	Addr        string
	Database    string
	Username    string
	Password    string
	DialTimeout int
	ReadTimeout int
	DsnParams   map[string]any
	SSH         bool
	SSHUser     string
	SSHAddr     string
	SSHKey      string
}

func GetConfigFromIni(cfg *ini.File) ClickhouseConfig {
	conf := ClickhouseConfig{}
	conf.Logsql = cfg.Section("clickhouse").Key("logsql").MustBool(false)
	conf.Addr = cfg.Section("clickhouse").Key("addr").Value()
	conf.Database = cfg.Section("clickhouse").Key("database").Value()
	conf.Username = cfg.Section("clickhouse").Key("username").Value()
	conf.Password = cfg.Section("clickhouse").Key("password").Value()
	conf.DialTimeout = cfg.Section("clickhouse").Key("dial_timeout").MustInt(10)
	conf.ReadTimeout = cfg.Section("clickhouse").Key("read_timeout").MustInt(20)
	return conf
}

func InitClickhouse(cfg ClickhouseConfig) {
	initRegistry()

	// init clickhouse connection
	var dsnParams string
	for name, value := range cfg.DsnParams {
		dsnParams += fmt.Sprintf("&%s=%v", name, value)
	}
	// clickhoue gorm
	if err := initClickhouseDB(cfg); err != nil {
		panic(err)
	}

	// clickhouse conn
	if err := initClickhouseConnect(cfg); err != nil {
		panic(err)
	}
}

func RegistryData(data any, target CkEntity) {
	t := reflect.TypeOf(data)
	dataParsers.Store(t, target)
	autoMigrates = append(autoMigrates, target)
}

func ParseData[T any](datas []T) (results []interface{}, err error) {
	if len(datas) == 0 {
		return []interface{}{}, nil
	}

	t := reflect.TypeOf(datas[0])
	target, ok := dataParsers.Load(t)
	if !ok {
		return nil, fmt.Errorf("type %T parser not found", datas[0])
	}
	for _, data := range datas {
		parser := target.(CkEntity).New()
		rs, err := parser.ParseData(data)
		if err != nil {
			return results, err
		}
		results = append(results, rs...)
	}
	return results, nil
}

// InitClickhouse 初始化ck
// https://github.com/go-gorm/clickhouse
// dsn: clickhouse://gorm:gorm@localhost:9942/gorm?dial_timeout=10s&read_timeout=20s
func initClickhouseDB(cfg ClickhouseConfig) (err error) {
	gormConfig := &gorm.Config{}
	if cfg.Logsql {
		gormConfig.Logger = logger.Default.LogMode(logger.Info) // 打印sql
	}

	// initial db
	ckOptions, err := getClickhouseOptions(cfg)
	if err != nil {
		return
	}
	ckDB = clickhouse0.OpenDB(ckOptions)
	db, err = gorm.Open(clickhouse.New(clickhouse.Config{Conn: ckDB}), gormConfig)
	if err != nil {
		return
	}
	return
}

func initClickhouseConnect(cfg ClickhouseConfig) (err error) {
	// initial conn
	ckOptions, err := getClickhouseOptions(cfg)
	if err != nil {
		return
	}

	conn, err = clickhouse0.Open(ckOptions)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = conn.Ping(ctx); err != nil {
		if exception, ok := err.(*clickhouse0.Exception); ok {
			err = fmt.Errorf("exception [%d] %s %s", exception.Code, exception.Message, exception.StackTrace)
		}
		return
	}
	return
}

func getClickhouseOptions(cfg ClickhouseConfig) (ckOptions *clickhouse0.Options, err error) {
	ckOptions = &clickhouse0.Options{
		Addr:        []string{cfg.Addr},
		DialTimeout: time.Duration(cfg.DialTimeout) * time.Second,
		ReadTimeout: time.Duration(cfg.ReadTimeout) * time.Second,
		Auth: clickhouse0.Auth{
			Database: "game",
			Username: cfg.Username,
			Password: cfg.Password,
		},
		ClientInfo: clickhouse0.ClientInfo{
			Products: []struct {
				Name    string
				Version string
			}{
				{Name: "game", Version: "0.1"},
			},
		},
		TLS: nil,
		// TLS: &tls.Config{
		// 	InsecureSkipVerify: true,
		// },
		// Settings: clickhouse0.Settings{
		// 	"max_execution_time": 60,
		// 	"max_query_size":     10000000, // sql最大长度单位byte
		// },
		Settings: clickhouse0.Settings(cfg.DsnParams),
		Compression: &clickhouse0.Compression{
			Method: clickhouse0.CompressionLZ4,
		},
		Debug: cfg.Logsql,
		Debugf: func(format string, v ...interface{}) {
			fmt.Printf(format, v)
		},
	}

	if cfg.SSH {
		k, e := os.ReadFile(cfg.SSHKey)
		if e != nil {
			return nil, e
		}
		signer, e := ssh.ParsePrivateKey(k)
		if e != nil {
			return nil, e
		}

		sshConfig := &ssh.ClientConfig{
			User: cfg.SSHUser,
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
		sshClient, e := ssh.Dial("tcp", cfg.SSHAddr, sshConfig)
		if e != nil {
			return nil, e
		}
		// 创建本地监听器 (本地端口)
		localListener, e := sshClient.Dial("tcp", cfg.Addr) // 远程 ClickHouse 地址及端口
		if e != nil {
			return nil, e
		}
		// 设置本地 dail 代理
		ckOptions.DialContext = func(ctx context.Context, addr string) (net.Conn, error) {
			return localListener, nil
		}
	}
	return
}

// AutoMigrateTables 自动对齐表结构，自动根据字段修改数据库表结构，只会加改不会删字段
func AutoMigrateTables() (err error) {
	if len(autoMigrates) == 0 {
		return
	}
	for _, entity := range autoMigrates {
		// Auto Migrate
		err = db.AutoMigrate(entity)
		if err != nil {
			glog.Errorf("type %T auto migrate clickhouse table error: %v", entity, err)
			zlog.Errorf("type %T auto migrate clickhouse table error: %v", entity, err)
			return
		}
	}
	// db.Set("gorm:table_options", "ENGINE=Distributed(cluster, default, hits)").AutoMigrate(&entity.TradeRecord{})
	return
}

func DB() *gorm.DB {
	return db
}

// sql查询数据
func Select(r interface{}, sql string, args ...any) (err error) {
	return db.Raw(sql, args...).Scan(r).Error
}

// 批量插入
func InsertBatch0[T any](datas []T) (err error) {
	if len(datas) == 0 {
		return
	}
	for _, data := range datas {
		// 优化原生批量插入: https://clickhouse.com/docs/en/integrations/go#batch-insert
		err = db.Create(data).Error
	}
	return
}

// 解析并插入ck
func ParseAndInsert[T any](datas []T) (err error) {
	results, err := ParseData(datas)
	if err != nil {
		return
	}
	err = InsertBatch(results)
	return
}

// 批量插入, 结构体用指针！
func InsertBatch[T any](datas []T) (err error) {
	if len(datas) == 0 {
		return
	}
	// conn.PrepareBatch()
	var tColumns = make(map[string][]string, 3)
	var tBatchs = make(map[string]driver.Batch, 3)

	for _, data := range datas {
		// 优化原生批量插入: https://clickhouse.com/docs/en/integrations/go#batch-insert
		// err = db.Create(data).Error
		table, columns, values, e := reflectGormData(data)
		if e != nil {
			err = e
			return
		}
		// 初始化该表 prepare
		if _, ok := tColumns[table]; !ok {
			tColumns[table] = columns

			prepareSql := fmt.Sprintf("INSERT INTO %s(", table)
			for i, column := range columns {
				if i == 0 {
					prepareSql += column
				} else {
					prepareSql += ("," + column)
				}
			}
			prepareSql += ") SETTINGS async_insert=1, wait_for_async_insert=0"
			batch, e := conn.PrepareBatch(context.Background(), prepareSql)
			if e != nil {
				err = e
				return
			}
			tBatchs[table] = batch
		}
		columns = tColumns[table]
		batch := tBatchs[table]
		var args = make([]any, 0, len(columns))
		for _, column := range columns {
			args = append(args, values[column])
		}
		err = batch.Append(args...)
		if err != nil {
			return
		}
	}
	// 批量插入
	for _, batch := range tBatchs {
		err = batch.Send()
		if err != nil {
			return
		}
	}
	return
}

func reflectGormData(data interface{}) (table string, columns []string, values map[string]interface{}, err error) {
	// value := reflect.ValueOf(data)
	// tableNameM := value.MethodByName("TableName")
	// _ = tableNameM.Call(nil)
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%T parse gorm data error %v", data, r)
		}
	}()

	values = make(map[string]interface{})
	refValue := reflect.ValueOf(data)
	refType := reflect.TypeOf(data)

	tableNameM := refValue.MethodByName("TableName")
	if !tableNameM.IsValid() {
		err = fmt.Errorf("type %T not has method TableName", data)
		return
	}
	rsp := tableNameM.Call(nil)
	table = rsp[0].Interface().(string)

	if refValue.Kind() == reflect.Ptr {
		refValue = refValue.Elem()
	}
	if refType.Kind() == reflect.Ptr {
		refType = refType.Elem()
	}

	for i := 0; i < refType.NumField(); i++ {
		field := refType.Field(i)
		value := refValue.Field(i).Interface()
		// value
		tag := field.Tag.Get("gorm")
		column := getGormTagColumnName(tag)
		if column == "" || strings.HasPrefix(column, "-") { // 忽略字段
			continue
		}
		columns = append(columns, column)
		values[column] = value
	}
	return
}

func getGormTagColumnName(tag string) (column string) {
	i1 := strings.Index(tag, "column:")
	if i1 < 0 {
		return
	}
	i2 := strings.Index(tag[i1+7:], ";")
	if i2 < 0 {
		return tag[i1+7:]
	}
	return tag[i1+7 : i2+7]
}
