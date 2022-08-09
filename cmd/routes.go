package cmd

import (
	"gin/handlers"
	"github.com/gin-gonic/gin"
)

func SetupRoutes() *gin.Engine {
	routes := gin.Default()

	routes.GET("/activity", handlers.Activity)

	cachedRoutes := routes.Group("/")
	{
		cachedRoutes.Use(handlers.CacheUserRequest)
		cachedRoutes.PUT("/book", handlers.CreateBook)
		cachedRoutes.POST("/book", handlers.UpdateBookTitleById)
		cachedRoutes.GET("/book", handlers.GetBookById)
		cachedRoutes.DELETE("/book", handlers.DeleteById)
		cachedRoutes.GET("/search", handlers.Search)
		cachedRoutes.GET("/store", handlers.Store)
	}

	return routes
}
