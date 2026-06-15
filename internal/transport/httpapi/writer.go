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

func writeSessionError(w http.ResponseWriter, err error) {
	if errors.Is(err, session.ErrNotFound) {
		writeError(w, http.StatusNotFound, errorCodeSessionNotFound, "会话不存在")
		return
	}
	writeInternalError(w, err)
}

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

func writeInvalidJSON(w http.ResponseWriter) {
	writeError(w, http.StatusBadRequest, errorCodeInvalidJSON, "请求体不是合法 JSON")
}

func writeRouteNotFound(w http.ResponseWriter) {
	writeError(w, http.StatusNotFound, errorCodeNotFound, "接口不存在")
}

func writeStreamingUnsupported(w http.ResponseWriter) {
	writeError(w, http.StatusInternalServerError, errorCodeStreamingUnsupported, "当前连接不支持流式响应")
}

func writeInternalError(w http.ResponseWriter, err error) {
	log.Printf("internal error: %v", err)
	writeError(w, http.StatusInternalServerError, errorCodeInternalServer, "服务器内部错误")
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, errorResponse{
		Code:    code,
		Message: message,
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write json response: %v", err)
	}
}
