package utils

import (
	"os"
)

func Hostname() string {
	var err error
	host, err := os.Hostname()
	if err != nil {
		host = "Unknown"
	}
	return host
}
