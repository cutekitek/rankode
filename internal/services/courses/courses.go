package courses

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	apierror "rankode/internal/errors"
	"rankode/internal/models"
	db "rankode/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CourseService struct {
	q  db.Querier
	db *pgxpool.Pool
}

func NewCourseService(q db.Querier, db *pgxpool.Pool) *CourseService {
	return &CourseService{q: q, db: db}
}

func (s *CourseService) CreateCourse(ctx context.Context, teacherID int32, dto models.CreateCourseDTO) (db.Course, error) {
	joinCode, err := generateJoinCode()
	if err != nil {
		return db.Course{}, apierror.WrapErrorApi(err, 500)
	}

	params := db.CreateCourseParams{
		TeacherID: teacherID,
		Name:      dto.Name,
		JoinCode:  joinCode,
	}

	if dto.Description != "" {
		params.Description = pgtype.Text{String: dto.Description, Valid: true}
	}

	course, err := s.q.CreateCourse(ctx, params)
	if err != nil {
		return db.Course{}, apierror.WrapErrorApi(err, 400)
	}

	return course, nil
}

func (s *CourseService) GetCourseByID(ctx context.Context, courseID int32) (db.Course, error) {
	course, err := s.q.GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Course{}, apierror.WrapErrorApi(fmt.Errorf("course not found"), 404)
		}
		return db.Course{}, apierror.WrapErrorApi(err, 500)
	}
	return course, nil
}

func (s *CourseService) GetCourseWithStats(ctx context.Context, courseID int32) (models.CourseResponse, error) {
	course, err := s.GetCourseByID(ctx, courseID)
	if err != nil {
		return models.CourseResponse{}, err
	}

	teacher, err := s.q.GetUserById(ctx, course.TeacherID)
	if err != nil {
		return models.CourseResponse{}, apierror.WrapErrorApi(err, 500)
	}

	row, err := s.q.GetCourseWithStats(ctx, courseID)
	if err != nil {
		return models.CourseResponse{}, apierror.WrapErrorApi(err, 500)
	}

	return models.CourseResponse{
		Course:          course,
		Teacher:         teacher,
		StudentCount:    int(row.StudentCount),
		AssignmentCount: int(row.AssignmentCount),
	}, nil
}

func (s *CourseService) ListCoursesByTeacher(ctx context.Context, teacherID int32) ([]db.Course, error) {
	courses, err := s.q.ListCoursesByTeacher(ctx, teacherID)
	if err != nil {
		return nil, apierror.WrapErrorApi(err, 500)
	}
	return courses, nil
}

func (s *CourseService) ListCoursesForStudent(ctx context.Context, studentID int32) ([]db.Course, error) {
	courses, err := s.q.ListCoursesForStudent(ctx, studentID)
	if err != nil {
		return nil, apierror.WrapErrorApi(err, 500)
	}
	return courses, nil
}

func (s *CourseService) UpdateCourse(ctx context.Context, courseID, teacherID int32, dto models.UpdateCourseDTO) error {
	params := db.UpdateCourseParams{
		ID:        courseID,
		Name:      dto.Name,
		TeacherID: teacherID,
	}

	if dto.Description != "" {
		params.Description = pgtype.Text{String: dto.Description, Valid: true}
	}

	err := s.q.UpdateCourse(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apierror.WrapErrorApi(fmt.Errorf("course not found or you don't have permission"), 404)
		}
		return apierror.WrapErrorApi(err, 500)
	}
	return nil
}

func (s *CourseService) DeleteCourse(ctx context.Context, courseID, teacherID int32) error {
	err := s.q.DeleteCourse(ctx, db.DeleteCourseParams{
		ID:        courseID,
		TeacherID: teacherID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apierror.WrapErrorApi(fmt.Errorf("course not found or you don't have permission"), 404)
		}
		return apierror.WrapErrorApi(err, 500)
	}
	return nil
}

func (s *CourseService) EnrollStudent(ctx context.Context, courseID, studentID int32, role string) error {
	params := db.EnrollStudentParams{
		CourseID: courseID,
		UserID:   studentID,
		Role:     pgtype.Text{String: role, Valid: true},
	}

	err := s.q.EnrollStudent(ctx, params)
	if err != nil {
		return apierror.WrapErrorApi(err, 400)
	}
	return nil
}

func (s *CourseService) EnrollByJoinCode(ctx context.Context, studentID int32, joinCode string) (db.Course, error) {
	course, err := s.q.GetCourseByJoinCode(ctx, joinCode)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Course{}, apierror.WrapErrorApi(fmt.Errorf("invalid join code"), 404)
		}
		return db.Course{}, apierror.WrapErrorApi(err, 500)
	}

	err = s.EnrollStudent(ctx, course.ID, studentID, "student")
	if err != nil {
		return db.Course{}, err
	}

	return course, nil
}

func (s *CourseService) UnenrollStudent(ctx context.Context, courseID, studentID int32) error {
	err := s.q.UnenrollStudent(ctx, db.UnenrollStudentParams{
		CourseID: courseID,
		UserID:   studentID,
	})
	if err != nil {
		return apierror.WrapErrorApi(err, 500)
	}
	return nil
}

func (s *CourseService) ListStudentsInCourse(ctx context.Context, courseID int32) ([]models.StudentInCourse, error) {
	rows, err := s.q.ListStudentsInCourse(ctx, courseID)
	if err != nil {
		return nil, apierror.WrapErrorApi(err, 500)
	}

	students := make([]models.StudentInCourse, 0, len(rows))
	for _, row := range rows {
		user := db.User{
			ID:           row.ID,
			Username:     row.Username,
			Email:        row.Email,
			PasswordHash: row.PasswordHash,
			Roles:        row.Roles,
			Elo:          row.Elo,
			CreatedAt:    row.CreatedAt,
			UpdatedAt:    row.UpdatedAt,
		}
		students = append(students, models.StudentInCourse{
			User:       user,
			EnrolledAt: row.EnrolledAt.Time.Format(time.RFC3339),
			Role:       row.EnrollmentRole.String,
		})
	}

	return students, nil
}

func (s *CourseService) CheckIfEnrolled(ctx context.Context, courseID, userID int32) (bool, error) {
	enrolled, err := s.q.CheckIfEnrolled(ctx, db.CheckIfEnrolledParams{
		CourseID: courseID,
		UserID:   userID,
	})
	if err != nil {
		return false, apierror.WrapErrorApi(err, 500)
	}
	return enrolled, nil
}

func generateJoinCode() (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const codeLength = 8

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var sb strings.Builder
	sb.Grow(codeLength)

	for i := 0; i < codeLength; i++ {
		sb.WriteByte(charset[r.Intn(len(charset))])
	}

	return sb.String(), nil
}
