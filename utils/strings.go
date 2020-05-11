package utils

import (
	"fmt"
)

func EllipsisCentral(str string, limit int) string {
	if len(str) <= 2*limit {
		return str
	}
	return fmt.Sprintf("%s...%s", str[:limit], str[len(str)-limit:])
}
