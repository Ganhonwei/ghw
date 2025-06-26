package pay

import (
	"goserver/internal/global/pay/service"

	"github.com/nsqio/go-nsq"
)

func InitNsqProducer() {
	var err error
	nsqAddr := cfg.Section("nsq").Key("addr").Value()
	service.Producer, err = nsq.NewProducer(nsqAddr, nsq.NewConfig())
	if err != nil {
		panic(err)
	}
}
