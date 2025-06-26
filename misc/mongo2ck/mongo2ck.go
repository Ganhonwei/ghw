package main

import (
	"goserver/pkg/data/ck"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bytedance/sonic"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

func SyncMongo2Clickhouse[T any](syncName string, col *mgo.Collection, pipeM []bson.M, records []T, batchSize int, args ...int) {
	var page, total, total2 = 1, 0, 0
	var pct = 0

	if len(args) > 0 {
		total = args[0]
	} else {
		totalM := append(pipeM, bson.M{"$count": "total"})
		// 查询总条数
		var totalR []bson.M
		err := col.Pipe(totalM).All(&totalR)
		if err != nil || len(totalR) == 0 {
			glog.Errorf("%s query total count error: %d, %v", syncName, total, err)
			return
		}
		total = totalR[0]["total"].(int)
		if total == 0 {
			glog.Warningf("%s query total = 0", syncName)
			return
		}
	}

	pageM := append(pipeM, bson.M{"$skip": 0})
	pageM = append(pageM, bson.M{"$limit": batchSize})
	mlen := len(pageM)

	startSec := time.Now().Unix()
	// 分页查询
	for ; ; page++ {
		pageM[mlen-2]["$skip"] = (page - 1) * batchSize
		err := col.Pipe(pageM).All(&records)
		total2 += len(records)
		if err != nil {
			glog.Errorf("clickhouse parse and insert error: %d, %d, %v", page, batchSize, err)
			break
		}
		err = ck.ParseAndInsert(records)
		if err != nil {
			glog.Error("clickhouse parse and insert error: %d, %d, %v", page, batchSize, err)
			break
		}
		if cur := float64(total2) / float64(total) * 100; int(cur/1) > pct {
			pct++
			glog.Infof("%s, current %.2f%%, %d/%d", syncName, cur, total2, total)
		}

		if len(records) < batchSize {
			useSec := time.Now().Unix() - startSec
			useMin := useSec / 60
			glog.Infof("%s sync full %T finish success sync %d, %d, use %dm%ds", syncName, records, total, total2, useMin, useSec-useMin*60)
			break
		}
	}
}

// 全量同步 detail 按数据时间分片
func Syncmongo2ClickhouseByTime[T any](syncName string, col *mgo.Collection, newRecords func() []T, timeField string, begin, end time.Time, timeParse func(time.Time) any, onceTime int, unit time.Duration, concurrent int, filter bson.M) {
	if begin.After(end) {
		return
	}
	var channel = make(chan []T, concurrent)
	// var finish = make(chan struct{}, 0)
	// var finished int64
	var rwg = &sync.WaitGroup{}
	rwg.Add(concurrent)
	var wwg = &sync.WaitGroup{}
	wwg.Add(concurrent)

	var readTotal, writeTotal int64

	// 写
	for i := 0; i < concurrent; i++ {
		go func(thread int) {
			for {
				datas, ok := <-channel
				if !ok {
					wwg.Done()
					// if atomic.CompareAndSwapInt64(&finished, 0, 1) {
					// 	close(finish)
					// }
					break
				}
				if len(datas) == 0 {
					continue
				}
				ms := time.Now().UnixMilli()
				err := ck.ParseAndInsert(datas)
				if err != nil {
					glog.Error("write %s thread-%d clickhouse error:", syncName, thread, err)
					continue
				}
				atomic.AddInt64(&writeTotal, int64(len(datas)))
				glog.Infof("write %s thread-%d %d data use %dms, wtotal=%d", syncName, thread, len(datas), time.Now().UnixMilli()-ms, writeTotal)
			}
		}(i)
	}

	for i := 0; i < concurrent; i++ {
		filterM := bson.M{}
		if filter != nil {
			b, _ := sonic.Marshal(filter)
			_ = sonic.Unmarshal(b, &filterM)
		}

		go func(thread int, filter bson.M) {
			for i := 0; ; i++ {
				offset := thread*onceTime + concurrent*onceTime*i

				stime := begin.Add(time.Duration(offset) * unit)
				etime := begin.Add(time.Duration(offset+onceTime) * unit)
				if stime.After(end) {
					rwg.Done()
					break
				}
				var empty = newRecords()
				channel <- empty // 等写空闲,避免堆积数据

				// 查询
				var ms = time.Now().UnixMilli()
				var records = newRecords()
				filter[timeField] = bson.M{"$gte": timeParse(stime), "$lt": timeParse(etime)}
				err := col.Find(filter).All(&records)
				if err != nil {
					glog.Errorf("%s %s ~ %s error: %v", syncName, stime.Format(utils.FORMAT), etime.Format(utils.FORMAT), err)
					continue
				}
				atomic.AddInt64(&readTotal, int64(len(records)))
				glog.Infof("read_ %s thread-%d %s~%s: %d data use %dms, rtotal=%d", syncName, thread, stime.Format(utils.FORMAT), etime.Format(utils.FORMAT), len(records), time.Now().UnixMilli()-ms, readTotal)
				if len(records) > 0 {
					channel <- records
				}
			}
		}(i, filterM)
	}

	rwg.Wait()     // 等读
	close(channel) // 读完毕
	wwg.Wait()     // 等写

	glog.Infof("%s sync finish success rtotal=%d, wtotal=%d", syncName, readTotal, writeTotal)
}
