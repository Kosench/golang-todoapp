package tasks_transport_http

import (
	"fmt"
	"net/http"

	core_errors "github.com/Kosench/golang-todoapp/internal/core/errors"
	core_logger "github.com/Kosench/golang-todoapp/internal/core/logger"
	core_http_request "github.com/Kosench/golang-todoapp/internal/core/transport/http/request"
	core_http_response "github.com/Kosench/golang-todoapp/internal/core/transport/http/response"
)

type GetTasksResponse []TaskDTOResponse

// GetTasks godoc
// @Summary Get all tasks
// @Description Retrieves a paginated list of tasks with optional user_id filter
// @Tags tasks
// @Produce json
// @Param user_id query int false "Filter tasks by user ID" minimum(1)
// @Param limit query int false "Maximum number of tasks to return" minimum(0)
// @Param offset query int false "Number of tasks to skip" minimum(0)
// @Success 200 {array} TaskDTOResponse "List of tasks"
// @Failure 400 {object} core_http_response.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} core_http_response.ErrorResponse "Internal server error"
// @Router /tasks [get]
func (h *TasksHTTPHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, w)

	userID, limit, offset, err := getUserIDLimitOffsetQueryParams(r)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get 'userID/limit/offset' query param",
		)

		return
	}

	userDomain, err := h.tasksService.GetTasks(ctx, userID, limit, offset)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get tasks",
		)
		return
	}

	response := taskDTOsFromDomains(userDomain)

	responseHandler.JSONResponse(response, http.StatusOK)

}

func getUserIDLimitOffsetQueryParams(r *http.Request) (*int, *int, *int, error) {
	const (
		userIDQueryParamKey = "user_id"
		limitQueryParamKey  = "limit"
		offsetQueryParamKey = "offset"
	)

	userID, err := core_http_request.GetIntQueryParam(r, userIDQueryParamKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get user_id query param: %w", err)
	}

	limit, err := core_http_request.GetIntQueryParam(r, limitQueryParamKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get limit query param: %w", err)
	}
	if limit != nil && *limit < 0 {
		return nil, nil, nil, fmt.Errorf("limit must be non-negative: %w", core_errors.ErrInvalidArgument)
	}

	offset, err := core_http_request.GetIntQueryParam(r, offsetQueryParamKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get offset query param: %w", err)
	}
	if offset != nil && *offset < 0 {
		return nil, nil, nil, fmt.Errorf("offset must be non-negative: %w", core_errors.ErrInvalidArgument)
	}

	return userID, limit, offset, nil
}
