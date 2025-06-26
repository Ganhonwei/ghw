package main

import (
	"fmt"
	"os"

	"goserver/internal/global/alert"
	"goserver/internal/global/betrank"
	"goserver/internal/global/online"
	"goserver/internal/global/pay"
	"goserver/internal/global/report"
	"goserver/internal/global/sms"
	"goserver/internal/global/sync"
	"goserver/internal/global/utr"
	"goserver/internal/global/withdraw"
	"goserver/pkg/flags"
)

type ServiceRunner interface {
	Run()
}

var services = map[string]ServiceRunner{
	"sms":      &sms.Service{},
	"sync":     &sync.Service{},
	"betrank":  &betrank.Service{},
	"alert":    &alert.Service{},
	"report":   &report.Service{},
	"pay":      &pay.Service{},
	"utr":      &utr.Service{},
	"online":   &online.Service{},
	"withdraw": &withdraw.Service{},
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
