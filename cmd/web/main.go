package main

import (
	"fmt"
	"os"

	"goserver/internal/web/admin"
	"goserver/internal/web/agent"
	"goserver/internal/web/stats"
	"goserver/pkg/flags"
)

type ServiceRunner interface {
	Run()
}

var services = map[string]ServiceRunner{
	"admin": &admin.Service{},
	"agent": &agent.Service{},
	"stats": &stats.Service{},
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
