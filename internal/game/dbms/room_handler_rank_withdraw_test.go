package dbms

// import (
// 	"encoding/base64"
// 	"fmt"
// 	"goserver/gen/pb"
// 	"goserver/pkg/data"
// 	"goserver/pkg/glog"
// 	"goserver/pkg/utils"
// 	"io"
// 	"strconv"
// 	"strings"
// 	"testing"
// 	"time"

// 	"github.com/gogo/protobuf/proto"
// 	ini "gopkg.in/ini.v1"
// )

// func initTest() {
// 	//日志定义
// 	// glog.Init()
// 	//加载配置
// 	cfg, err = ini.Load("../bin/config/conf.ini")
// 	if err != nil {
// 		panic(err)
// 	}
// 	host := cfg.Section("mongod").Key("host").Value()
// 	port := cfg.Section("mongod").Key("port").Value()
// 	user := cfg.Section("mongod").Key("user").Value()
// 	passwd := cfg.Section("mongod").Key("passwd").Value()
// 	dbname := cfg.Section("mongod").Key("name").Value()
// 	data.InitMgo(host, port, user, passwd, dbname)

// 	InitRedis()
// 	InitNsqProducer()
// 	InitNsqConsumer()
// }

// func TestRankWithdrawPublish(t *testing.T) {
// 	initTest()

// 	for i := 0; i < 30; i++ {
// 		msg := &pb.EventRankWithdraw{
// 			Robot:  utils.RandBool(),
// 			Userid: "test_" + strconv.Itoa(i),
// 			Amount: utils.RandInt32N(10000),
// 		}
// 		t.Logf("userid: %s, amount: %d", msg.Userid, msg.Amount)
// 		body2, _ := proto.Marshal(msg)
// 		producer.Publish(data.TopicRankWithdraw, body2)
// 	}

// 	time.Sleep(5 * time.Second)
// }

// func TestSave2Yesterday(t *testing.T) {
// 	initTest()

// 	fmt.Println("2")
// 	rankWithdrawTodayExpire()
// 	// rankWithdrawWeekExpire()
// }

// func TestRankWithdrawRandomRobotsClean(t *testing.T) {
// 	initTest()

// 	fmt.Println("4")
// 	rankWithdrawRandomRobotsClean()
// }

// func TestUserDecode(t *testing.T) {
// 	user := "CAESCTEyMzQ1NjcwMhoJQmlkeWFuYW5kIgEyKKCNBjgF"
// 	r := base64.NewDecoder(base64.StdEncoding, strings.NewReader(user))
// 	body, err := io.ReadAll(r)
// 	if err != nil {
// 		glog.Error("rank withdraw decode user err", err)
// 		t.Error(err)
// 		return
// 	}
// 	msg := &pb.EventRankWithdraw{}
// 	err = msg.Unmarshal(body)
// 	if err != nil {
// 		glog.Error("rank withdraw unmarshal user err", err)
// 		t.Error(err)
// 		return
// 	}
// 	fmt.Printf("user=%+v\n", msg)
// }

// func TestDrawNumbers(t *testing.T) {
// 	initTest()

// 	// data := &data.LuckyDrawNumber{
// 	// 	Id:     bson.NewObjectId().String(),
// 	// 	Round:  3,
// 	// 	Number: 451,
// 	// 	Userid: "100002",
// 	// 	Robot:  false,
// 	// 	Ctime:  time.Now(),
// 	// }
// 	// data.Save()

// 	numbers := data.GetDrawNumbersByRoundMN(1, 20)
// 	for _, num := range numbers {
// 		fmt.Printf("%#v\n", num)
// 	}
// 	fmt.Println("=================")
// 	roundUserNumbers := make(map[int64]map[string][]int32)
// 	var round int64
// 	for _, num := range numbers {
// 		if num.Round != round {
// 			round = num.Round
// 			roundUserNumbers[round] = make(map[string][]int32)
// 		}
// 		roundUserNumbers[round][num.Userid] = append(roundUserNumbers[round][num.Userid], num.Number)
// 	}
// 	fmt.Println(roundUserNumbers)
// }

// func TestDrawGives(t *testing.T) {
// 	// now := time.Now()
// 	now, err := time.Parse(utils.FORMAT, "2024-04-28 10:00:00")
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	today := now.Weekday()
// 	offset := 7 - today
// 	if today == time.Sunday {
// 		offset = 0
// 	}
// 	saturday := now.AddDate(0, 0, int(offset))
// 	weekStr := saturday.Format("2006-01-02")
// 	fmt.Println(weekStr)
// 	etime := utils.Str2Time(fmt.Sprintf("%s 23:59:59", weekStr))
// 	// a.rankWithdrawWeekEndTime = etime.Unix()
// 	fmt.Println(etime)
// }
