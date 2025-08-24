package utils

import "github.com/google/uuid"

// EqualUUID 判断两个UUID是否相等
func EqualUUID(a, b *uuid.UUID) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
