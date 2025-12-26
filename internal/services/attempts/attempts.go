package attempts

import (
	"context"
	"errors"
	"fmt"

	apierror "rankode/internal/errors"
	"rankode/internal/mappers"
	"rankode/internal/models"
	db "rankode/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AttemptsQueue interface {
	SendAttempt(models.AttemptRequest) error
}

type AttemptsService struct {
	q     db.DynamicQuerier
	queue AttemptsQueue
	db    *pgxpool.Pool
}

func NewAttemptsService(q db.DynamicQuerier, db *pgxpool.Pool, queue AttemptsQueue) *AttemptsService {
	return &AttemptsService{
		q:     q,
		queue: queue,
		db:    db,
	}
}

func (s *AttemptsService) NewAttempt(ctx context.Context, user db.User, params models.CreateAttemptRequest) error {
	testCases, err := s.q.GetTaskTestCases(ctx, int32(params.TaskID))
	if err != nil {
		return fmt.Errorf("GetTaskTestCases: %w", err)
	}
	if len(testCases) == 0 {
		return apierror.WrapErrorApi(errors.New("task not found"), 404)
	}
	return pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		assignmentID := pgtype.Int4{Valid: false}
		if params.AssignmentID != nil {
			assignmentID = pgtype.Int4{Int32: int32(*params.AssignmentID), Valid: true}
		}

		attempts, err := q.GetUserAttemptsByTask(ctx, db.GetUserAttemptsByTaskParams{
			UserID:       user.ID,
			TaskID:       int32(params.TaskID),
			AssignmentID: assignmentID,
		})
		if err != nil {
			return fmt.Errorf("GetUserAttemptsByTask: %w", err)
		}
		for _, a := range attempts {
			if a.AttemptStatus == int32(models.AttemptStatusCreated) {
				return apierror.WrapErrorApi(errors.New("still running"), 400)
			}
		}

		attempt, err := q.CreateAttempt(ctx, db.CreateAttemptParams{
			UserID:        user.ID,
			TaskID:        int32(params.TaskID),
			Code:          params.Code,
			AttemptStatus: int32(models.AttemptStatusCreated),
			Language:      params.Language,
			AssignmentID:  assignmentID,
		})
		if err != nil {
			return fmt.Errorf("CreateAttempt: %w", err)
		}
		checkReq := models.AttemptRequest{
			Id:            attempt.ID,
			Language:      params.Language,
			Code:          params.Code,
			MemoryLimit:   256 * 1024 * 1024,
			Timeout:       1000,
			MaxOutputSize: 10 * 1024 * 1024,
			TestCases:     mappers.DbTestCasesToModelTestCase(testCases),
		}
		return s.queue.SendAttempt(checkReq)
	})
}

func (s *AttemptsService) GetUserTaskAttempts(ctx context.Context, user db.User, taskId int32, assignmentID *int32) ([]db.Attempt, error) {
	params := db.GetUserAttemptsByTaskParams{
		UserID: user.ID,
		TaskID: taskId,
	}
	if assignmentID != nil {
		params.AssignmentID = pgtype.Int4{Int32: *assignmentID, Valid: true}
	}
	return s.q.GetUserAttemptsByTask(ctx, params)
}
