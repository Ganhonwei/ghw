package config

import "goserver/pkg/data"

var Lhd data.LhdStrategy

func InitLHD() {
	Lhd = data.GetLHDStrategy()
}

func GetLHDStrategy() data.LhdStrategy {
	return Lhd
}

func SetLHDStrategy(s data.LhdStrategy) {
	Lhd = s
}
