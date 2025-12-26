package models

import db "rankode/internal/repository"

type CreateGroupDTO struct {
	CourseID int32  `json:"course_id" validate:"required"`
	Name     string `json:"name" validate:"required,min=2,max=255"`
}

type GroupResponse struct {
	db.Group
	StudentCount int `json:"student_count"`
}

type AddStudentToGroupDTO struct {
	UserID int32 `json:"user_id" validate:"required"`
}
