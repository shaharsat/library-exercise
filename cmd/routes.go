package cmd

import (
	"gin/handlers"
	"github.com/gin-gonic/gin"
)

func SetupRoutes() *gin.Engine {
	routes := gin.Default()

	routes.GET("/activity:username", handlers.Activity)

	cachedRoutes := routes.Group("/")
	{
		cachedRoutes.Use(handlers.CacheUserRequest)
		cachedRoutes.PUT("/book", handlers.CreateBook)
		cachedRoutes.POST("/book/:id", handlers.UpdateBookTitleById)
		cachedRoutes.GET("/book/:id", handlers.GetBookById)
		cachedRoutes.DELETE("/book/:id", handlers.DeleteById)
		cachedRoutes.GET("/search", handlers.Search)
		cachedRoutes.GET("/store", handlers.Store)
	}

	return routes
}
