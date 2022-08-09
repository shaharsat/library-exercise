package main

import (
	"gin/cmd"
	"gin/models"
)

func main() {
	models.SetupRedis()
	models.SetupElasticSearch()
	routes := cmd.SetupRoutes()
	routes.Run()
}
