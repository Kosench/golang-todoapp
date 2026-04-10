package users_transport_http

import (
	"net/http"

	core_logger "github.com/Kosench/golang-todoapp/internal/core/logger"
	core_http_request "github.com/Kosench/golang-todoapp/internal/core/transport/http/request"
	core_http_response "github.com/Kosench/golang-todoapp/internal/core/transport/http/response"
)

// DeleteUser godoc
// @Summary Delete a user
// @Description Deletes a user by their ID
// @Tags users
// @Produce json
// @Param id path int true "User ID" minimum(1)
// @Success 204 "User deleted successfully"
// @Failure 400 {object} core_http_response.ErrorResponse "Invalid user ID"
// @Failure 404 {object} core_http_response.ErrorResponse "User not found"
// @Failure 500 {object} core_http_response.ErrorResponse "Internal server error"
// @Router /users/{id} [delete]
func (h *UsersHTTPHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, w)

	id, err := core_http_request.GetIntPathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed do get userID path value",
		)
		return
	}

	if err = h.userService.DeleteUser(ctx, id); err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to delete user",
		)
		return
	}

	responseHandler.NoContentResponse()
}
