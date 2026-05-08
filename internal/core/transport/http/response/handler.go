package core_http_response

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Kosench/golang-todoapp/internal/core/domain"
	core_errors "github.com/Kosench/golang-todoapp/internal/core/errors"
	core_logger "github.com/Kosench/golang-todoapp/internal/core/logger"
	"go.uber.org/zap"
)

const (
	errorCodeInvalidArgument = "invalid_argument"
	errorCodeNotFound        = "not_found"
	errorCodeConflict        = "conflict"
	errorCodeInternal        = "internal"
)

type HTTPResponseHandler struct {
	log *core_logger.Logger
	rw  http.ResponseWriter
}

func NewHTTPResponseHandler(
	log *core_logger.Logger,
	rw http.ResponseWriter) *HTTPResponseHandler {
	return &HTTPResponseHandler{
		log: log,
		rw:  rw,
	}
}

func (h *HTTPResponseHandler) JSONResponse(responseBody any, statusCode int) {
	h.rw.Header().Set("Content-Type", "application/json")
	h.rw.WriteHeader(statusCode)

	if err := json.NewEncoder(h.rw).Encode(responseBody); err != nil {
		h.log.Error("write HTTP response", zap.Error(err))
	}
}

func (h *HTTPResponseHandler) NoContentResponse() {
	h.rw.WriteHeader(http.StatusNoContent)
}

func (h *HTTPResponseHandler) HTMLResponse(htmlFile domain.File) {
	h.rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	h.rw.WriteHeader(http.StatusOK)
	if _, err := h.rw.Write(htmlFile.Buffer()); err != nil {
		h.log.Error("write HTML HTTP response", zap.Error(err))
	}
}

func (h *HTTPResponseHandler) ErrorResponse(err error, msg string) {
	var (
		statusCode    int
		code          string
		publicMessage string
		logFunc       func(string, ...zap.Field)
	)

	switch {
	case errors.Is(err, core_errors.ErrInvalidArgument):
		statusCode = http.StatusBadRequest
		code = errorCodeInvalidArgument
		publicMessage = "invalid request"
		logFunc = h.log.Warn

	case errors.Is(err, core_errors.ErrNotFound):
		statusCode = http.StatusNotFound
		code = errorCodeNotFound
		publicMessage = "resource not found"
		logFunc = h.log.Debug

	case errors.Is(err, core_errors.ErrConflict):
		statusCode = http.StatusConflict
		code = errorCodeConflict
		publicMessage = "resource conflict"
		logFunc = h.log.Warn

	default:
		statusCode = http.StatusInternalServerError
		code = errorCodeInternal
		publicMessage = "internal server error"
		logFunc = h.log.Warn
	}

	logFunc(msg, zap.Error(err))
	h.errorResponse(statusCode, code, publicMessage)
}

func (h *HTTPResponseHandler) PanicResponse(p any, msg string) {
	statusCode := http.StatusInternalServerError

	h.log.Error(msg, zap.Any("panic", p))
	h.errorResponse(statusCode, errorCodeInternal, "internal server error")
}

func (h *HTTPResponseHandler) errorResponse(statusCode int, code string, msg string) {
	response := ErrorResponse{
		Code:    code,
		Message: msg,
	}

	h.JSONResponse(response, statusCode)
}
