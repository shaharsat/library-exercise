package main

import (
	"gin/config"
	"gin/db"
	"gin/service"
)

func main() {
	config.SetupRedis()
	db.SetupElasticLibrary()
	routes := service.SetupRoutes()
	routes.Run()
}
