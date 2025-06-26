package betrank

import (
	"context"
	"encoding/base64"
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data/mq"
	"goserver/pkg/game/handler"
	"goserver/pkg/table"
	"goserver/pkg/utils"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func beforeInitNats() {
	natsUrl := "nats://@127.0.0.1:4222"
	if err := mq.InitNats(natsUrl); err != nil {
		panic(err)
	}
	redisAddr := "127.0.0.1:6379"
	var err error
	if rdb, err = InitRedis(redisAddr, 0); err != nil {
		panic(err)
	}

	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()
	// err := rdb.Del(ctx, betRankUsersKey, betRankDailyKey, betRankWeeklyKey, betRankMonthlyKey, jackpotDailyKey, jackpotWeeklyKey, jackpotMonthlyKey).Err()
	// if err != nil {
	// 	panic(err)
	// }
}

func TestRList(t *testing.T) {
	beforeInitNats()
	ctx := context.Background()
	// key := "persons"
	// r, err := rdb.HGet(ctx, key, "jack").Result()
	// fmt.Println(err, r)
	// r, err = rdb.HGet(ctx, key, "jacker").Result()
	// fmt.Println(err, r)
	// r, err = rdb.HGet(ctx, key+"xx", "jacker").Result()
	// fmt.Println(err, r)

	robotBases, err := rdb.HMGet(ctx, betRankRobotBaseKey, "307", "3077", "398").Result()
	fmt.Println(err, robotBases)
	for i, rb := range robotBases {
		if rb == nil {
			fmt.Println("nillll", i)
		}
	}
	// r, err := rdb.ZRevRangeWithScores(ctx, key, 0, 0).Result()
	// if err != nil {
	// 	t.Error(err)
	// 	return
	// }
	// fmt.Printf("%#v\n", r)
}

func TestBetRank(t *testing.T) {
	beforeInitNats()

	testPublishBets(t)
	testGetBetRankList(t)
}

func testPublishBets(t *testing.T) {
	bets := &pb.PublishGameBets{
		UserPid:    actor.NewPID("127.0.0.1:1234", "a04"),
		Userid:     "10001",
		Gtype:      int32(pb.HUA),
		Bets:       1000,
		Ts:         time.Now().Unix(),
		Username:   "T1",
		Photo:      "1",
		RegistArea: 0,
		VipLv:      2,
	}
	if err := mq.NatsPublish(mq.TopicGameBets, bets); err != nil {
		t.Error(err)
		return
	}
	bets.Userid = "10002"
	bets.Username = "T2"
	bets.Bets = 2000
	if err := mq.NatsPublish(mq.TopicGameBets, bets); err != nil {
		t.Error(err)
		return
	}
	bets.Userid = "10003"
	bets.Username = "T3"
	bets.Bets = 3000
	if err := mq.NatsPublish(mq.TopicGameBets, bets); err != nil {
		t.Error(err)
		return
	}
}

func testGetBetRankList(t *testing.T) {
	fmt.Println("5")
	req := &mq.RequestActivityBetRankListArgs{
		Userid:     "10001",
		RegistArea: 0,
	}
	rsp := &pb.ActivityBetRankRsp{}
	err := mq.NatsRequest(mq.RequestActivityBetRankList, rsp, req)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(rsp)
}

func TestTimes(t *testing.T) {
	// updateBetRankTimes(time.Now())
	InitBetRank()

	stime1 := betRankDailyEtime - int64((time.Hour * 24 * 1).Seconds()) + 1
	stime2 := betRankWeeklyEtime - int64((time.Hour * 24 * 7).Seconds()) + 1
	stime3 := time.Unix(betRankMonthlyEtime, 0).AddDate(0, -1, 0).Unix() + 1

	fmt.Println(time.Unix(stime1, 0).Format(utils.FORMAT), time.Unix(betRankDailyEtime, 0).Format(utils.FORMAT))
	fmt.Println(time.Unix(stime2, 0).Format(utils.FORMAT), time.Unix(betRankWeeklyEtime, 0).Format(utils.FORMAT))
	fmt.Println(time.Unix(stime3, 0).Format(utils.FORMAT), time.Unix(betRankMonthlyEtime, 0).Format(utils.FORMAT))
}

func TestLobby(t *testing.T) {
	beforeInitNats()
	InitBetRank()

	if err := mq.NatsPublish(mq.TopicGameToLobby, &pb.PublishGameToLobby{
		UserPid: actor.NewPID("127.0.0.1:1234", "a05"),
		Userid:  "10001",
		Ts:      time.Now().Unix(),
	}); err != nil {
		t.Error(err)
		return
	}
}

func TestMarshal(t *testing.T) {
	r := &pb.ActivityBetRankList{Jackpot: 10086}
	bytes, err := r.Marshal()
	if err != nil {
		t.Error(err)
		return
	}
	r1 := &pb.ActivityBetRankList{}
	r2 := &pb.ActivityBetRankList{Jackpot: 1}
	if err = r1.Unmarshal(bytes); err != nil {
		t.Error(err)
		return
	}
	if err = r2.Unmarshal(bytes); err != nil {
		t.Error(err)
		return
	}
	fmt.Println(r1.Jackpot)
	fmt.Println(r2.Jackpot)
}

func TestGetRobotBase(t *testing.T) {
	var err error
	//加载配置表
	err = table.LoadTables() // ../../../configs/table/json/
	if err != nil {
		panic(err)
	}

	if rdbHead, err = InitRedis("127.0.0.1:6379", 3); err != nil {
		panic(err)
	}

	for i := 0; i < 1000; i++ {
		name, photo, sex := GetUserBase(i)
		fmt.Println(i, name, photo, sex)
	}

	head := getHead("man")
	fmt.Println(head)
}

func TestLink(t *testing.T) {
	p := "share-8189359"
	shareStr := utils.Split(p, "-")
	superId := shareStr[1]
	shareSource := 0
	if shares := strings.Split(superId, "_"); len(shares) > 1 {
		superId = shares[0]
		shareSource, _ = strconv.Atoi(shares[1])
	}
	fmt.Println(superId, shareSource)
}

func TestBoom(t *testing.T) {
	var times, times101 int
	for range 2000 {
		times++
		r := rand.New(rand.NewSource(int64(time.Now().Nanosecond())))
		m := handler.CrashBoomMultiple(r)
		if m == 101 {
			times101++
		}
	}

	fmt.Printf("______测试次数=%d, 101倍次数=%d\n", times, times101)
	fmt.Printf("%f%%\n", float64(times101)/float64(times)*100)
}

func TestD(t *testing.T) {
	now := time.Now()
	m := handler.CrashTimeMultiple(now, now)
	fmt.Println(m)
	ts := handler.CrashBoomDuration(101)
	fmt.Println(ts.Nanoseconds())
}

func TestTurnPrizeBytes(t *testing.T) {
	var b = "Chg2ODUxMGFkOGY2ZmUwODAwMDE5YTg0NzcQARignAEgASoTMjAyNS0wNi0xNyAxMTo1NzozNjDYlcTCBg=="
	prize := new(pb.ActivityTurnPrize)
	prizeBytes, err := base64.StdEncoding.DecodeString(b)
	if err != nil {
		t.Error(err)
		return
	}
	err = prize.Unmarshal(prizeBytes)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Printf("%#v\n", prize)

	prize.State = 1
	bytes, _ := prize.Marshal()
	prizeB64 := base64.StdEncoding.EncodeToString(bytes)
	fmt.Println(prizeB64)
}
