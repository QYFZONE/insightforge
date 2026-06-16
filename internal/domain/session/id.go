package session

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// newID 生成带业务前缀的短随机 ID。
func newID(prefix string) string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%s_fallback", prefix)
	}
	return prefix + "_" + hex.EncodeToString(b[:])
}

// NewID 为领域对象生成统一格式的 ID。
func NewID(prefix string) string {
	return newID(prefix)
}
