package models

import (
	"fmt"
	"strings"
	"time"
)

type Book struct {
	Title          string  `json:"title" binding:"required"`
	AuthorName     string  `json:"author_name"`
	Price          float64 `json:"price"`
	EbookAvailable bool    `json:"ebook_available"`
	PublishDate    Date    `json:"publish_date"`
}

const PUBLISH_DATE_TIME_FORMAT = "2006-01-02"

type Date time.Time

// UnmarshalJSON Parses the json string in the custom format
func (ct *Date) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), `"`)
	nt, err := time.Parse(PUBLISH_DATE_TIME_FORMAT, s)
	*ct = Date(nt)
	return
}

// MarshalJSON writes a quoted string in the custom format
func (ct Date) MarshalJSON() ([]byte, error) {
	return []byte(ct.String()), nil
}

// String returns the time in the custom format
func (ct *Date) String() string {
	t := time.Time(*ct)
	return fmt.Sprintf("%q", t.Format(PUBLISH_DATE_TIME_FORMAT))
}
