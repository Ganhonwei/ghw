package gate

import (
	"context"

	"github.com/nsqio/go-nsq"
	"github.com/redis/go-redis/v9"
)

var (
	producer *nsq.Producer
	client   *redis.Client
)

func InitNsq() {
	InitRedis()
	InitNsqProducer()
}

func InitNsqProducer() {
	var err error
	nsqAddr := cfg.Section("nsq").Key("addr").Value()
	producer, err = nsq.NewProducer(nsqAddr, nsq.NewConfig())
	if err != nil {
		panic(err)
	}
}

func InitRedis() {
	redisAddr := cfg.Section("redis").Key("addr").Value()
	client = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}
}
