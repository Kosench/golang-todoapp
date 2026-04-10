package users_transport_http

import (
	"net/http"

	core_logger "github.com/Kosench/golang-todoapp/internal/core/logger"
	core_http_request "github.com/Kosench/golang-todoapp/internal/core/transport/http/request"
	core_http_response "github.com/Kosench/golang-todoapp/internal/core/transport/http/response"
)

type GetUserResponse UserDTOResponse

// GetUser godoc
// @Summary Get a user by ID
// @Description Retrieves a single user by their ID
// @Tags users
// @Produce json
// @Param id path int true "User ID" minimum(1)
// @Success 200 {object} GetUserResponse "User found"
// @Failure 400 {object} core_http_response.ErrorResponse "Invalid user ID"
// @Failure 404 {object} core_http_response.ErrorResponse "User not found"
// @Failure 500 {object} core_http_response.ErrorResponse "Internal server error"
// @Router /users/{id} [get]
func (h *UsersHTTPHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, w)

	id, err := core_http_request.GetIntPathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed do get user id path value",
		)
		return
	}

	user, err := h.userService.GetUser(ctx, id)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get user")
		return
	}

	response := GetUserResponse(userDTOFromDomain(user))
	responseHandler.JSONResponse(response, http.StatusOK)
}
