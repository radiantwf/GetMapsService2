package services

import (
	"fmt"
	"os"

	"github.com/facebookgo/inject"
)

// WorkService 定义
type WorkService struct {
	Config    *ConfigService    `inject:""`
	GetMaps   *GetMapsService   `inject:""`
	WebSocket *WebSocketService `inject:""`
}

// Run 定义
func (w *WorkService) Run() {
	if err := inject.Populate(w.Config); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	w.WebSocket.submitCallback = w.GetMaps.GetMaps
	w.WebSocket.Start()
}
