package data

import (
	"database/sql"

	_ "github.com/go-jet/jet/v2/sqlite"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}
