package research

import "errors"

var (
	// ErrEmptyContent 表示用户消息去掉空白字符后为空。
	ErrEmptyContent = errors.New("research: content is empty")
)
