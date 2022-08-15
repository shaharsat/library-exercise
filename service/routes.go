package service

import (
	"github.com/gin-gonic/gin"
)

func SetupRoutes() *gin.Engine {
	routes := gin.Default()

	routes.GET("/activity/:username", Activity)

	cachedRoutes := routes.Group("/")
	{
		cachedRoutes.Use(CacheUserRequest)

		cachedRoutes.PUT("/book", CreateBook)
		cachedRoutes.POST("/book/:id", UpdateBookTitleById)
		cachedRoutes.GET("/book/:id", GetBookById)
		cachedRoutes.DELETE("/book/:id", DeleteBookById)
		cachedRoutes.GET("/search", SearchBooks)
		cachedRoutes.GET("/store", Store)
	}

	return routes
}
