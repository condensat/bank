package handlers

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var digits = []rune("123456890")
var digitsCount = len(digits)

func randSeq(n int) string {
	seq := make([]rune, 0, n)
	for len(seq) < n {
		r := digits[rand.Intn(digitsCount)]

		// skip first 0 digit
		if r == '0' && len(seq) == 0 {
			continue
		}

		seq = append(seq, r)
	}
	return string(seq)
}
