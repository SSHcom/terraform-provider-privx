package utils

import "strings"

func IsPrivxNotFound(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToUpper(err.Error())
	return strings.Contains(s, "OBJECT_NOT_FOUND") ||
		strings.Contains(s, "NOT_FOUND") ||
		strings.Contains(s, "404")
}
