package data

import (
	"context"
	"database/sql"

	"github.com/Broderick-Westrope/flower/gen/model"
	_ "github.com/go-jet/jet/v2/sqlite"
)

type Respository interface {
	CreateTask(ctx context.Context, name string, description string) (*model.Tasks, error)
	ListTasks(ctx context.Context) ([]model.Tasks, error)
	DeleteTask(ctx context.Context, id int) error
	DeleteAllTasks(ctx context.Context) error
}

type repository_Impl struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) Respository {
	return &repository_Impl{
		DB: db,
	}
}
