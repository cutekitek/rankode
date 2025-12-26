package models

import (
	"time"

	db "rankode/internal/repository"
)

type CreateAssignmentDTO struct {
	CourseID           int32      `json:"course_id" validate:"required"`
	Title              string     `json:"title" validate:"required,min=3,max=255"`
	Description        string     `json:"description" validate:"max=1000"`
	StartDate          *time.Time `json:"start_date"`
	DueDate            *time.Time `json:"due_date"`
	MaxAttemptsPerTask *int32     `json:"max_attempts_per_task" validate:"min=1,max=100"`
	GroupID            *int32     `json:"group_id"`
}

type UpdateAssignmentDTO struct {
	Title              string     `json:"title" validate:"required,min=3,max=255"`
	Description        string     `json:"description" validate:"max=1000"`
	StartDate          *time.Time `json:"start_date"`
	DueDate            *time.Time `json:"due_date"`
	MaxAttemptsPerTask *int32     `json:"max_attempts_per_task" validate:"min=1,max=100"`
	GroupID            *int32     `json:"group_id"`
}

type AssignmentResponse struct {
	db.Assignment
	Course     db.Course       `json:"course"`
	Tasks      []TaskWithOrder `json:"tasks"`
	TotalTasks int             `json:"total_tasks"`
	Group      *db.Group       `json:"group,omitempty"`
}

type TaskWithOrder struct {
	db.Task
	OrderIndex int32   `json:"order_index"`
	Weight     float64 `json:"weight"`
}

type AddTaskToAssignmentDTO struct {
	TaskID     int32   `json:"task_id" validate:"required"`
	OrderIndex int32   `json:"order_index" validate:"min=0"`
	Weight     float64 `json:"weight" validate:"min=0.1,max=5.0"`
}

type AssignmentStats struct {
	TotalStudents    int     `json:"total_students"`
	SubmissionsCount int     `json:"submissions_count"`
	AverageGrade     float64 `json:"average_grade"`
	CompletedCount   int     `json:"completed_count"`
}
