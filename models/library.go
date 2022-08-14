package models

type Id string

type Library interface {
	Create(b *Book) (Id, error)
	Update(id Id, book *Book) error
	Delete(id Id) error
	GetById(id Id) (*Book, error)
	Search(title, authorName, minPrice, maxPrice string) ([]*Book, error)
	Store()
}
