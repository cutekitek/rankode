package tasks

import (
	"context"
	"errors"
	"fmt"
	"io"

	apierror "rankode/internal/errors"
	"rankode/internal/models"
	db "rankode/internal/repository"
	"rankode/internal/services/files"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskService struct {
	q     db.DynamicQuerier
	db    *pgxpool.Pool
	files *files.FileStorage
}

func NewTaskService(pool *pgxpool.Pool, files *files.FileStorage) *TaskService {
	return &TaskService{q: db.New(pool), db: pool, files: files}
}

func (s *TaskService) CreateTask(ctx context.Context, params db.CreateTaskParams) (db.Task, error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return db.Task{}, apierror.WrapErrorApi(err, 400)
	}
	defer tx.Rollback(ctx)

	q := db.New(s.db).WithTx(tx)

	if err := s.updateTopicCounters(ctx, q, nil, params.Topics); err != nil {
		return db.Task{}, err
	}
	task, err := q.CreateTask(ctx, params)
	if err != nil {
		return db.Task{}, err
	}

	return task, tx.Commit(ctx)
}

func (s *TaskService) TaskById(ctx context.Context, id int32) (db.Task, error) {
	return s.q.GetTaskById(ctx, id)
}

func (s *TaskService) ListTasksByFilter(ctx context.Context, filter db.TaskListFilter) ([]db.Task, error) {
	tasks, err := s.q.GetTaskListByFilter(ctx, filter)
	if err != nil {
		return nil, err
	}
	if tasks == nil {
		tasks = make([]db.Task, 0)
	}
	return tasks, nil
}

func (s *TaskService) ListTopics(ctx context.Context, filter models.ListTopicsDTO) ([]db.Topic, error) {
	if filter.Name != "" {
		return s.q.TopicsByName(ctx, filter.Name)
	}
	return s.q.ListTopics(ctx)
}

func (s *TaskService) UpdateTask(ctx context.Context, params db.UpdateTaskParams) error {
	return s.q.UpdateTask(ctx, params)
}

func (s *TaskService) UpdateTaskVerificationFile(ctx context.Context, taskId int32, filename string) error {
	return s.q.UpdateTaskVerificationFile(ctx, db.UpdateTaskVerificationFileParams{
		ID:               taskId,
		VerificationFile: pgtype.Text{String: filename, Valid: true},
	})
}

func (s *TaskService) UploadVerificationFile(ctx context.Context, taskId int32, userID int32, file io.Reader, size int64) error {
	task, err := s.q.GetTaskById(ctx, taskId)
	if err != nil {
		return err
	}
	if task.UserID != userID {
		return apierror.WrapErrorApi(errors.New("unauthorized"), 403)
	}

	filename := fmt.Sprintf("verification-%d", taskId)
	err = s.files.UploadFile(ctx, files.UploadFileParams{
		Bucket: "verification",
		Name:   filename,
		File:   file,
		Size:   size,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return s.UpdateTaskVerificationFile(ctx, taskId, filename)
}

func (s *TaskService) DeleteTask(ctx context.Context, id int32) error {
	return s.q.DeleteTask(ctx, id)
}

func (s *TaskService) TopicsByIds(ctx context.Context, ids []int32) ([]db.Topic, error) {
	return s.q.ListTopicsByIDs(ctx, ids)
}

func (s *TaskService) AddTopic(ctx context.Context, name string) (db.Topic, error) {
	return s.q.CreateTopic(ctx, name)
}

func (s *TaskService) updateTopicCounters(ctx context.Context, q db.Querier, removed, added []int32) error {
	all := append(removed, added...)
	topics, err := q.ListTopicsByIDs(ctx, all)
	if err != nil {
		return apierror.WrapErrorApi(err, 400)
	}
	if len(all) != len(topics) {
		return apierror.WrapErrorApi(errors.New("topic not found"), 400)
	}

	if len(removed) > 0 {
		if err := q.UpdateTopicsCounters(ctx, db.UpdateTopicsCountersParams{
			Diff:     -1,
			TopicIds: removed,
		}); err != nil {
			return err
		}
	}
	if len(added) > 0 {
		if err := q.UpdateTopicsCounters(ctx, db.UpdateTopicsCountersParams{
			Diff:     1,
			TopicIds: added,
		}); err != nil {
			return err
		}
	}
	return nil
}
