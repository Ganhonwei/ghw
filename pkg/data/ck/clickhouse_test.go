package ck

import (
	"context"
	"fmt"
	"goserver/pkg/data"
	"goserver/pkg/utils"
	"log"
	"reflect"
	"strconv"
	"testing"
	"time"

	"gopkg.in/ini.v1"

	"github.com/segmentio/kafka-go"
)

var (
	dsn = "clickhouse://:@192.168.0.137:9000/game?dial_timeout=10s&read_timeout=20s"
)

func TestReflect(t *testing.T) {
	data := &TradeRecord{OrderID: "10086"}
	refValue := reflect.ValueOf(data)
	refType := reflect.TypeOf(data)

	tableNameM := refValue.MethodByName("TableName")
	// fmt.Println(tableNameM.IsZero())
	// fmt.Println(tableNameM.IsNil())
	fmt.Println(tableNameM.IsValid())
	rsp := tableNameM.Call(nil)
	fmt.Println(rsp)

	if refValue.Kind() == reflect.Ptr {
		refValue = refValue.Elem()
	}
	if refType.Kind() == reflect.Ptr {
		refType = refType.Elem()
	}

	for i := 0; i < refType.NumField(); i++ {
		field := refType.Field(i)
		value := refValue.Field(i)
		// value
		tag := field.Tag.Get("gorm")
		fmt.Printf("%s=%v\n", getGormTagColumnName(tag), value)
	}
}

func TestReflectGormData(t *testing.T) {
	data := &TradeRecord{OrderID: "10086"}
	table, columns, values, err := reflectGormData(data)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(table)
	fmt.Println(columns)
	fmt.Println(values)
}

func TestCkParser(t *testing.T) {
	//加载配置
	cfg, err := ini.Load("../../bin/config/conf.ini")
	if err != nil {
		t.Error(err)
	}
	InitClickhouse(GetConfigFromIni(cfg))

	// record := &data.TradeRecord{OrderID: "123456", PayWay: 1}
	// rs, err := ParseData([]*data.TradeRecord{record})

	// if err != nil {
	// 	t.Error(err)
	// }
	// for _, r := range rs {
	// 	fmt.Printf("%#v\n", r)
	// }
}

func TestCkParser2(t *testing.T) {
	//加载配置
	cfg, err := ini.Load("../../bin/config/conf.ini")
	if err != nil {
		t.Error(err)
	}
	InitClickhouse(GetConfigFromIni(cfg))

	record := &data.TradeRecord{OrderID: "123456", PayWay: 1}
	record2 := &data.TradeRecord{OrderID: "456789", Ctime: time.Now()}
	// r, err := ParseData(record)
	rs, err := ParseData([]*data.TradeRecord{record, record2})
	if err != nil {
		t.Error(err)
	}
	for _, r := range rs {
		fmt.Printf("%#v\n", r)
	}
}

func TestFind(t *testing.T) {
	//加载配置
	cfg, err := ini.Load("../../bin/config/conf.ini")
	if err != nil {
		t.Error(err)
	}
	InitClickhouse(GetConfigFromIni(cfg))

	// Select
	// r := &entity.TradeRecord{}
	var r []*TradeRecord
	// err = db.Find(&r, "order_id = ?", "1").Error
	err = db.Raw(`
		select * from t_trade_record final where order_id = ?
	`, "1").Scan(&r).Error
	if err != nil {
		t.Error(err)
	}
	for _, rr := range r {
		fmt.Printf("%#v\n", rr)
	}
}

func TestInsert(t *testing.T) {
	//加载配置
	cfg, err := ini.Load("../../bin/config/conf.ini")
	if err != nil {
		t.Error(err)
	}
	InitClickhouse(GetConfigFromIni(cfg))

	for i := 0; i < 100; i++ {
		var records []*TradeRecord
		for j := 0; j < 100; j++ {
			records = append(records, &TradeRecord{
				OrderID: strconv.Itoa(i*100 + j),
				Userid:  "10086",
				Ctime:   time.Now(),
				Ver:     time.Now().Unix(),
			})
		}
		err := InsertBatch(records)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestKafka(t *testing.T) {
	// to produce messages
	broker := "192.168.0.137:9092"
	topic := "topic1"

	go producer(topic, broker)

	consumer(topic, broker)
}

func producer(topic string, brokers ...string) {
	// w := kafka.NewWriter(kafka.WriterConfig{
	// 	Brokers: brokers,
	// 	Topic:   topic,
	// 	// Partition: 0,
	// 	// MinBytes:  10e3, // 10KB
	// 	// MaxBytes:  10e6, // 10MB
	// })

	partition := 0
	conn, err := kafka.DialLeader(context.Background(), "tcp", brokers[0], topic, partition)
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	for i := 0; i < 100; i++ {
		_, err := conn.WriteMessages(
			kafka.Message{Value: []byte("one!" + strconv.Itoa(i))},
			kafka.Message{Value: []byte("two!" + strconv.Itoa(i))},
			kafka.Message{Value: []byte("three!" + strconv.Itoa(i))},
		)
		if err != nil {
			log.Fatal("failed to write messages:", err)
		}
	}

	// if err := conn.Close(); err != nil {
	// 	log.Fatal("failed to close writer:", err)
	// }
}

func consumer(topic string, brokers ...string) {
	// conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	// batch := conn.ReadBatch(10e3, 1e6) // fetch 10KB min, 1MB max
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   brokers,
		Topic:     topic,
		GroupID:   "consumer-group-1",
		Partition: 0,
		MinBytes:  10e3, // 10KB
		MaxBytes:  10e6, // 10MB
	})
	// r.SetOffset(42)
	for {
		m, err := r.FetchMessage(context.Background())
		// m, err := r.ReadMessage(context.Background()) // read 自动提交
		if err != nil {
			log.Fatal(err)
			break
		}

		// fetch 后不 comment 会重复推送
		if err := r.CommitMessages(context.Background(), m); err != nil {
			log.Fatal("failed to commit messages:", err)
		}
		fmt.Printf("message at offset %d: %s = %s\n", m.Offset, string(m.Key), string(m.Value))
	}

	if err := r.Close(); err != nil {
		log.Fatal("failed to close reader:", err)
	}

}

func TestDays(t *testing.T) {
	begin, err := time.Parse(utils.FORMAT, "2023-11-01 00:00:00")
	if err != nil {
		t.Error(err)
	}
	var _, days, concurrent int = 1, 7, 8

	// thread := 0
	for thread := 0; thread < concurrent; thread++ {
		for i := 0; i < 10; i++ {
			offset := thread*days + concurrent*days*i
			stime := begin.AddDate(0, 0, offset)
			etime := begin.AddDate(0, 0, offset+days)
			fmt.Printf("%v-%v\n", stime, etime)
		}
		fmt.Println("=================================")
	}

}

func TestQueryBets(t *testing.T) {
	InitClickhouse(ClickhouseConfig{
		Logsql:      true,
		Addr:        "127.0.0.1:9000",
		Database:    "game",
		DialTimeout: 10,
		ReadTimeout: 10,
	})

	var bets int64
	err := Select(&bets, `
		SELECT SUM(bet_amount) bet_amount FROM (
			SELECT SUM(bet_amount) bet_amount
			FROM game.col_detail t1 FINAL WHERE userid = ? AND robot = 0 and win_type in (1,2,3) 
				UNION ALL
			SELECT SUM(amount) bet_amount
			FROM game.col_nsq_log_external_bet t2 FINAL WHERE user_id = ? AND amount != 0 
		) s1
	`, "120914", "120914")
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("bets", bets)
}
