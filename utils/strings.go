package utils

import (
	"fmt"
	"strings"
)

const (
	Ellipsis = "..."
)

func EllipsisCentral(str string, limit int) string {
	if len(str) <= 2*limit {
		return str
	}
	return fmt.Sprintf("%s%s%s", str[:limit], Ellipsis, str[len(str)-limit:])
}

func ContainEllipsis(str string) bool {
	return strings.Contains(str, Ellipsis)
}

func SplitEllipsis(str string) []string {
	return strings.Split(str, Ellipsis)
}
