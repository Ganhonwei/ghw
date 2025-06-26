package robot

import (
	"context"
	"fmt"
	"goserver/pkg/glog"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
)

type HeadInfo struct {
	Id   uint32
	Name string
	Sex  uint32
}

var (
	redisClient *redis.Client
	manMutex    *redsync.Mutex
	womanMutex  *redsync.Mutex
)

func initHead() {
	redisAddr := cfg.Section("redis").Key("addr").Value()
	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       3,
	})

	pool := goredis.NewPool(redisClient)
	rs := redsync.New(pool)
	manMutex = rs.NewMutex("man-mutex")
	womanMutex = rs.NewMutex("woman-mutex")

	// getHead("man")
	// getHead("man")
	// getHead("man")
	// getHead("man")
	// getHead("man")
	// returnHead("man", 2096)
	if node == "1" {
		cleanHead("man", time.Now().Add(24*time.Hour).Unix())
		cleanHead("woman", time.Now().Add(24*time.Hour).Unix())
	}
}

func getHead(poolName string) *HeadInfo {
	var mutex *redsync.Mutex
	if poolName == "man" {
		mutex = manMutex
	} else if poolName == "woman" {
		mutex = womanMutex
	} else {
		return nil
	}

	if err := mutex.Lock(); err != nil {
		glog.Error(err)
		return nil
	}
	defer func() {
		if ok, err := mutex.Unlock(); !ok || err != nil {
			glog.Error(err)
		}
	}()

	res, err := redisClient.SPop(context.Background(), fmt.Sprintf("%s:set", poolName)).Result()
	if err != nil {
		glog.Error(err)
		return nil
	}
	if res == "" {
		glog.Errorf("spop res is empty")
		return nil
	}

	res, err = redisClient.HGet(context.Background(), fmt.Sprintf("%s:hash", poolName), res).Result()
	if err != nil {
		glog.Error(err)
		return nil
	}
	if res == "" {
		glog.Errorf("hget res is empty")
		return nil
	}

	var headInfo HeadInfo
	err = json.Unmarshal([]byte(res), &headInfo)
	if err != nil {
		glog.Error(err)
		return nil
	}

	expiry := float64(time.Now().Add(24 * time.Hour).Unix())
	_, err = redisClient.ZAdd(context.Background(), fmt.Sprintf("%s:zset", poolName), redis.Z{Score: expiry, Member: headInfo.Id}).Result()
	if err != nil {
		glog.Error(err)
	}

	return &headInfo
}

func returnHead(poolName string, id uint32) {
	var mutex *redsync.Mutex
	if poolName == "man" {
		mutex = manMutex
	} else if poolName == "woman" {
		mutex = womanMutex
	} else {
		return
	}

	if err := mutex.Lock(); err != nil {
		glog.Error("[returnHead],err:", err)
		return
	}
	defer func() {
		if ok, err := mutex.Unlock(); !ok || err != nil {
			glog.Error("[returnHead],err:", err)
		}
	}()

	err := redisClient.ZRem(context.Background(), fmt.Sprintf("%s:zset", poolName), id).Err()
	if err != nil {
		glog.Error("[returnHead],err:", err)
		return
	}

	err = redisClient.SAdd(context.Background(), fmt.Sprintf("%s:set", poolName), id).Err()
	if err != nil {
		glog.Error("[returnHead],err:", err)
		return
	}
}

func cleanHead(poolName string, during int64) {
	var mutex *redsync.Mutex
	if poolName == "man" {
		mutex = manMutex
	} else if poolName == "woman" {
		mutex = womanMutex
	} else {
		return
	}

	if err := mutex.Lock(); err != nil {
		glog.Error(err)
		return
	}
	defer func() {
		if ok, err := mutex.Unlock(); !ok || err != nil {
			glog.Error(err)
		}
	}()

	// now := float64()
	expiredObjects, err := redisClient.ZRangeByScore(context.Background(), fmt.Sprintf("%s:zset", poolName), &redis.ZRangeBy{
		Min: "-inf",
		Max: fmt.Sprintf("%d", during),
	}).Result()

	if err != nil {
		glog.Error(err)
		return
	}

	for _, id := range expiredObjects {
		err := redisClient.SAdd(context.TODO(), fmt.Sprintf("%s:set", poolName), id).Err()
		if err != nil {
			glog.Error(err)
		}
		err = redisClient.ZRem(context.TODO(), fmt.Sprintf("%s:zset", poolName), id).Err()
		if err != nil {
			glog.Error(err)
		}
	}

}
