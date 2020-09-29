package services

type RequestPaging struct {
	Page      int `json:"page"`
	PageCount int `json:"pageCount"`
}
