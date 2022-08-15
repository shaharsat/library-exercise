package db

import "gin/models"

type Id string

type LibraryDatabaseManager interface {
	Create(b *models.Book) (Id, error)
	Update(id Id, book *models.Book) error
	Delete(id Id) error
	GetById(id Id) (*models.Book, error)
	Search(title, authorName, minPrice, maxPrice string) ([]*models.Book, error)
	Store()
}
