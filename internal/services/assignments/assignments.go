package assignments

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	apierror "rankode/internal/errors"
	"rankode/internal/models"
	db "rankode/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AssignmentService struct {
	q  db.Querier
	db *pgxpool.Pool
}

func NewAssignmentService(q db.Querier, db *pgxpool.Pool) *AssignmentService {
	return &AssignmentService{q: q, db: db}
}

func (s *AssignmentService) CreateAssignment(ctx context.Context, teacherID int32, dto models.CreateAssignmentDTO) (db.Assignment, error) {
	// Verify teacher is course teacher
	course, err := s.q.GetCourseByID(ctx, dto.CourseID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Assignment{}, apierror.WrapErrorApi(fmt.Errorf("course not found"), 404)
		}
		return db.Assignment{}, apierror.WrapErrorApi(err, 500)
	}

	if course.TeacherID != teacherID {
		return db.Assignment{}, apierror.WrapErrorApi(fmt.Errorf("you are not the teacher of this course"), 403)
	}

	params := db.CreateAssignmentParams{
		CourseID: dto.CourseID,
		Title:    dto.Title,
	}

	if dto.Description != "" {
		params.Description = pgtype.Text{String: dto.Description, Valid: true}
	}

	if dto.StartDate != nil {
		params.StartDate = pgtype.Timestamp{Time: *dto.StartDate, Valid: true}
	}

	if dto.DueDate != nil {
		params.DueDate = pgtype.Timestamp{Time: *dto.DueDate, Valid: true}
	}

	if dto.MaxAttemptsPerTask != nil {
		params.MaxAttemptsPerTask = pgtype.Int4{Int32: *dto.MaxAttemptsPerTask, Valid: true}
	}

	if dto.GroupID != nil {
		params.GroupID = pgtype.Int4{Int32: *dto.GroupID, Valid: true}
	}

	assignment, err := s.q.CreateAssignment(ctx, params)

	if err != nil {
		return db.Assignment{}, apierror.WrapErrorApi(err, 400)
	}

	return assignment, nil
}

func (s *AssignmentService) GetAssignmentByID(ctx context.Context, assignmentID int32) (db.Assignment, error) {
	assignment, err := s.q.GetAssignmentByID(ctx, assignmentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Assignment{}, apierror.WrapErrorApi(fmt.Errorf("assignment not found"), 404)
		}
		return db.Assignment{}, apierror.WrapErrorApi(err, 500)
	}
	return assignment, nil
}

func (s *AssignmentService) GetAssignmentWithTasks(ctx context.Context, assignmentID int32) (models.AssignmentResponse, error) {
	assignment, err := s.GetAssignmentByID(ctx, assignmentID)
	if err != nil {
		return models.AssignmentResponse{}, err
	}

	course, err := s.q.GetCourseByID(ctx, assignment.CourseID)
	if err != nil {
		return models.AssignmentResponse{}, apierror.WrapErrorApi(err, 500)
	}

	taskRows, err := s.q.GetTasksForAssignment(ctx, assignmentID)
	if err != nil {
		return models.AssignmentResponse{}, apierror.WrapErrorApi(err, 500)
	}

	tasks := make([]models.TaskWithOrder, 0, len(taskRows))
	for _, row := range taskRows {
		task := db.Task{
			ID:          row.ID,
			UserID:      row.UserID,
			Title:       row.Title,
			Description: row.Description,
			Difficulty:  row.Difficulty,
			Passes:      row.Passes,
			Score:       row.Score,
			Topics:      row.Topics,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
			CourseID:    row.CourseID,
			IsPublic:    row.IsPublic,
		}
		tasks = append(tasks, models.TaskWithOrder{
			Task:       task,
			OrderIndex: row.OrderIndex,
			Weight:     parseWeight(row.Weight),
		})
	}

	response := models.AssignmentResponse{
		Assignment: assignment,
		Course:     course,
		Tasks:      tasks,
		TotalTasks: len(tasks),
	}

	if assignment.GroupID.Valid {
		group, err := s.q.GetGroupByID(ctx, assignment.GroupID.Int32)
		if err == nil {
			response.Group = &group
		}
	}

	return response, nil
}

func (s *AssignmentService) ListAssignmentsByCourse(ctx context.Context, courseID int32) ([]db.Assignment, error) {
	assignments, err := s.q.ListAssignmentsByCourse(ctx, courseID)
	if err != nil {
		return nil, apierror.WrapErrorApi(err, 500)
	}
	if assignments == nil {
		assignments = make([]db.Assignment, 0)
	}
	return assignments, nil
}

func (s *AssignmentService) ListAssignmentsForStudent(ctx context.Context, studentID int32) ([]db.Assignment, error) {
	assignments, err := s.q.ListAssignmentsForStudent(ctx, studentID)
	if err != nil {
		return nil, apierror.WrapErrorApi(err, 500)
	}
	if assignments == nil {
		assignments = make([]db.Assignment, 0)
	}
	return assignments, nil
}

