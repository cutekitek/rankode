package models

import db "rankode/internal/repository"

type CreateCourseDTO struct {
	Name        string `json:"name" validate:"required,min=3,max=255"`
	Description string `json:"description" validate:"max=1000"`
}

type UpdateCourseDTO struct {
	Name        string `json:"name" validate:"required,min=3,max=255"`
	Description string `json:"description" validate:"max=1000"`
}

type CourseResponse struct {
	db.Course
	Teacher         db.User `json:"teacher"`
	StudentCount    int     `json:"student_count"`
	AssignmentCount int     `json:"assignment_count"`
}

type JoinCourseDTO struct {
	JoinCode string `json:"join_code" validate:"required,min=6,max=10"`
}

type CourseStats struct {
	TotalStudents     int `json:"total_students"`
	TotalAssignments  int `json:"total_assignments"`
	ActiveAssignments int `json:"active_assignments"`
	CompletedTasks    int `json:"completed_tasks"`
}
