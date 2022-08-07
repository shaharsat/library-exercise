package handlers

import (
	"encoding/json"
	"fmt"
	"gin/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/olivere/elastic/v7"
	"gopkg.in/redis.v5"
	"html"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	ElasticClient *elastic.Client
	RedisClient   *redis.Client
}

func (server *Server) validateSearchBook(title string, titleOk bool, authorName string, authorNameOk bool, price string, priceOk bool, ebookAvailable string, ebookAvailableOk bool, publishDate string, publishDateOk bool) error {
	if (!titleOk || title == "") && (!authorNameOk || authorName == "") && (!priceOk || price == "") && (!ebookAvailableOk || ebookAvailable == "") && (!publishDateOk || publishDate == "") {
		return fmt.Errorf("no parameter given. requires at least one of the following: 'title', 'authorName', 'price', 'ebookAvailable', 'publishDate")
	}
	return nil
}

func (server *Server) validateCreateBook(title, authorName, price, ebookAvailable, publishDate string) error {
	missingParameter := make([]string, 0)

	if title == "" {
		missingParameter = append(missingParameter, "'title'")
	}
	if authorName == "" {
		missingParameter = append(missingParameter, "'authorName'")
	}
	if price == "" {
		missingParameter = append(missingParameter, "'price'")
	}
	if ebookAvailable == "" {
		missingParameter = append(missingParameter, "'ebookAvailable'")
	}
	if publishDate == "" {
		missingParameter = append(missingParameter, "'publishDate'")
	}

	if len(missingParameter) != 0 {
		return fmt.Errorf("the following parameters are missing: %v", strings.Join(missingParameter, ","))
	}

	return nil
}

func (server *Server) CreateBook(c *gin.Context) {
	title := c.PostForm("title")
	authorName := c.PostForm("authorName")
	price := c.PostForm("price")
	ebookAvailable := c.PostForm("ebookAvailable")
	publishDate := c.PostForm("publishDate")

	validationError := server.validateCreateBook(title, authorName, price, ebookAvailable, publishDate)
	if validationError != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": validationError.Error()})
		return
	}

	priceF, err := strconv.ParseFloat(price, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	ebookAvailableB, err := strconv.ParseBool(ebookAvailable)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return

	}

	publishDateT, err := time.Parse("2006-01-02", publishDate)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	id := uuid.New().String()

	book := models.Book{
		Id:             id,
		Title:          title,
		AuthorName:     authorName,
		Price:          priceF,
		EbookAvailable: ebookAvailableB,
		PublishDate:    models.CustomTime(publishDateT),
	}

	_, err = server.ElasticClient.Index().Index("books").BodyJson(book).Do(c)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "created",
		"id":     id,
	})
}

func (server *Server) UpdateBookTitleById(c *gin.Context) {
	id, idOk := c.GetPostForm("id")
	title, titleOk := c.GetPostForm("title")

	missingFields := make([]string, 0)
	if !idOk || id == "" {
		missingFields = append(missingFields, "'id'")
	}
	if !titleOk || title == "" {
		missingFields = append(missingFields, "'title'")
	}
	if len(missingFields) != 0 {
		validationError := fmt.Errorf("the following fields are missing: %v", missingFields)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": validationError.Error()})
		return
	}

	scriptFormat := "ctx._source.%s = \"%v\"" // Is this vulnerable to ElasticSearch Injection?
	scriptString := fmt.Sprintf(scriptFormat, "title", title)
	script := elastic.NewScript(scriptString).Lang("painless")

	_, err := server.ElasticClient.
		UpdateByQuery().
		Index("books").
		Query(elastic.NewTermQuery("index._id", id)).
		MaxDocs(1).
		Script(script).
		Do(c)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (server *Server) GetBookById(c *gin.Context) {
	id, idOk := c.GetQuery("id")

	missingFields := make([]string, 0)
	if !idOk || id == "" {
		missingFields = append(missingFields, "'id'")
	}
	if len(missingFields) != 0 {
		validationError := fmt.Errorf("the following fields are missing: %v", missingFields)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": validationError.Error()})
		return
	}

	doc, err := server.ElasticClient.
		Search().
		Index("books").
		Query(elastic.NewTermQuery("index._id", id)).
		Do(c)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if doc.Hits.TotalHits.Value == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("book with Id: '%v' not found", id)})
		return
	}

	var book models.Book
	err = json.Unmarshal(doc.Hits.Hits[0].Source, &book)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	}

	c.JSON(http.StatusOK, book)
}

