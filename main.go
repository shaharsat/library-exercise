package main

import (
	"gin/config"
	"gin/internal"
)

func main() {
	config.SetupRedis()
	config.SetupElasticSearch()
	routes := internal.SetupRoutes()
	routes.Run()
}
