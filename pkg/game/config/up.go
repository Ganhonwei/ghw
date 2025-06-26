package config

import "goserver/pkg/data"

var seven data.SevenStrategy

func InitSeven() {
	seven = data.GetSevenStrategy()
}

func GetSevenStrategy() data.SevenStrategy {
	return seven
}

func SetSevenStrategy(s data.SevenStrategy) {
	seven = s
}
