package config

import "goserver/pkg/data"

var av data.AviatorStrategy

func InitAviator() {
	av = data.GetAviatorStrategy()
}

func GetAviatorStrategy() data.AviatorStrategy {
	return av
}

func SetAviatorStrategy(s data.AviatorStrategy) {
	av = s
}
