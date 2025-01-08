package data

import (
	"context"

	"github.com/Broderick-Westrope/flower/gen/model"
	"github.com/Broderick-Westrope/flower/gen/table"
	"github.com/go-jet/jet/v2/sqlite"
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

func (repo *repository_Impl) ListTasks(ctx context.Context) ([]model.Tasks, error) {
	var tasks []model.Tasks

	err := table.Tasks.
		SELECT(table.Tasks.AllColumns).
		QueryContext(ctx, repo.DB, &tasks)
	if err != nil {
		return nil, err
	}
	return tasks, err

}

func (repo *repository_Impl) DeleteTask(ctx context.Context, id int) error {
	_, err := table.Tasks.
		DELETE().
		WHERE(table.Tasks.ID.EQ(sqlite.Int(int64(id)))).
		ExecContext(ctx, repo.DB)
	if err != nil {
		if isNotFoundError(err) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (repo *repository_Impl) DeleteAllTasks(ctx context.Context) error {
	_, err := table.Tasks.
		DELETE().
		WHERE(sqlite.Bool(true)).
		ExecContext(ctx, repo.DB)
	if err != nil {
		return err
	}
	return nil
}
