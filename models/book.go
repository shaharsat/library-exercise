package models

import (
	"fmt"
	"strings"
	"time"
)

type Book struct {
	Id             string     `json:"index._id" binding:"required"`
	Title          string     `json:"title" binding:"required"`
	AuthorName     string     `json:"author_name"  binding:"required"`
	Price          float64    `json:"price"  binding:"required"`
	EbookAvailable bool       `json:"ebook_available"  binding:"required"`
	PublishDate    CustomTime `json:"publish_date" binding:"required"`
}

type CustomTime time.Time

const ctLayout = "2006-01-02"

// UnmarshalJSON Parses the json string in the custom format
func (ct *CustomTime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), `"`)
	nt, err := time.Parse(ctLayout, s)
	*ct = CustomTime(nt)
	return
}

// MarshalJSON writes a quoted string in the custom format
func (ct CustomTime) MarshalJSON() ([]byte, error) {
	return []byte(ct.String()), nil
}

// String returns the time in the custom format
func (ct *CustomTime) String() string {
	t := time.Time(*ct)
	return fmt.Sprintf("%q", t.Format(ctLayout))
}
