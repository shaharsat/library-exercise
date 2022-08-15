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

type ElasticLibraryDatabase struct {
	IndexName string
}

func CreateElasticLibrary(indexName string) *ElasticLibraryDatabase {
	return &ElasticLibraryDatabase{indexName}
}

func (library *ElasticLibraryDatabase) Delete(id Id) error {
	_, err := config.ElasticClient.
		Delete().
		Index(library.IndexName).
		Id(string(id)).
		Do(context.Background())

	return err
}

func (library *ElasticLibraryDatabase) GetById(id Id) (*models.Book, error) {
	doc, err := config.ElasticClient.
		Get().
		Index(library.IndexName).
		Id(string(id)).
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

func (library *ElasticLibraryDatabase) Search(title, authorName, minPrice, maxPrice string) ([]*models.Book, error) {
	index := config.ElasticClient.Search().Index(library.IndexName).Pretty(false).Size(10000)

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

	index.Query(boolQuery)

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

func (library *ElasticLibraryDatabase) Store() (map[string]interface{}, error) {
	query := config.ElasticClient.Search().
		Index(library.IndexName)

	titleAggregation := elastic.NewCardinalityAggregation().Field("_id")
	authorsAggregation := elastic.NewCardinalityAggregation().Field("author_name.keyword")
	query.Aggregation("number_of_books", titleAggregation)
	query.Aggregation("number_of_authors", authorsAggregation)
	query.Size(0)

	results, err := query.Do(context.Background())
	if err != nil {
		return nil, err
	}

	numberOfBooks, _ := results.Aggregations.Cardinality("number_of_books")
	numberOfAuthors, _ := results.Aggregations.Cardinality("number_of_authors")

	return map[string]interface{}{
		"number_of_books":   numberOfBooks.Value,
		"number_of_authors": numberOfAuthors.Value,
	}, nil
}

func (library *ElasticLibraryDatabase) Create(book *models.Book) (Id, error) {
	doc, err := config.ElasticClient.Index().Index(library.IndexName).BodyJson(book).Do(context.Background())
	return Id(doc.Id), err
}

func (library *ElasticLibraryDatabase) Update(id Id, book *models.Book) error {
	_, err := config.ElasticClient.
		Update().
		Index(library.IndexName).
		Id(string(id)).
		Doc(gin.H{"title": book.Title}).
		Do(context.Background())

	return err
}
