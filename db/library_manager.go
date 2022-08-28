package db

import "gin/models"

type LibraryManager interface {
	Create(b *models.Book) (string, error)
	Update(id string, book *models.Book) error
	Delete(id string) error
	GetById(id string) (*models.Book, error)
	Search(title, authorName, minPrice, maxPrice string) ([]*models.Book, error)
	Store()
}
