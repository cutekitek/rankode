package groups

import (
	"context"
	"fmt"
	apierror "rankode/internal/errors"
	"rankode/internal/models"
	db "rankode/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

type GroupService struct {
	q  db.Querier
	db *pgxpool.Pool
}

func NewGroupService(q db.Querier, db *pgxpool.Pool) *GroupService {
	return &GroupService{q: q, db: db}
}

func (s *GroupService) CreateGroup(ctx context.Context, teacherID int32, dto models.CreateGroupDTO) (db.Group, error) {
	// Verify teacher
	course, err := s.q.GetCourseByID(ctx, dto.CourseID)
	if err != nil {
		return db.Group{}, apierror.WrapErrorApi(err, 404)
	}
	if course.TeacherID != teacherID {
		return db.Group{}, apierror.WrapErrorApi(fmt.Errorf("not the teacher of this course"), 403)
	}

	return s.q.CreateGroup(ctx, db.CreateGroupParams{
		CourseID: dto.CourseID,
		Name:     dto.Name,
	})
}

func (s *GroupService) ListGroupsByCourse(ctx context.Context, courseID int32) ([]db.Group, error) {
	return s.q.ListGroupsByCourse(ctx, courseID)
}

func (s *GroupService) AddStudentToGroup(ctx context.Context, teacherID int32, groupID int32, studentID int32) error {
	group, err := s.q.GetGroupByID(ctx, groupID)
	if err != nil {
		return apierror.WrapErrorApi(err, 404)
	}
	course, err := s.q.GetCourseByID(ctx, group.CourseID)
	if err != nil {
		return apierror.WrapErrorApi(err, 500)
	}
	if course.TeacherID != teacherID {
		return apierror.WrapErrorApi(fmt.Errorf("not the teacher of this course"), 403)
	}

	return s.q.AddStudentToGroup(ctx, db.AddStudentToGroupParams{
		GroupID: groupID,
		UserID:  studentID,
	})
}

func (s *GroupService) RemoveStudentFromGroup(ctx context.Context, teacherID int32, groupID int32, studentID int32) error {
	group, err := s.q.GetGroupByID(ctx, groupID)
	if err != nil {
		return apierror.WrapErrorApi(err, 404)
	}
	course, err := s.q.GetCourseByID(ctx, group.CourseID)
	if err != nil {
		return apierror.WrapErrorApi(err, 500)
	}
	if course.TeacherID != teacherID {
		return apierror.WrapErrorApi(fmt.Errorf("not the teacher of this course"), 403)
	}

	return s.q.RemoveStudentFromGroup(ctx, db.RemoveStudentFromGroupParams{
		GroupID: groupID,
		UserID:  studentID,
	})
}

func (s *GroupService) ListStudentsInGroup(ctx context.Context, groupID int32) ([]db.User, error) {
	return s.q.ListStudentsInGroup(ctx, groupID)
}

func (s *GroupService) DeleteGroup(ctx context.Context, teacherID int32, groupID int32) error {
	group, err := s.q.GetGroupByID(ctx, groupID)
	if err != nil {
		return apierror.WrapErrorApi(err, 404)
	}
	course, err := s.q.GetCourseByID(ctx, group.CourseID)
	if err != nil {
		return apierror.WrapErrorApi(err, 500)
	}
	if course.TeacherID != teacherID {
		return apierror.WrapErrorApi(fmt.Errorf("not the teacher of this course"), 403)
	}
	return s.q.DeleteGroup(ctx, groupID)
}
