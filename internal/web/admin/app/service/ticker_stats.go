package service

import (
	"fmt"
	"goserver/internal/web/admin/app/entity"
	"goserver/pkg/data/ck"
	"goserver/pkg/utils"
	"runtime/debug"
	"sync"
	"time"

	"github.com/astaxie/beego"
	"github.com/globalsign/mgo/bson"
)

func InitTickerStats() {
	stat := new(TickerStat)
	stat.Init()
	beego.Info("ticker stats initialed.")
}

type TickerTask struct {
	Name     string
	Interval time.Duration
	Stat     func(TickerTaskArg) error
	Ticker   int64
}

type TickerTaskArg struct {
	CurDate   uint32 // 20250429
	CurMinute uint32 // 235959 -> 23:59:59
	Now       time.Time
	DayStime  time.Time
	DayEtime  time.Time
}

// 定时统计任务
type TickerStat struct {
	sync.Mutex
	tickerStats []*TickerTask
}

func (s *TickerStat) Init() {
	s.addTask("FinanceWalletStat", 10*time.Minute, FinanceWalletStat)

	s.Run()
}

func (s *TickerStat) addTask(name string, interval time.Duration, stat func(TickerTaskArg) error) {
	s.tickerStats = append(s.tickerStats, &TickerTask{
		Name:     name,
		Interval: interval,
		Stat:     stat,
		Ticker:   0,
	})
}

func (s *TickerStat) Run() {
	for _, task := range s.tickerStats {
		go func(task *TickerTask) {
			ticker := time.NewTicker(time.Second)
			for ; ; <-ticker.C {
				task.Ticker++
				now := time.Now().In(location)
				minuteTime := now.Hour()*10000 + now.Minute()*100 + now.Second()

				if task.Ticker < int64(task.Interval.Seconds()) && minuteTime < 235959 {
					continue
				}
				task.Ticker = 0

				func() {
					defer func() {
						if r := recover(); r != nil {
							beego.Error("recover ticker stat error:", task.Name, now, r)
							debug.PrintStack()
						}
					}()

					stime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
					etime := stime.AddDate(0, 0, 1).Add(-time.Second)
					dateTime := now.Year()*10000 + int(now.Month())*100 + now.Day()
					err := task.Stat(TickerTaskArg{
						Now:       now,
						DayStime:  stime,
						DayEtime:  etime,
						CurDate:   uint32(dateTime),
						CurMinute: uint32(minuteTime),
					})
					if err != nil {
						beego.Error("ticker task exec error: ", task.Name, now, err)
					}
				}()
			}
		}(task)
	}
}

// 经济日报钱包统计
func FinanceWalletStat(arg TickerTaskArg) (err error) {
	fmt.Printf("FinanceWalletStat exec: %v, %#v\n", arg.Now, arg.CurDate)

	var stats = make(map[string]*entity.FinanceWalletStat)

	// 所有玩家钱包汇总
	var wallet_datas []map[string]any
	err = ck.Select(&wallet_datas, `
		select ad__bundle_id, regist_area, SUM(diamond) cash, SUM(out_diamond) withdrawable, 
			(cash - withdrawable) non_withdrawable, SUM(vbbank) bonus
		from game.col_user s0 final
		group by ad__bundle_id, regist_area
	`)
	if err != nil {
		return
	}
	for _, data := range wallet_datas {
		ad__bundle_id := data["ad__bundle_id"].(string)
		regist_area := utils.ToInt64(data["regist_area"])
		cash := utils.ToInt64(data["cash"])
		withdrawable := utils.ToInt64(data["withdrawable"])
		non_withdrawable := utils.ToInt64(data["non_withdrawable"])
		bonus := utils.ToInt64(data["bonus"])

		id := fmt.Sprintf("%d-%d-%s", arg.CurDate, regist_area, ad__bundle_id)
		stat := &entity.FinanceWalletStat{
			Id:         id,
			Date:       arg.CurDate,
			BundleId:   ad__bundle_id,
			RegistArea: int32(regist_area),
			Ctime:      arg.Now,
		}
		stats[id] = stat

		stat.Cash = cash
		stat.Withdrawable = withdrawable
		stat.NonWithdrawable = non_withdrawable
		stat.Bonus = bonus
	}

	// 日活钱包统计
	var live_wallet_datas []map[string]any
	err = ck.Select(&live_wallet_datas, `
		select ad__bundle_id, regist_area, SUM(diamond) cash, SUM(out_diamond) withdrawable, (cash - withdrawable) non_withdrawable, SUM(vbbank) bonus
		from game.col_user s0 final
		join (
			select s1.userid 
			from game.col_log_login s1 final
			where s1.login_time between ? and ?
			group by s1.userid
		) s1 on s0.userid = s1.userid
		group by ad__bundle_id, regist_area
	`, arg.DayStime, arg.DayEtime)
	if err != nil {
		return
	}
	for _, data := range live_wallet_datas {
		ad__bundle_id := data["ad__bundle_id"].(string)
		regist_area := utils.ToInt64(data["regist_area"])
		cash := utils.ToInt64(data["cash"])
		withdrawable := utils.ToInt64(data["withdrawable"])
		non_withdrawable := utils.ToInt64(data["non_withdrawable"])
		bonus := utils.ToInt64(data["bonus"])

		id := fmt.Sprintf("%d-%d-%s", arg.CurDate, regist_area, ad__bundle_id)
		stat, ok := stats[id]
		if !ok {
			stat = &entity.FinanceWalletStat{
				Id:         id,
				Date:       arg.CurDate,
				BundleId:   ad__bundle_id,
				RegistArea: int32(regist_area),
			}
			stats[id] = stat
		}
		stat.LiveCash = cash
		stat.LiveWithdrawable = withdrawable
		stat.LiveNonWithdrawable = non_withdrawable
		stat.LiveBonus = bonus
	}

	for _, stat := range stats {
		Upsert(FinanceWalletStats, bson.M{"_id": stat.Id}, stat)
	}
	return
}
