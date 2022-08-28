package service

import (
	"encoding/json"
	"gin/cache"
	"gin/db"
	"gin/models"
	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
	"net/http"
)

const INDEX_NAME = "books"
const MAX_NUMBER_CACHED = 3
const STATUS_KEY = "status"
const STATUS_DELETED = "deleted"
const STATUS_CREATED = "created"
const STATUS_UPDATED = "updated"

var ElasticLibrary = db.CreateElasticLibrary(INDEX_NAME)
var RedisCache = cache.CreateRedisCache(MAX_NUMBER_CACHED)

func CreateBook(c *gin.Context) {
	var book models.Book
	if err := c.ShouldBindJSON(&book); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	}

	id, err := ElasticLibrary.Create(&book)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		STATUS_KEY: STATUS_CREATED,
		"id":       id,
	})
}

func UpdateBookTitleById(c *gin.Context) {
	id := c.Param("id")

	var book *models.Book
	if err := c.ShouldBindJSON(&book); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	err := ElasticLibrary.Update(id, book)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{STATUS_KEY: STATUS_UPDATED})
}

func GetBookById(c *gin.Context) {
	id := c.Param("id")

	book, err := ElasticLibrary.GetById(id)

	switch t := err.(type) {
	case *elastic.Error:
		c.AbortWithStatusJSON(t.Status, gin.H{"message": t.Error()})
		return
	case error:
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, book)
}

func DeleteBookById(c *gin.Context) {
	id := c.Param("id")

	err := ElasticLibrary.Delete(id)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{STATUS_KEY: STATUS_DELETED})
}

func SearchBooks(c *gin.Context) {
	title := c.Query("title")
	authorName := c.Query("author_name")
	minPrice := c.Query("min_price")
	maxPrice := c.Query("max_price")

	if title == "" && authorName == "" && minPrice == "" && maxPrice == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "no field to search by found"})
		return
	}

	books, err := ElasticLibrary.Search(title, authorName, minPrice, maxPrice)

	switch err.(type) {
	case *elastic.Error:
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	case error:
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, books)
}

func Store(c *gin.Context) {
	numberOfBooks, numberOfAuthors, err := ElasticLibrary.Store()

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"number_of_books":   numberOfBooks,
		"number_of_authors": numberOfAuthors,
	},
	)
}

func Activity(c *gin.Context) {
	username := c.Param("username")

	userRequests, err := RedisCache.Read(username)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	userRequestsRaw := make([]models.UserRequest, 0)

	var userRequest models.UserRequest
	for _, request := range userRequests {
		err := json.Unmarshal([]byte(request), &userRequest)
		if err != nil {
			return
		}
		userRequestsRaw = append(userRequestsRaw, userRequest)
	}

	c.JSON(http.StatusOK, userRequestsRaw)
}

func CacheUserRequest(c *gin.Context) {
	userRequest := models.UserRequest{
		Method: c.Request.Method,
		Route:  c.Request.URL.Path,
	}

	username, ok := c.GetQuery("username")

	if !ok {
		return
	}

	request, err := json.Marshal(userRequest)
	if err == nil {
		RedisCache.Write(username, request)
	}

	c.Next()
}
