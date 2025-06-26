package online

import (
	"context"
	"encoding/json"
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data/mq"
	"goserver/pkg/glog"
	"time"
)

const ONLINE_USER_KEY = "online_user_key"
const LAST_ACTIVE_KEY = "last_active_key"
const TIMEOUT_SEC = 30

type onlineUser struct {
	Userid string `json:"userid"`
	Gameid string `json:"gameid"`
	Roomid string `json:"roomid"`
}

func initNatsConsumer() {
	consumerName := "onlineuser"

	// 在线玩家处理
	mq.NatsCreateConsumer(consumerName, mq.StreamOnlineUser, mq.TopicOnlineUser, func() *pb.OnlineUserTTL { return new(pb.OnlineUserTTL) },
		func(f *pb.OnlineUserTTL) (err error) {
			onlineUser := &onlineUser{
				Userid: f.Userid,
				Gameid: f.GameId,
				Roomid: f.RoomId,
			}
			userJson, _ := json.Marshal(onlineUser)
			rdb.HSet(context.Background(), ONLINE_USER_KEY, onlineUser.Userid, userJson)
			_ = refreshActiveUser(onlineUser)
			return
		})
}

func offlineUser(userid string) {
	rdb.HDel(context.Background(), ONLINE_USER_KEY, userid)
	rdb.HDel(context.Background(), LAST_ACTIVE_KEY, userid)
}

func refreshActiveUser(user *onlineUser) error {
	return rdb.HSet(context.Background(), LAST_ACTIVE_KEY, user.Userid, time.Now().Unix()).Err()
}

func checkActiveUser() {
	glog.Infof("check user heartbeat")
	now := time.Now().Unix()
	users, err := rdb.HGetAll(context.Background(), LAST_ACTIVE_KEY).Result()
	if err != nil {
		glog.Errorf("checkActiveUser error: %v", err)
		return
	}
	for userID, lastActiveStr := range users {
		var lastActive int64
		fmt.Sscanf(lastActiveStr, "%d", &lastActive)
		if now-lastActive > TIMEOUT_SEC {
			offlineUser(userID)
		}
	}
}

func StartActiveUserChecker() {
	go func() {
		ticker := time.NewTicker(time.Duration(TIMEOUT_SEC) * time.Second)
		defer ticker.Stop()
		for {
			<-ticker.C
			checkActiveUser()
		}
	}()
}
