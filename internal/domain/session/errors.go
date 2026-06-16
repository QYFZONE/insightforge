package session

import "errors"

// ErrNotFound 表示指定 session 不存在。
var ErrNotFound = errors.New("session not found")
