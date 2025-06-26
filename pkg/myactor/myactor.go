package myactor

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"gopkg.in/ini.v1"
)

var (
	logger *actor.PID
)

func InitLogger(cfg *ini.File) {
	bind := cfg.Section("logger").Key("bind").String()
	kind := cfg.Section("logger").Key("kind").String()
	logger = actor.NewPID(bind, kind)
}

func Logger() *actor.PID {
	if logger == nil {
		panic("logger not init")
	}
	return logger
}
