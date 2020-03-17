package utils

import (
	"time"

	"math/rand"
)

const (
	cstMemzeroScramble = true
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func Memzero(buff []byte) {
	memreset(buff, cstMemzeroScramble)
}

func memreset(buff []byte, scramble bool) {
	for i := 0; i < len(buff); i++ {
		buff[i] = 0
	}
	if !scramble {
		return
	}
	// do not use crypto/rand
	_, _ = rand.Read(buff)
}
