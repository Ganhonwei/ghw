package config

import "goserver/pkg/data"

var ab data.ABStrategy

func InitAB() {
	ab = data.GetABStrategy()
}

func GetABStrategy() data.ABStrategy {
	return ab
}

func SetABStrategy(s data.ABStrategy) {
	ab = s
}
