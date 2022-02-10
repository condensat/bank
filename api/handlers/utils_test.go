package handlers

import (
	"math/rand"
	"testing"
)

func Test_randSeq(t *testing.T) {
	for i := 1; i < 1000; i++ {
		l := 4 + rand.Intn(32-4)
		seq := randSeq(l)
		if len(seq) != l {
			t.Errorf("Invalid randSeq() len = %v, want %v", len(seq), l)
		}
		if seq[0] == '0' {
			t.Errorf("Invalid randSeq() %v", seq)
		}
	}
}
