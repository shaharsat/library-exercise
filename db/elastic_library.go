package db

import "gin/config"

const INDEX_NAME = "books"

var ElasticLibrary *ElasticLibraryManager

func SetupElasticLibrary() {
	elasticClient, err := config.SetupElasticSearch()

	if err != nil {
		panic(err)
	}

	ElasticLibrary = CreateElasticLibrary(INDEX_NAME, elasticClient)
}
