package config

import "goserver/pkg/data"

var strategy data.CrashStrategy

func InitCrash() {
	strategy = data.GetCrashStrategy()
}

func GetCrashStrategy() data.CrashStrategy {
	return strategy
}

func SetCrashStrategy(s data.CrashStrategy) {
	strategy = s
}
