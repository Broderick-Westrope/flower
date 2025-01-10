package data

import (
	"context"

	"github.com/Broderick-Westrope/flower/gen/model"
	"github.com/Broderick-Westrope/flower/gen/table"
	"github.com/go-jet/jet/v2/sqlite"
)

type Task struct {
	ID          int
	Name        string
	Description string
}

func taskFromDataModel(task *model.Tasks) *Task {
	return &Task{
		ID:          int(task.ID),
		Name:        task.Name,
		Description: task.Description,
	}
}

func taskSliceFromDataModelSlice(tasks []model.Tasks) []Task {
	out := make([]Task, 0)
	for _, t := range tasks {
		out = append(out, *taskFromDataModel(&t))
	}
	return out
}

func (repo *Respository) CreateTask(ctx context.Context, name string, description string) (*Task, error) {
	task := &model.Tasks{
		ID:          0,
		Name:        name,
		Description: description,
	}

	err := table.Tasks.
		INSERT(table.Tasks.AllColumns.Except(table.Tasks.ID)).
		MODEL(task).
		RETURNING(table.Tasks.AllColumns).
		QueryContext(ctx, repo.db, task)
	if err != nil {
		return nil, err
	}
	return taskFromDataModel(task), err
}

func (repo *Respository) GetTask(ctx context.Context, id int) (*Task, error) {
	task := new(model.Tasks)

	err := table.Tasks.
		SELECT(table.Tasks.AllColumns).
		WHERE(table.Tasks.ID.EQ(sqlite.Int(int64(id)))).
		QueryContext(ctx, repo.db, task)
	if err != nil {
		if isNotFoundError(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return taskFromDataModel(task), err
}

func (repo *Respository) ListTasks(ctx context.Context) ([]Task, error) {
	tasks := make([]model.Tasks, 0)

	err := table.Tasks.
		SELECT(table.Tasks.AllColumns).
		QueryContext(ctx, repo.db, &tasks)
	if err != nil {
		if isNotFoundError(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return taskSliceFromDataModelSlice(tasks), err
}

func (repo *Respository) DeleteTask(ctx context.Context, id int) error {
	_, err := table.Tasks.
		DELETE().
		WHERE(table.Tasks.ID.EQ(sqlite.Int(int64(id)))).
		ExecContext(ctx, repo.db)
	if err != nil {
		if isNotFoundError(err) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (repo *Respository) DeleteAllTasks(ctx context.Context) error {
	_, err := table.Tasks.
		DELETE().
		WHERE(sqlite.Bool(true)).
		ExecContext(ctx, repo.db)
	if err != nil {
		if isNotFoundError(err) {
			return ErrNotFound
		}
		return err
	}
	return nil
}
