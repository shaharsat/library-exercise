package main

import (
	"fmt"
	"gin/handlers"
	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
	"gopkg.in/redis.v5"
	"os"
)

func setupServer() (*handlers.Server, error) {
	elasticUrl := os.Getenv("ELASTIC_URL")
	redisUrl := os.Getenv("REDIS_URL")

	elasticClient, err := elastic.NewClient(elastic.SetURL(elasticUrl))

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisUrl,
	})

	server := handlers.Server{
		ElasticClient: elasticClient,
		RedisClient:   redisClient,
	}

	return &server, nil
}

func setupRoutes(server *handlers.Server) *gin.Engine {
	routes := gin.Default()

	cachedRoutes := routes.Group("/")
	cachedRoutes.Use(server.CacheUserRequest)

	cachedRoutes.PUT("/book", server.CreateBook)
	cachedRoutes.POST("/book", server.UpdateBookTitleById)
	cachedRoutes.GET("/book", server.GetBookById)
	cachedRoutes.DELETE("/book", server.DeleteById)
	cachedRoutes.GET("/search", server.Search)
	cachedRoutes.GET("/stores", server.Store)
	routes.GET("/activity", server.Activity)
	return routes
}

func main() {
	server, err := setupServer()

	if err != nil {
		fmt.Println(fmt.Errorf("failed to setup server due to the following error: %v", err.Error()))
		os.Exit(-1)
	}

	routes := setupRoutes(server)
	routes.Run()
}
