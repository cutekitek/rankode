package models

import db "rankode/internal/repository"

type CreateTaskDTO struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Difficulty  int32   `json:"difficulty"`
	Topics      []int32 `json:"topics"`
}

type UpdateTaskDTO struct {
	db.UpdateTaskParams
	UserID int
}


type TaskByIdResponse struct {
	db.Task
	TestCases []db.TaskTestCase `json:"test_cases"`
}