func (s *AssignmentService) UpdateAssignment(ctx context.Context, assignmentID, teacherID int32, dto models.UpdateAssignmentDTO) error {
	assignment, err := s.GetAssignmentByID(ctx, assignmentID)
	if err != nil {
		return err
	}

	// Verify teacher is course teacher
	course, err := s.q.GetCourseByID(ctx, assignment.CourseID)
	if err != nil {
		return apierror.WrapErrorApi(err, 500)
	}

	if course.TeacherID != teacherID {
		return apierror.WrapErrorApi(fmt.Errorf("you are not the teacher of this course"), 403)
	}

	params := db.UpdateAssignmentParams{
		ID:    assignmentID,
		Title: dto.Title,
	}

	if dto.Description != "" {
		params.Description = pgtype.Text{String: dto.Description, Valid: true}
	}

	if dto.StartDate != nil {
		params.StartDate = pgtype.Timestamp{Time: *dto.StartDate, Valid: true}
	}

	if dto.DueDate != nil {
		params.DueDate = pgtype.Timestamp{Time: *dto.DueDate, Valid: true}
	}

	if dto.MaxAttemptsPerTask != nil {
		params.MaxAttemptsPerTask = pgtype.Int4{Int32: *dto.MaxAttemptsPerTask, Valid: true}
	}

	if dto.GroupID != nil {
		params.GroupID = pgtype.Int4{Int32: *dto.GroupID, Valid: true}
	}

	err = s.q.UpdateAssignment(ctx, params)
	if err != nil {
		return apierror.WrapErrorApi(err, 500)
	}

	return nil
}

func (s *AssignmentService) DeleteAssignment(ctx context.Context, assignmentID, teacherID int32) error {
	assignment, err := s.GetAssignmentByID(ctx, assignmentID)
	if err != nil {
		return err
	}

	// Verify teacher is course teacher
	course, err := s.q.GetCourseByID(ctx, assignment.CourseID)
	if err != nil {
		return apierror.WrapErrorApi(err, 500)
	}

	if course.TeacherID != teacherID {
		return apierror.WrapErrorApi(fmt.Errorf("you are not the teacher of this course"), 403)
	}

	err = s.q.DeleteAssignment(ctx, assignmentID)
	if err != nil {
		return apierror.WrapErrorApi(err, 500)
	}

	return nil
}

func (s *AssignmentService) AddTaskToAssignment(ctx context.Context, assignmentID, teacherID int32, dto models.AddTaskToAssignmentDTO) error {
	assignment, err := s.GetAssignmentByID(ctx, assignmentID)
	if err != nil {
		return err
	}

	// Verify teacher is course teacher
	course, err := s.q.GetCourseByID(ctx, assignment.CourseID)
	if err != nil {
		return apierror.WrapErrorApi(err, 500)
	}

	if course.TeacherID != teacherID {
		return apierror.WrapErrorApi(fmt.Errorf("you are not the teacher of this course"), 403)
	}

	// Verify task exists and belongs to same course (if private)
	task, err := s.q.GetTaskById(ctx, dto.TaskID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apierror.WrapErrorApi(fmt.Errorf("task not found"), 404)
		}
		return apierror.WrapErrorApi(err, 500)
	}

	if task.CourseID.Valid && task.CourseID.Int32 != assignment.CourseID {
		return apierror.WrapErrorApi(fmt.Errorf("task does not belong to this course"), 400)
	}

	params := db.AddTaskToAssignmentParams{
		AssignmentID: assignmentID,
		TaskID:       dto.TaskID,
		OrderIndex:   dto.OrderIndex,
		Weight:       pgtype.Numeric{Int: big.NewInt(int64(dto.Weight * 100)), Exp: -2, Valid: true},
	}

	err = s.q.AddTaskToAssignment(ctx, params)
	if err != nil {
		return apierror.WrapErrorApi(err, 400)
	}

	return nil
}

func (s *AssignmentService) RemoveTaskFromAssignment(ctx context.Context, assignmentID, taskID, teacherID int32) error {
	assignment, err := s.GetAssignmentByID(ctx, assignmentID)
	if err != nil {
		return err
	}

	// Verify teacher is course teacher
	course, err := s.q.GetCourseByID(ctx, assignment.CourseID)
	if err != nil {
		return apierror.WrapErrorApi(err, 500)
	}

	if course.TeacherID != teacherID {
		return apierror.WrapErrorApi(fmt.Errorf("you are not the teacher of this course"), 403)
	}

	err = s.q.RemoveTaskFromAssignment(ctx, db.RemoveTaskFromAssignmentParams{
		AssignmentID: assignmentID,
		TaskID:       taskID,
	})
	if err != nil {
		return apierror.WrapErrorApi(err, 500)
	}

	return nil
}

func (s *AssignmentService) GetAssignmentStats(ctx context.Context, assignmentID int32) (models.AssignmentStats, error) {
	assignment, err := s.GetAssignmentByID(ctx, assignmentID)
	if err != nil {
		return models.AssignmentStats{}, err
	}

	studentCount, err := s.q.CountStudentsInCourse(ctx, assignment.CourseID)
	if err != nil {
		return models.AssignmentStats{}, apierror.WrapErrorApi(err, 500)
	}

	avgGrade, err := s.q.GetAverageGradeForAssignment(ctx, assignmentID)
	if err != nil {
		return models.AssignmentStats{}, apierror.WrapErrorApi(err, 500)
	}

	submissionsCount, err := s.q.CountAssignmentSubmissions(ctx, assignmentID)
	if err != nil {
		return models.AssignmentStats{}, apierror.WrapErrorApi(err, 500)
	}

	completedCount, err := s.q.CountStudentsWhoFinishedAssignment(ctx, assignmentID)
	if err != nil {
		return models.AssignmentStats{}, apierror.WrapErrorApi(err, 500)
	}

	return models.AssignmentStats{
		TotalStudents:    int(studentCount),
		AverageGrade:     parseNumericToFloat(avgGrade),
		SubmissionsCount: int(submissionsCount),
		CompletedCount:   int(completedCount),
	}, nil
}

func parseWeight(w pgtype.Numeric) float64 {
	if !w.Valid {
		return 1.0
	}
	// Convert numeric to float64
	// Simple conversion for demo
	f, _ := w.Float64Value()
	return f.Float64
}

func parseNumericToFloat(n pgtype.Numeric) float64 {
	if !n.Valid {
		return 0.0
	}
	f, _ := n.Float64Value()
	return f.Float64
}
