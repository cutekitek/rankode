package models

import db "rankode/internal/repository"

type EnrollStudentDTO struct {
	UserID int32  `json:"user_id" validate:"required"`
	Role   string `json:"role" validate:"oneof=student ta"`
}

type EnrollmentResponse struct {
	db.Enrollment
	User   db.User   `json:"user"`
	Course db.Course `json:"course"`
}

type StudentInCourse struct {
	db.User
	EnrolledAt string         `json:"enrolled_at"`
	Role       string         `json:"role"`
	Grades     []GradeSummary `json:"grades,omitempty"`
}

type GradeSummary struct {
	AssignmentID    int32   `json:"assignment_id"`
	AssignmentTitle string  `json:"assignment_title"`
	AverageGrade    float64 `json:"average_grade"`
	TasksCount      int     `json:"tasks_count"`
}
