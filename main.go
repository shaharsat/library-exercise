package main

import (
	"gin/service"
)

func main() {
	routes := service.SetupRoutes()
	routes.Run()
}
