package db

type TaskListFilter struct {
	TopicIDs     []int   `json:"topics" query:"topics" validate:"omitempty,max=20"`
	Difficulties []int   `json:"difficulties" query:"difficulties" validate:"omitempty"`
	Title        string  `json:"title" query:"title"`
	SortBy       *string `json:"sort" query:"sort" validate:"omitempty,oneof=name difficulty score"`
	Limit        int     `json:"limit" query:"limit" validate:"min=10,max=25"`
	Offset       int     `json:"offset" query:"offset"`
	CourseID     *int32  `json:"course_id" query:"course_id"`
	IsPublic     *bool   `json:"is_public" query:"is_public"`
	UserID       *int32  `json:"user_id" query:"user_id"`
}
