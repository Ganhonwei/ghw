package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
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

func InitHead(addr string) {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       3,
	})

	pool := goredis.NewPool(redisClient)
	rs := redsync.New(pool)
	manMutex = rs.NewMutex("man-mutex")
	womanMutex = rs.NewMutex("woman-mutex")
}

func GetHead() *HeadInfo {
	if !utils.RandWan(8000) {
		return nil
	}

	poolName := "man"
	mutex := manMutex
	if utils.RandWan(1000) {
		poolName = "woman"
		mutex = womanMutex
	}

	// if poolName == "man" {
	// 	mutex = manMutex
	// } else if poolName == "woman" {
	// 	mutex = womanMutex
	// } else {
	// 	return nil
	// }

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

	hashPoolKey := fmt.Sprintf("%s:hash", poolName)
	res, err = redisClient.HGet(context.Background(), hashPoolKey, res).Result()
	if err != nil {
		glog.Errorf("get head error: %s, %s, %v", hashPoolKey, res, err)
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

func ReturnHead(sex uint32, id uint32) {
	var poolName string
	var mutex *redsync.Mutex
	if sex == 1 {
		poolName = "man"
		mutex = manMutex
	} else {
		poolName = "woman"
		mutex = womanMutex
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
