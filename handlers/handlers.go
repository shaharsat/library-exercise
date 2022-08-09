package handlers

import (
	"encoding/json"
	"fmt"
	"gin/models"
	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
	"html"
	"net/http"
	"strconv"
)

const INDEX_NAME = "books"

func CreateBook(c *gin.Context) {
	var book models.Book
	if err := c.ShouldBindJSON(&book); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	}

	doc, err := models.ElasticClient.Index().Index(INDEX_NAME).BodyJson(book).Do(c)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "created",
		"id":     doc.Id,
	})
}

func UpdateBookTitleById(c *gin.Context) {
	id := c.Param("id")

	var book *models.Book
	if err := c.ShouldBindJSON(&book); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	_, err := models.ElasticClient.
		Update().
		Index(INDEX_NAME).
		Id(id).
		Doc(gin.H{"title": book.Title}).
		Do(c)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func GetBookById(c *gin.Context) {
	id := c.Param("id")

	doc, err := models.ElasticClient.
		Get().
		Index(INDEX_NAME).
		Id(id).
		Do(c)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if !doc.Found {
		c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("book with Id: '%v' not found", id)})
		return
	}

	var book models.Book
	err = json.Unmarshal(doc.Source, &book)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	}

	c.JSON(http.StatusOK, book)
}

func DeleteById(c *gin.Context) {
	id := c.Param("id")

	doc, err := models.ElasticClient.
		Delete().
		Index(INDEX_NAME).
		Id(id).
		Do(c)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted", "id": doc.Id})
}

func Search(c *gin.Context) {
	title := c.Query("title")
	authorName := c.Query("authorName")
	minPrice, minPriceOk := c.GetQuery("min_price")
	maxPrice, maxPriceOk := c.GetQuery("max_price")

	index := models.ElasticClient.Search().Index(INDEX_NAME).Pretty(false).Size(10000)

	boolQuery := elastic.NewBoolQuery()
	if title != "" {
		boolQuery.Must(elastic.NewMatchQuery("title", html.UnescapeString(title)))
	}
	if authorName != "" {
		boolQuery.Must(elastic.NewMatchQuery("author_name", html.UnescapeString(authorName)))
	}

	priceRangeQuery := elastic.NewRangeQuery("price")

	shouldIncludePriceRangeQuery := false
	if minPriceOk {
		price, err := strconv.ParseFloat(minPrice, 64)
		if err == nil {
			index.Query(priceRangeQuery.Gte(price))
			shouldIncludePriceRangeQuery = true
		} else {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
	}

	if maxPriceOk {
		price, err := strconv.ParseFloat(maxPrice, 64)
		if err == nil {
			index.Query(priceRangeQuery.Lte(price))
			shouldIncludePriceRangeQuery = true
		} else {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
	}

	if shouldIncludePriceRangeQuery {
		boolQuery.Must(priceRangeQuery)
	}

	index.Query(boolQuery)

	result, err := index.Do(c)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	books := make([]models.Book, len(result.Hits.Hits))
	for i, hit := range result.Hits.Hits {
		err = json.Unmarshal(hit.Source, &books[i])
		if err != nil {
			fmt.Println(err)
		}
	}

	c.JSON(http.StatusOK, books)
}

func Store(c *gin.Context) {
	query := models.ElasticClient.Search().
		Index(INDEX_NAME)

	titleAggregation := elastic.NewCardinalityAggregation().Field("_id")
	authorsAggregation := elastic.NewCardinalityAggregation().Field("author_name.keyword")
	query.Aggregation("number_of_books", titleAggregation)
	query.Aggregation("number_of_authors", authorsAggregation)
	query.Size(0)

	results, err := query.Do(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	numberOfBooks, _ := results.Aggregations.Cardinality("number_of_books")
	numberOfAuthors, _ := results.Aggregations.Cardinality("number_of_authors")

	c.JSON(
		http.StatusOK,
		gin.H{
			"number_of_books":   numberOfBooks.Value,
			"number_of_authors": numberOfAuthors.Value,
		},
	)
}

func Activity(c *gin.Context) {
	username := c.Param("username")

	userRequests, err := models.RedisClient.LRange(username, 0, 2).Result()

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	ops := make([]models.UserRequest, 0)

	var userRequest models.UserRequest
	for _, request := range userRequests {
		err := json.Unmarshal([]byte(request), &userRequests)
		if err != nil {
			return
		}
		ops = append(ops, userRequest)
	}

	c.JSON(http.StatusOK, ops)
}

func CacheUserRequest(c *gin.Context) {
	operation := models.UserRequest{
		Method: c.Request.Method,
		Route:  c.Request.URL.Path,
	}

	username, ok := c.GetQuery("username")

	if !ok {
		return
	}

	obj, _ := json.Marshal(operation)
	// Not failing a request if there's a problem caching it
	models.RedisClient.LPush(username, obj)
	models.RedisClient.LTrim(username, 0, 2)
	c.Next()
}
