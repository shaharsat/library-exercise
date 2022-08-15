package main

import (
	"gin/config"
	"gin/service"
)

func main() {
	config.SetupRedis()
	config.SetupElasticSearch()
	routes := service.SetupRoutes()
	routes.Run()
}
