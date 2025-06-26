package config

import "goserver/pkg/data"

var rb data.RBStrategy

func InitRB() {
	rb = data.GetRBStrategy()
}

func GetRBStrategy() data.RBStrategy {
	return rb
}

func SetRBStrategy(s data.RBStrategy) {
	rb = s
}
