package main

import (
	"gin/db"
	"gin/service"
)

func main() {
	db.SetupElasticLibrary()
	routes := service.SetupRoutes()
	routes.Run()
}
