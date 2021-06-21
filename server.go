package main

import (
	// "product_code/check_stream/alarm"
	"product_code/check_stream/stream"
	"product_code/check_stream/web"
	"time"

	log4plus "common/log4go"
)

type CheckServer struct {
}

func NewCheckServer() *CheckServer {
	log4plus.Info("NewCheckServer")
	server := &CheckServer{}

	//initialize stream
	stream.GetManager().Initailzie()

	return server
}

func (server *CheckServer) Run() {
	//StartWeb
	web.StartWeb()

	//start manager
	stream.GetManager().Start()

	//main loop
	for {
		time.Sleep(5 * time.Second)
		stream.GetManager().CheckTask()

		// log4plus.Debug("NotifyManager message left len=%d", alarm.MessageLen())
	}
}
