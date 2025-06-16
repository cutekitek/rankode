package tasksvalidator

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"rankode/internal/models"
	db "rankode/internal/repository"
	"rankode/internal/services/files"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)



type fileStorage interface {
	GetFile(ctx context.Context, params files.GetFileParams) (io.Reader, error)
}

type TasksValidator struct {
	q db.DynamicQuerier
	db *pgxpool.Pool
	files fileStorage
}

func NewTasksValidator(q db.DynamicQuerier, db *pgxpool.Pool, files fileStorage) *TasksValidator {
	return &TasksValidator{
		q:     q,
		db:    db,
		files: files,
	}
}

func (s *TasksValidator) ValidateAndUpdate(ctx context.Context, data models.AttemptResponse) {
	attempt, err := s.q.GetAttemptById(ctx, data.Id)
	if err != nil {
		slog.Error("TasksValidator: failed to get attempt", "error", err)
		return
	}

	if data.Status != models.AttemptStatusSuccessful {
		s.setErrorStatus(ctx, data.Id, data.Status, data.Error)
		return
	}

	testCases, err := s.q.GetTaskTestCases(ctx, attempt.TaskID)
	if err != nil {
		slog.Error("TasksValidator: failed to get test cases", "error", err, "task_id", attempt.TaskID)
		return
	}

	var runningTime int64
	for i, test := range data.Tests {
		outFile, err := s.getOutFile(ctx,  testCases[i].OutputFile)
		if err != nil {
			slog.Error("TasksValidator: failed to get output file", "error", err, "task_id", attempt.TaskID, "filename", testCases[i].OutputFile)
			s.setErrorStatus(ctx, data.Id, models.AttemptStatusInternalError, fmt.Sprintf("Ошибка в тесте #%d", test.CaseId+1))
			return
		}
		fmt.Println(len(outFile), len(test.Output))
		if string(outFile) != test.Output {
			s.setErrorStatus(ctx, data.Id, models.AttemptStatusWrongAnswer, "")
			return
		}
		runningTime += test.ExecutionTime
	}
	
	alreadySubmitted, err := s.q.CheckFirstSuccessfulAttempt(ctx, db.CheckFirstSuccessfulAttemptParams{
		UserID:        attempt.UserID,
		TaskID:        attempt.TaskID,
		AttemptStatus: int32(models.AttemptStatusSuccessful),
	})
	if err != nil {
		slog.Error("TasksValidator: failed check already submitted", "error", err, "attempt_id", attempt.ID)
		s.setErrorStatus(ctx, data.Id, models.AttemptStatusInternalError, "")
		return
	}

	task, err := s.q.GetTaskById(ctx, attempt.TaskID)
	if err != nil {
		slog.Error("TasksValidator: failed to get task", "error", err, "task_id", attempt.TaskID, )
		s.setErrorStatus(ctx, data.Id, models.AttemptStatusInternalError, "")
		return
	}

	err = pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		if err := q.UpdateAttemptStatus(ctx, db.UpdateAttemptStatusParams{
			AttemptStatus: int32(models.AttemptStatusSuccessful),
			RunningTime:   pgtype.Int4{
				Int32: int32(runningTime),
				Valid: true,
			},
			Memory:        pgtype.Int4{
				Int32: int32(data.MemoryUsage),
				Valid: true,
			},
			ID:            data.Id,
		}); err != nil {
			return fmt.Errorf("failed update attempt status: %w",  err)
		}
		if !alreadySubmitted {
			if err := q.IncreaseTaskPasses(ctx, attempt.TaskID); err != nil {
				return fmt.Errorf("failed increase task passes: %w",  err)
			}
			if err := q.IncreaseUserElo(ctx, db.IncreaseUserEloParams{ID: attempt.UserID, Elo: task.Difficulty + 1}); err != nil {
				return fmt.Errorf("failed increase user elo: %w",  err)
			}
		}
		return nil
	})
	if err != nil {
		slog.Error("TasksValidator: failed to set succesful status", "error", err, "task_id", attempt.TaskID, "user_id", attempt.UserID)
		s.setErrorStatus(ctx, data.Id, models.AttemptStatusInternalError, "")
	}
}

func (s *TasksValidator) setErrorStatus(ctx context.Context, id int64, status models.AttemptStatus, errorStr string) {
	if err := s.q.UpdateAttemptStatus(ctx, db.UpdateAttemptStatusParams{
		AttemptStatus: int32(status),
		Error:         pgtype.Text{
			String: errorStr,
			Valid: true,
		},
		ID:            id,
	}); err != nil {
		slog.Error("TasksValidator: failed update attempt status", "error", err)
	}
}

func (s *TasksValidator) getOutFile(ctx context.Context, file string) ([]byte, error) {
	outFile, err := s.files.GetFile(ctx, files.GetFileParams{
		Bucket: "tasks",
		Name:   file,
	})
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(outFile)
	if err != nil {
		return nil, err
	}
	return data, nil
}