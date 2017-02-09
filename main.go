package main

import (
	"GetMapsService2/services"
	"fmt"
	"os"
	"runtime"

	"github.com/facebookgo/inject"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	var work services.WorkService
	config, err := services.NewConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := inject.Populate(&work, &config); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	work.Run()
	// webSocketService := NewWebSocketService("resources/web","home.html",config.Port)
	// getBaiduMap := NewGetBaiduMap(config,webSocketService.BroadcastMessage)
	// webSocketService.submitCallback = getBaiduMap.Run

	// webSocketService.Start()
}
