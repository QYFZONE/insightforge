package session

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func newID(prefix string) string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%s_fallback", prefix)
	}
	return prefix + "_" + hex.EncodeToString(b[:])
}
