package data

import (
	"context"

	"github.com/Broderick-Westrope/flower/gen/model"
	"github.com/Broderick-Westrope/flower/gen/table"
)

func (repo *repository_Impl) CreateTask(ctx context.Context, name string, description string) (*model.Tasks, error) {
	task := &model.Tasks{
		ID:          0,
		Name:        name,
		Description: description,
	}

	err := table.Tasks.
		INSERT(table.Tasks.AllColumns.Except(table.Tasks.ID)).
		MODEL(task).
		RETURNING(table.Tasks.AllColumns).
		QueryContext(ctx, repo.DB, task)
	if err != nil {
		return nil, err
	}
	return task, err
}
