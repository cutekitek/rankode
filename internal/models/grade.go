package models

import db "rankode/internal/repository"

type CreateGradeDTO struct {
	AssignmentID int32  `json:"assignment_id" validate:"required"`
	TaskID       int32  `json:"task_id" validate:"required"`
	UserID       int32  `json:"user_id" validate:"required"`
	Grade        int16  `json:"grade" validate:"required,min=1,max=5"`
	Feedback     string `json:"feedback" validate:"max=1000"`
}

type UpdateGradeDTO struct {
	Grade    int16  `json:"grade" validate:"required,min=1,max=5"`
	Feedback string `json:"feedback" validate:"max=1000"`
}

type GradeResponse struct {
	db.Grade
	StudentUsername  string `json:"student_username"`
	TaskTitle        string `json:"task_title"`
	AssignmentTitle  string `json:"assignment_title"`
	GradedByUsername string `json:"graded_by_username,omitempty"`
}

type GradeStats struct {
	AverageGrade      float64       `json:"average_grade"`
	HighestGrade      int16         `json:"highest_grade"`
	LowestGrade       int16         `json:"lowest_grade"`
	GradeDistribution map[int16]int `json:"grade_distribution"`
}
