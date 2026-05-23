package models

type ListTopicsDTO struct {
	Name string `query:"name"`
}

type AddTopicDTO struct {
	Name string `json:"name" query:"name"`
}
