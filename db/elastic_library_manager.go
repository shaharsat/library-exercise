package db

import (
	"context"
	"encoding/json"
	"gin/config"
	"gin/models"
	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
	"html"
	"strconv"
)

type ElasticLibraryManager struct {
	IndexName string
}

func CreateElasticLibrary(indexName string) *ElasticLibraryManager {
	return &ElasticLibraryManager{indexName}
}

func (library *ElasticLibraryManager) Create(book *models.Book) (string, error) {
	doc, err := config.ElasticClient.Index().Index(library.IndexName).BodyJson(book).Do(context.Background())
	return doc.Id, err
}

func (library *ElasticLibraryManager) Update(id string, book *models.Book) error {
	_, err := config.ElasticClient.
		Update().
		Index(library.IndexName).
		Id(id).
		Doc(gin.H{"title": book.Title}).
		Do(context.Background())

	return err
}

func (library *ElasticLibraryManager) GetById(id string) (*models.Book, error) {
	doc, err := config.ElasticClient.
		Get().
		Index(library.IndexName).
		Id(id).
		Do(context.Background())

	if err != nil {
		return nil, err
	}

	var book models.Book
	err = json.Unmarshal(doc.Source, &book)

	if err != nil {
		return nil, err
	}

	return &book, nil
}

func (library *ElasticLibraryManager) Delete(id string) error {
	_, err := config.ElasticClient.
		Delete().
		Index(library.IndexName).
		Id(id).
		Do(context.Background())

	return err
}

func (library *ElasticLibraryManager) Search(title, authorName, minPrice, maxPrice string) ([]*models.Book, error) {
	boolQuery := elastic.NewBoolQuery()
	if title != "" {
		boolQuery.Must(elastic.NewTermQuery("title.keyword", title))
	}
	if authorName != "" {
		boolQuery.Must(elastic.NewMatchQuery("author_name", html.UnescapeString(authorName)))
	}

	priceRangeQuery := elastic.NewRangeQuery("price")

	shouldIncludePriceRangeQuery := false
	if minPrice != "" {
		price, err := strconv.ParseFloat(minPrice, 64)
		if err != nil {
			return nil, err
		}
		priceRangeQuery = priceRangeQuery.Gte(price)
		shouldIncludePriceRangeQuery = true
	}

	if maxPrice != "" {
		price, err := strconv.ParseFloat(maxPrice, 64)
		if err != nil {
			return nil, err
		}
		priceRangeQuery = priceRangeQuery.Lte(price)
		shouldIncludePriceRangeQuery = true
	}

	if shouldIncludePriceRangeQuery {
		boolQuery.Must(priceRangeQuery)
	}

	index := config.ElasticClient.Search().
		Index(library.IndexName).
		Pretty(false).
		Size(10000).
		Query(boolQuery)

	result, err := index.Do(context.Background())

	if err != nil {
		return nil, err
	}

	books := make([]*models.Book, result.Hits.TotalHits.Value)
	for i, hit := range result.Hits.Hits {
		json.Unmarshal(hit.Source, &books[i])
	}

	return books, nil
}

func (library *ElasticLibraryManager) Store() (*float64, *float64, error) {
	titleAggregation := elastic.NewCardinalityAggregation().Field("_id")
	authorsAggregation := elastic.NewCardinalityAggregation().Field("author_name.keyword")

	query := config.ElasticClient.Search().
		Index(library.IndexName).
		Aggregation("number_of_books", titleAggregation).
		Aggregation("number_of_authors", authorsAggregation).
		Size(0)

	results, err := query.Do(context.Background())
	if err != nil {
		return nil, nil, err
	}

	numberOfBooks, _ := results.Aggregations.Cardinality("number_of_books")
	numberOfAuthors, _ := results.Aggregations.Cardinality("number_of_authors")

	return numberOfBooks.Value, numberOfAuthors.Value, nil
}