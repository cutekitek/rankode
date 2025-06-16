package models

import "io"

type NewTestCaseReq  struct{
	TaskID int32 `json:"task_id"`
}

type UploadTestCaseReq struct {
	Type string `query:"type"`
}

type UploadTestCaseFileParams struct {
	UserID int
	TestCaseID int32
	Type string
	Reader io.Reader
	FileSize int64
	
}