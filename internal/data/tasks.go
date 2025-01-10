package data

import (
	"context"
	"errors"
	"fmt"

	"github.com/Broderick-Westrope/flower/gen/model"
	"github.com/Broderick-Westrope/flower/gen/table"
	"github.com/Broderick-Westrope/flower/internal"
	"github.com/go-jet/jet/v2/sqlite"
)

type Task struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Parent      *Task  `json:"parent"`
}

func (t *Task) toDataModel() *model.Tasks {
	var parentID *int32
	if t.Parent != nil {
		parentID = internal.ToPointer(int32(t.Parent.ID))
	}

	return &model.Tasks{
		ID:          int32(t.ID),
		Name:        t.Name,
		Description: t.Description,
		ParentID:    parentID,
	}
}

func (repo *Repository) taskFromDataModel(ctx context.Context, task *model.Tasks) (*Task, error) {
	// Create a map to track visited task IDs to prevent infinite recursion
	// in case the database somehow contains a cycle
	visited := make(map[int32]struct{})
	return repo.taskFromDataModelWithVisited(ctx, task, visited)
}

func (repo *Repository) taskFromDataModelWithVisited(ctx context.Context, task *model.Tasks, visited map[int32]struct{}) (*Task, error) {
	// Check if we've already visited this task
	if _, exists := visited[task.ID]; exists {
		return nil, fmt.Errorf("cycle detected in task hierarchy at task %d", task.ID)
	}

	// Mark this task as visited
	visited[task.ID] = struct{}{}

	var parent *Task
	if task.ParentID != nil {
		// Get the parent task from the database
		parentModel := new(model.Tasks)
		err := table.Tasks.
			SELECT(table.Tasks.AllColumns).
			WHERE(table.Tasks.ID.EQ(sqlite.Int(int64(*task.ParentID)))).
			QueryContext(ctx, repo.db, parentModel)
		if err != nil {
			if isNotFoundError(err) {
				return nil, fmt.Errorf("parent task %d not found", *task.ParentID)
			}
			return nil, fmt.Errorf("retrieving parent task %d: %w", *task.ParentID, err)
		}

		// Recursively get the parent's parent
		parent, err = repo.taskFromDataModelWithVisited(ctx, parentModel, visited)
		if err != nil {
			return nil, err
		}
	}

	return &Task{
		ID:          int(task.ID),
		Name:        task.Name,
		Description: task.Description,
		Parent:      parent,
	}, nil
}

func (repo *Repository) taskSliceFromDataModelSlice(ctx context.Context, tasks []model.Tasks) ([]Task, error) {
	out := make([]Task, 0)
	for _, t := range tasks {
		newTask, err := repo.taskFromDataModel(ctx, &t)
		if err != nil {
			return nil, err
		}
		out = append(out, *newTask)
	}
	return out, nil
}

func (repo *Repository) CreateTask(ctx context.Context, taskIn *Task) (*Task, error) {
	task := taskIn.toDataModel()

	err := table.Tasks.
		INSERT(table.Tasks.AllColumns.Except(table.Tasks.ID)).
		MODEL(task).
		RETURNING(table.Tasks.AllColumns).
		QueryContext(ctx, repo.db, task)
	if err != nil {
		return nil, err
	}

	result, err := repo.taskFromDataModel(ctx, task)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (repo *Repository) GetTask(ctx context.Context, id int) (*Task, error) {
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

	result, err := repo.taskFromDataModel(ctx, task)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (repo *Repository) ListTasks(ctx context.Context) ([]Task, error) {
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

	result, err := repo.taskSliceFromDataModelSlice(ctx, tasks)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (repo *Repository) DeleteTask(ctx context.Context, id int) error {
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

func (repo *Repository) DeleteAllTasks(ctx context.Context) error {
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

func (repo *Repository) DetectParentTaskCycle(ctx context.Context, parentID int) (bool, error) {
	currentID := parentID
	visited := make(map[int]struct{})

	// Follow the parent chain up until we either:
	// 1. Find no more parents (DAG preserved)
	// 2. Find the original parentID (cycle detected)
	// 3. Visit the same node twice (cycle detected)
	for currentID != 0 {
		// Check if we've seen this node before
		if _, exists := visited[currentID]; exists {
			return true, nil
		}
		visited[currentID] = struct{}{}

		// Get the current task
		task, err := repo.GetTask(ctx, currentID)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return false, nil
			}
			return false, err
		}

		// Check if we've found our original task
		if task.Parent != nil && task.Parent.ID == parentID {
			return true, nil
		}

		// Move up to the parent
		if task.Parent == nil {
			break
		}
		currentID = task.Parent.ID
	}
	return false, nil
}
