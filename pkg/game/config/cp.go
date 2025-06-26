package config

import "goserver/pkg/data"

var cp data.CPStrategy

func InitCP() {
	cp = data.GetCPStrategy()
}

func GetCPStrategy() data.CPStrategy {
	return cp
}

func SetCPStrategy(s data.CPStrategy) {
	cp = s
}
