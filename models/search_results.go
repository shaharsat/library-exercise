package models

type AggregationValue struct {
	Value int `json:"value" binding:"required"`
}

type SearchResult struct {
	NumberOfBooks   AggregationValue `json:"number_of_books" binding:"required"`
	NumberOfAuthors AggregationValue `json:"number_of_authors" binding:"required"`
}
