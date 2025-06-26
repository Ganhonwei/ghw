package service

import (
	"fmt"
	"goserver/internal/web/stats/app/entity"

	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2/bson"
)

type statsLogService struct{}

// GetByLogDate log date
func (*statsLogService) GetByLogDate(statsName, date string) (log *entity.StatsLog) {
	var logs []*entity.StatsLog
	err := StatsLogs.
		Find(bson.M{"stats_name": statsName, "date": date}).
		All(&logs)
	if err != nil {
		beego.Error(fmt.Sprintf("GetByLogDate %s, %s error %v", statsName, date, err))
		return
	}
	if len(logs) > 0 {
		log = logs[0]
	}
	if len(logs) > 1 {
		beego.Warn(fmt.Sprintf("GetByLogDate duplicate %s, %s error %v", statsName, date, logs))
	}
	return
}

// SaveStatsLog log date
func (*statsLogService) SaveStatsLog(log *entity.StatsLog) {
	log.Id = bson.NewObjectId().Hex()
	if !Insert(StatsLogs, log) {
		beego.Error(fmt.Sprintf("insert stats log error %v", log))
	}
}
