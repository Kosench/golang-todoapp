package core_http_response

import "net/http"

var StatusCodeUninitialized = -1

type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		statusCode:     StatusCodeUninitialized,
	}
}

func (r *ResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.statusCode = statusCode
}

func (r *ResponseWriter) GetStatusCode() int {
	if r.statusCode == StatusCodeUninitialized {
		return http.StatusOK
	}

	return r.statusCode
}
