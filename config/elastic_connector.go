package config

import (
	"github.com/olivere/elastic/v7"
	"os"
)

var ElasticClient *elastic.Client

func SetupElasticSearch() {
	elasticUrl := os.Getenv("ELASTIC_URL")

	var err error
	ElasticClient, err = elastic.NewClient(elastic.SetURL(elasticUrl))

	if err != nil {
		panic(err)
	}
}
