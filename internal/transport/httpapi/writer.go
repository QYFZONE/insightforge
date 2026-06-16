package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"insightforge/internal/app/research"
	"insightforge/internal/domain/session"
)

const (
	errorCodeEmptyContent         = "EMPTY_CONTENT"
	errorCodeInternalServer       = "INTERNAL_SERVER_ERROR"
	errorCodeInvalidJSON          = "INVALID_JSON"
	errorCodeNotFound             = "NOT_FOUND"
	errorCodeSessionNotFound      = "SESSION_NOT_FOUND"
	errorCodeStreamingUnsupported = "STREAMING_UNSUPPORTED"
)

// writeSSE 按 SSE 协议写入一条事件。
func writeSSE(w http.ResponseWriter, event session.Event) error {
	data, err := json.Marshal(toEventResponse(event))
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "event: %s\n", event.Type); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
		return err
	}
	return nil
}

// writeSessionError 把 session 领域错误翻译为 HTTP 错误响应。
func writeSessionError(w http.ResponseWriter, err error) {
	if errors.Is(err, session.ErrNotFound) {
		writeError(w, http.StatusNotFound, errorCodeSessionNotFound, "会话不存在")
		return
	}
	writeInternalError(w, err)
}

// writeResearchError 把研究业务层错误翻译为 HTTP 错误响应。
func writeResearchError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, session.ErrNotFound):
		writeError(w, http.StatusNotFound, errorCodeSessionNotFound, "会话不存在")
	case errors.Is(err, research.ErrEmptyContent):
		writeError(w, http.StatusBadRequest, errorCodeEmptyContent, "消息内容不能为空")
	default:
		writeInternalError(w, err)
	}
}

// writeInvalidJSON 返回请求体解析失败错误。
func writeInvalidJSON(w http.ResponseWriter) {
	writeError(w, http.StatusBadRequest, errorCodeInvalidJSON, "请求体不是合法 JSON")
}

// writeRouteNotFound 返回统一的路由不存在错误。
func writeRouteNotFound(w http.ResponseWriter) {
	writeError(w, http.StatusNotFound, errorCodeNotFound, "接口不存在")
}

// writeStreamingUnsupported 返回当前连接不支持流式响应的错误。
func writeStreamingUnsupported(w http.ResponseWriter) {
	writeError(w, http.StatusInternalServerError, errorCodeStreamingUnsupported, "当前连接不支持流式响应")
}

// writeInternalError 记录内部错误，但对用户隐藏具体实现细节。
func writeInternalError(w http.ResponseWriter, err error) {
	log.Printf("internal error: %v", err)
	writeError(w, http.StatusInternalServerError, errorCodeInternalServer, "服务器内部错误")
}

// writeError 写入统一结构的错误响应。
func writeError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, errorResponse{
		Code:    code,
		Message: message,
	})
}

// writeJSON 写入 JSON 响应并设置 Content-Type。
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write json response: %v", err)
	}
}
