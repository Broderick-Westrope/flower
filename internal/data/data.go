package data

import (
	"database/sql"

	_ "github.com/go-jet/jet/v2/sqlite"
)

type Respository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Respository {
	return &Respository{
		db: db,
	}
}
