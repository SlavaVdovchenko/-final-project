package models

import (
	"time"
)

type ListDB struct {
	ID         int
	Date       time.Time
	Title      string
	Comment    string
	Repeat     string
	RepeatRule string
	Interval   int
}

type List struct {
	ID      int    `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title" validate:"required,min=1"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

type GetByIdRequest struct {
	ID int `json:"id"`
}

type ListPostResponse struct {
	ID int `json:"id"`
}

type ListGetResponse struct {
	Tasks []*List `json:"tasks"`
}

type ListPutResponse struct {
	ID int `json:"id"`
}
