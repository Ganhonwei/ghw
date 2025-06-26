package service

import (
	"goserver/pkg/data"
	"testing"

	"gopkg.in/ini.v1"
)

func TestPublishRankWithdraw(t *testing.T) {
	cfg, err := ini.Load("../../bin/config/conf.ini")
	if err != nil {
		panic(err)
	}
	InitMgo(cfg)
	InitNsq(cfg)
	order := &data.PayNotify{MerOrderID: "234167843720396802"}
	publishRankWithdraw(order)
}
