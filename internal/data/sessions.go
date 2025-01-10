package data

import (
	"context"
	"fmt"
	"time"

	"github.com/Broderick-Westrope/flower/gen/model"
	"github.com/Broderick-Westrope/flower/gen/table"
	"github.com/Broderick-Westrope/flower/internal"

	"github.com/go-jet/jet/v2/sqlite"
)

type Session struct {
	ID        int        `json:"id"`
	Task      *Task      `json:"task"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at"`
}

func (repo *Respository) sessionFromDataModel(ctx context.Context, session *model.Sessions) (*Session, error) {
	task, err := repo.GetTask(ctx, int(session.TaskID))
	if err != nil {
		return nil, fmt.Errorf("retrieving task %d associated with session %d: %w", session.TaskID, session.ID, err)
	}

	var endedAt *time.Time
	if session.EndedAt != 0 {
		endedAt = internal.ToPointer(time.Unix(int64(session.EndedAt), 0))
	}

	return &Session{
		ID:        int(session.ID),
		Task:      task,
		StartedAt: time.Unix(int64(session.StartedAt), 0),
		EndedAt:   endedAt,
	}, nil
}

func (repo *Respository) sessionSliceFromDataModelSlice(ctx context.Context, sessions []model.Sessions) ([]Session, error) {
	out := make([]Session, 0)
	for _, dmSession := range sessions {
		newSession, err := repo.sessionFromDataModel(ctx, &dmSession)
		if err != nil {
			return nil, err
		}
		out = append(out, *newSession)
	}
	return out, nil
}

func (repo *Respository) StartSession(ctx context.Context, taskID int) (*Session, error) {
	session := &model.Sessions{
		ID:        0,
		TaskID:    int32(taskID),
		StartedAt: int32(time.Now().Unix()),
		EndedAt:   0,
	}

	err := table.Sessions.
		INSERT(table.Sessions.AllColumns.Except(table.Sessions.ID)).MODEL(session).
		RETURNING(table.Sessions.AllColumns).
		QueryContext(ctx, repo.db, session)
	if err != nil {
		return nil, err
	}

	result, err := repo.sessionFromDataModel(ctx, session)
	if err != nil {
		return nil, err
	}
	return result, err
}

func (repo *Respository) StopSession(ctx context.Context, id int) (*Session, error) {
	session := &model.Sessions{
		EndedAt: int32(time.Now().Unix()),
	}

	err := table.Sessions.
		UPDATE(table.Sessions.EndedAt).MODEL(session).
		WHERE(table.Sessions.ID.EQ(sqlite.Int(int64(id)))).
		RETURNING(table.Sessions.AllColumns).
		QueryContext(ctx, repo.db, session)
	if err != nil {
		if isNotFoundError(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	result, err := repo.sessionFromDataModel(ctx, session)
	if err != nil {
		return nil, err
	}
	return result, err
}

func (repo *Respository) ListOpenSessions(ctx context.Context) ([]Session, error) {
	sessions := make([]model.Sessions, 0)

	err := table.Sessions.
		SELECT(table.Sessions.AllColumns).
		WHERE(table.Sessions.EndedAt.EQ(sqlite.Int(0))).
		ORDER_BY(table.Sessions.StartedAt.ASC()).
		QueryContext(ctx, repo.db, &sessions)
	if err != nil {
		if isNotFoundError(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	result, err := repo.sessionSliceFromDataModelSlice(ctx, sessions)
	if err != nil {
		return nil, err
	}
	return result, err
}

func (repo *Respository) ListClosedSessions(ctx context.Context) ([]Session, error) {
	sessions := make([]model.Sessions, 0)

	err := table.Sessions.
		SELECT(table.Sessions.AllColumns).
		WHERE(table.Sessions.EndedAt.NOT_EQ(sqlite.Int(0))).
		ORDER_BY(table.Sessions.StartedAt.ASC()).
		QueryContext(ctx, repo.db, &sessions)
	if err != nil {
		if isNotFoundError(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	result, err := repo.sessionSliceFromDataModelSlice(ctx, sessions)
	if err != nil {
		return nil, err
	}
	return result, err
}

func (repo *Respository) ListSessions(ctx context.Context) ([]Session, error) {
	sessions := make([]model.Sessions, 0)

	err := table.Sessions.
		SELECT(table.Sessions.AllColumns).
		ORDER_BY(table.Sessions.StartedAt.ASC()).
		QueryContext(ctx, repo.db, &sessions)
	if err != nil {
		if isNotFoundError(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	result, err := repo.sessionSliceFromDataModelSlice(ctx, sessions)
	if err != nil {
		return nil, err
	}
	return result, err
}
