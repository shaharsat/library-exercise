package config

import (
	"github.com/olivere/elastic/v7"
	"os"
)

func SetupElasticSearch() (*elastic.Client, error) {
	return elastic.NewClient(elastic.SetURL(os.Getenv("ELASTIC_URL")))
}