func (server *Server) DeleteById(c *gin.Context) {
	id, idOk := c.GetQuery("id")

	missingFields := make([]string, 0)
	if !idOk || id == "" {
		missingFields = append(missingFields, "'id'")
	}

	if len(missingFields) != 0 {
		validationError := fmt.Errorf("the following fields are missing: %v", missingFields)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": validationError.Error()})
		return
	}

	doc, err := server.ElasticClient.
		Delete().
		Index("books").
		Id(id).
		Do(c)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted", "id": doc.Id})
}

func (server *Server) Search(c *gin.Context) {
	title, titleOk := c.GetQuery("title")
	authorName, authorNameOk := c.GetQuery("authorName")
	price, priceOk := c.GetQuery("price")
	ebookAvailable, ebookAvailableOk := c.GetQuery("ebookAvailable")
	publishDate, publishDateOk := c.GetQuery("publishDate")

	validationError := server.validateSearchBook(title, titleOk, authorName, authorNameOk, price, priceOk, ebookAvailable, ebookAvailableOk, publishDate, publishDateOk)
	if validationError != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": validationError.Error()})
		return
	}

	index := server.ElasticClient.Search().Index("books").Pretty(false).Size(10000)

	if titleOk && title != "" {
		index.Query(elastic.NewMatchQuery("title", "*"+html.UnescapeString(title)+"*"))
	}
	if authorNameOk && authorName != "" {
		index.Query(elastic.NewMatchQuery("author_name", html.UnescapeString(authorName)))
	}

	if priceOk {
		priceF, err := strconv.ParseFloat(price, 64)
		if err == nil {
			index.Query(elastic.NewMatchQuery("price", priceF))
		} else {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

	}

	if ebookAvailableOk {
		ebookAvailableB, err := strconv.ParseBool(ebookAvailable)
		if err == nil {
			index.Query(elastic.NewMatchQuery("ebook_available", ebookAvailableB))
		} else {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
	}

	if publishDateOk {
		_, err := time.Parse("2006-01-01", publishDate) // valid time validation
		if err == nil {
			index.Query(elastic.NewMatchQuery("publish_date", publishDate))
		} else {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
	}

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

func (server *Server) Store(c *gin.Context) {
	query := server.ElasticClient.Search().
		Index("books").
		Query(elastic.NewMatchAllQuery())

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

	var searchResult models.SearchResult
	res, err := json.Marshal(results.Aggregations)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	err = json.Unmarshal(res, &searchResult)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"number_of_books":   searchResult.NumberOfBooks.Value,
		"number_of_authors": searchResult.NumberOfBooks.Value,
	})
}

func (server *Server) Activity(c *gin.Context) {
	username := c.Query("username")

	get, err := server.RedisClient.LRange(username, 0, 2).Result()

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	var ops [3]models.UserRequest

	for idx, val := range get {
		err := json.Unmarshal([]byte(val), &ops[idx])
		if err != nil {
			return
		}
	}

	c.JSON(http.StatusOK, ops[:len(get)])
}

func (server *Server) CacheUserRequest(c *gin.Context) {
	operation := models.UserRequest{
		Method: c.Request.Method,
		Route:  c.Request.URL.Path,
	}

	username, ok := c.GetQuery("username")

	if !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "missing 'username' query parameter"})
		return
	}

	obj, _ := json.Marshal(operation)
	// Not failing a request if there's a problem caching it
	server.RedisClient.LPush(username, obj)
	server.RedisClient.LTrim(username, 0, 2)
	c.Next()
}
