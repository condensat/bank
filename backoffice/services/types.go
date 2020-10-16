// Copyright 2020 Condensat Tech <contact@condensat.tech>. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package services

type RequestPaging struct {
	Page         int    `json:"page"`
	PageCount    int    `json:"pageCount"`
	CountPerPage int    `json:"countPerPage"`
	Start        string `json:"start"`
	Next         string `json:"next"`
}
