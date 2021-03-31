package main

import (
	"flag"

	"jiange/config"
	"jiange/handler"
	"./log"
)

func main() {
	configPath := flag.String("config", "", "配置信息路径")
	port := flag.Int("port", -1, "服务监听端口")
	flag.Parse()
	config.InitConfig(*configPath, *port)

	// set log format
	defaultFields := log.Fields{}
	if err := log.InitPath(config.Config.LogPath, "text", "info", defaultFields); err != nil {
		panic(err.Error())
	}
	//init acc log
	log.Init()
	handler.Start()
}
