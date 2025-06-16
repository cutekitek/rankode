package test_cases

import (
	"context"
	"errors"
	"fmt"
	apierror "rankode/internal/errors"
	"rankode/internal/models"
	"rankode/internal/repository"
	"rankode/internal/services/files"
	"slices"

	"github.com/jackc/pgx/v5"
)

type taskService interface {
	TaskById(ctx context.Context, id int32) (db.Task, error)
}

type fileService interface{
	UploadFile(ctx context.Context, params files.UploadFileParams) error
}

type TestCasesService struct {
	tasks taskService
	files fileService
	q db.DynamicQuerier
}



func NewTestCasesService(tasks taskService, files fileService, q db.DynamicQuerier) *TestCasesService {
	return &TestCasesService{
		tasks: tasks,
		files: files,
		q: q,
	}
}

func (s *TestCasesService) NewTestCase(ctx context.Context, userID int, req models.NewTestCaseReq) (*db.TaskTestCase, error) {
	task, err := s.tasks.TaskById(ctx, int32(req.TaskID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apierror.WrapErrorApi(fmt.Errorf("task not found"), 404)
		}
		return nil, err
	}
	if task.UserID != int32(userID) {
		return nil, apierror.WrapErrorApi(fmt.Errorf("wrong user"), 401)
	}
	cases, err := s.q.GetTaskTestCases(ctx, int32(req.TaskID))
	if err != nil {
		return nil, fmt.Errorf("GetTaskTestCases: %w", err)
	}
	if len(cases) >= 5 {
		return nil,  apierror.WrapErrorApi(fmt.Errorf("max 5 test cases allowed"), 401)
	}
	var order int32
	if len(cases) > 0 {
		order = slices.MaxFunc(cases, func(a db.TaskTestCase, b db.TaskTestCase) int {
			return int(a.CaseOrder) - int(b.CaseOrder)
		}).CaseOrder + 1
	}
	testCase, err := s.q.CreateTaskTestCase(ctx, db.CreateTaskTestCaseParams{
		TaskID:    int32(req.TaskID),
		CaseOrder: order,
		InputFile:  fmt.Sprintf("test_case_input_%d_%d", req.TaskID, order),
		OutputFile:  fmt.Sprintf("test_case_output_%d_%d", req.TaskID, order),
	})
	return &testCase, err
}

func (s *TestCasesService) GetTestCasesByTaskID(ctx context.Context, taskID int32) ([]db.TaskTestCase, error) {
	return s.q.GetTaskTestCases(ctx, taskID)
}  

func (s *TestCasesService) UploadTestCaseFile(ctx context.Context, params models.UploadTestCaseFileParams) error {
	if params.FileSize > 20 * 1024 * 1024 {
		return apierror.WrapErrorApi(fmt.Errorf("file size cannot be greater than 20 MB"), 404)
	}
	testCase, err := s.q.GetTestCaseByID(ctx, int64(params.TestCaseID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apierror.WrapErrorApi(fmt.Errorf("task not found"), 404)
		}
		return err
	}
	task, err := s.tasks.TaskById(ctx, int32(testCase.TaskID))
	if err != nil {
		return err
	}
	if task.UserID != int32(params.UserID) {
		return apierror.WrapErrorApi(fmt.Errorf("wrong user"), 401)
	}
	uploadParams := files.UploadFileParams{
		Bucket: "tasks",
		File:   params.Reader,
		Size:   params.FileSize,
	}
	switch params.Type {
	case "input":
		uploadParams.Name = testCase.InputFile
	case "output":
		uploadParams.Name = testCase.OutputFile
	default:
		return apierror.WrapErrorApi(fmt.Errorf("wrong upload type"), 400)
	}
	return s.files.UploadFile(ctx, uploadParams)
}