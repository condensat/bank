package services

type RequestPaging struct {
	Page         int    `json:"page"`
	PageCount    int    `json:"pageCount"`
	CountPerPage int    `json:"countPerPage"`
	Start        string `json:"start"`
	Next         string `json:"next"`
}
