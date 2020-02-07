package utils

import (
	"errors"
	"io"

	"crypto/rand"
)

func GenerateRand(buff []byte) error {
	n, err := io.ReadFull(rand.Reader, buff[:])
	if err != nil {
		return errors.New("Failed to read rand")
	}
	if n != len(buff) {
		return errors.New("Nonce rand not complete")
	}

	return nil
}

func GenerateRandN(n int) []byte {
	buff := make([]byte, n)
	err := GenerateRand(buff)
	if err != nil {
		return nil
	}
	return buff
}
