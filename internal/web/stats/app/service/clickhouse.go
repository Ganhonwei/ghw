package service

import (
	"goserver/pkg/data/ck"

	"github.com/astaxie/beego"
	"gorm.io/gorm"
)

func InitClickhouse() {
	logsql := beego.AppConfig.DefaultBool("clickhouse.logsql", false)
	addr := beego.AppConfig.String("clickhouse.addr")
	database := beego.AppConfig.String("clickhouse.database")
	username := beego.AppConfig.String("clickhouse.username")
	password := beego.AppConfig.String("clickhouse.password")
	dial_timeout := beego.AppConfig.DefaultInt("clickhouse.dial_timeout", 10)
	read_timeout := beego.AppConfig.DefaultInt("clickhouse.read_timeout", 20)
	ssh := beego.AppConfig.DefaultBool("clickhouse.ssh", false)
	sshUser := beego.AppConfig.String("clickhouse.sshUser")
	sshAddr := beego.AppConfig.String("clickhouse.sshAddr")
	sshKey := beego.AppConfig.String("clickhouse.sshKey")

	ck.InitClickhouse(ck.ClickhouseConfig{
		Logsql:      logsql,
		Addr:        addr,
		Database:    database,
		Username:    username,
		Password:    password,
		DialTimeout: dial_timeout,
		ReadTimeout: read_timeout,
		SSH:         ssh,
		SSHUser:     sshUser,
		SSHAddr:     sshAddr,
		SSHKey:      sshKey,
		DsnParams: map[string]any{
			"max_execution_time": 60,
			"max_query_size":     10000000, // sql最大长度单位byte
		},
	})
}

func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset, limit := PageCalc(page, pageSize)
		return db.Offset(offset).Limit(limit)
	}
}

func PageCalc(page, pageSize int) (offset, limit int) {
	if page <= 0 {
		page = 1
	}
	switch {
	case pageSize > 100000:
		pageSize = 100000
	case pageSize <= 0:
		pageSize = 20
	}
	offset = (page - 1) * pageSize
	return offset, pageSize
}
