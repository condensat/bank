package appcontext

import (
	"io/ioutil"
	"os"
	"strings"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func SecretOrPassword(secret string) string {
	content, err := ioutil.ReadFile(secret)
	if err != nil {
		return secret
	}

	return strings.TrimRightFunc(string(content), func(c rune) bool {
		return c == '\r' || c == '\n'
	})
}
