package users_transport_http

import (
	"fmt"
	"net/http"

	core_errors "github.com/Kosench/golang-todoapp/internal/core/errors"
	core_logger "github.com/Kosench/golang-todoapp/internal/core/logger"
	core_http_request "github.com/Kosench/golang-todoapp/internal/core/transport/http/request"
	core_http_response "github.com/Kosench/golang-todoapp/internal/core/transport/http/response"
)

type GetUsersResponse []UserDTOResponse

// GetUsers godoc
// @Summary Get all users
// @Description Retrieves a paginated list of users with optional limit and offset
// @Tags users
// @Produce json
// @Param limit query int false "Maximum number of users to return" minimum(0)
// @Param offset query int false "Number of users to skip" minimum(0)
// @Success 200 {array} UserDTOResponse "List of users"
// @Failure 400 {object} core_http_response.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} core_http_response.ErrorResponse "Internal server error"
// @Router /users [get]
func (h *UsersHTTPHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, w)

	limit, offset, err := getLimitOffsetQueryParams(r)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get 'limit/offset' query param",
		)

		return
	}

	userDomains, err := h.userService.GetUsers(ctx, limit, offset)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get users",
		)
		return
	}

	response := GetUsersResponse(usersDTOFromDomains(userDomains))

	responseHandler.JSONResponse(response, 200)
}

func getLimitOffsetQueryParams(r *http.Request) (*int, *int, error) {
	const (
		limitQueryParamKey  = "limit"
		offsetQueryParamKey = "offset"
	)

	limit, err := core_http_request.GetIntQueryParam(r, limitQueryParamKey)
	if err != nil {
		return nil, nil, fmt.Errorf("get limit query param: %w", err)
	}
	if limit != nil && *limit < 0 {
		return nil, nil, fmt.Errorf("limit must be non-negative: %w", core_errors.ErrInvalidArgument)
	}

	offset, err := core_http_request.GetIntQueryParam(r, offsetQueryParamKey)
	if err != nil {
		return nil, nil, fmt.Errorf("get offset query param: %w", err)
	}
	if offset != nil && *offset < 0 {
		return nil, nil, fmt.Errorf("offset must be non-negative: %w", core_errors.ErrInvalidArgument)
	}

	return limit, offset, nil
}
