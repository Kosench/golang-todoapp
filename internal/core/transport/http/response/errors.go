package core_http_response

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
