package main

import (
	"fmt"
	"os"

	"goserver/internal/game/ak47"
	"goserver/internal/game/andarbahar"
	"goserver/internal/game/crash"
	"goserver/internal/game/dbms"
	"goserver/internal/game/fortune_gems"
	"goserver/internal/game/fortune_gems2"
	"goserver/internal/game/gate"
	"goserver/internal/game/hua"
	"goserver/internal/game/joker"
	"goserver/internal/game/lhd"
	"goserver/internal/game/login"
	"goserver/internal/game/lottery"
	"goserver/internal/game/mines"
	"goserver/pkg/flags"

	"goserver/internal/game/logger"
	"goserver/internal/game/plane"
	"goserver/internal/game/redblack"
	"goserver/internal/game/rm2"
	"goserver/internal/game/robot"
	"goserver/internal/game/up7"
)

type ServiceRunner interface {
	Run()
}

var services = map[string]ServiceRunner{
	"dbms":          &dbms.Service{},
	"gate":          &gate.Service{},
	"login":         &login.Service{},
	"hua":           &hua.Service{},
	"joker":         &joker.Service{},
	"ak47":          &ak47.Service{},
	"rm2":           &rm2.Service{},
	"lhd":           &lhd.Service{},
	"up7":           &up7.Service{},
	"crash":         &crash.Service{},
	"plane":         &plane.Service{},
	"andarbahar":    &andarbahar.Service{},
	"redblack":      &redblack.Service{},
	"lottery":       &lottery.Service{},
	"mines":         &mines.Service{},
	"robot":         &robot.Service{},
	"logger":        &logger.Service{},
	"fortune_gems2": &fortune_gems2.Service{},
	"fortune_gems":  &fortune_gems.Service{},
}

func main() {
	target := flags.Target
	if target == "" {
		fmt.Println("Available services:")
		for name := range services {
			fmt.Printf("  - %s\n", name)
		}
		os.Exit(1)
	}

	service, ok := services[target]
	if !ok {
		fmt.Printf("Unknown service: %s\n", target)
		os.Exit(1)
	}

	service.Run()
}
