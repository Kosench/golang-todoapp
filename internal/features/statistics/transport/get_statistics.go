package statistics_transport_http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Kosench/golang-todoapp/internal/core/domain"
	core_logger "github.com/Kosench/golang-todoapp/internal/core/logger"
	core_http_request "github.com/Kosench/golang-todoapp/internal/core/transport/http/request"
	core_http_response "github.com/Kosench/golang-todoapp/internal/core/transport/http/response"
)

type GetStatisticsResponse struct {
	TasksCreated               int      `json:"tasks_created"`
	TasksCompleted             int      `json:"tasks_completed"`
	TasksCompletedRate         *float64 `json:"tasks_completed_rate"`
	TasksAverageCompletionTime *string  `json:"tasks_average_completion_time"`
}

// GetStatistics godoc
// @Summary Get task statistics
// @Description Retrieves task completion statistics with optional user_id and date range filters
// @Tags statistics
// @Produce json
// @Param user_id query int false "Filter statistics by user ID" minimum(1)
// @Param from query string false "Start date filter (ISO 8601 format)" example(2024-01-01T00:00:00Z)
// @Param to query string false "End date filter (ISO 8601 format)" example(2024-12-31T23:59:59Z)
// @Success 200 {object} GetStatisticsResponse "Statistics retrieved successfully"
// @Failure 400 {object} core_http_response.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} core_http_response.ErrorResponse "Internal server error"
// @Router /statistics [get]
func (h *StatisticsHTTPHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, w)

	userID, from, to, err := getUserIDFromToQueryParams(r)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get 'userID/from/to' query param",
		)
		return
	}

	stats, err := h.statisticsService.GetStatistics(ctx, userID, from, to)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get statistics",
		)
		return
	}

	response := toDTOFromDomain(stats)
	responseHandler.JSONResponse(response, http.StatusOK)
}

func getUserIDFromToQueryParams(r *http.Request) (*int, *time.Time, *time.Time, error) {
	const (
		userIDQueryParamKey = "user_id"
		fromQueryParamKey   = "from"
		toQueryParamKey     = "to"
	)

	userID, err := core_http_request.GetIntQueryParam(r, userIDQueryParamKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get user_id query param: %w", err)
	}

	from, err := core_http_request.GetDateQueryParams(r, fromQueryParamKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get `FROM` query param: %w", err)
	}

	to, err := core_http_request.GetDateQueryParams(r, toQueryParamKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get `TO` query param: %w", err)
	}

	return userID, from, to, nil
}

func toDTOFromDomain(stats domain.Statistics) GetStatisticsResponse {
	var avgTime *string
	if stats.TasksAverageCompletionTime != nil {
		duration := stats.TasksAverageCompletionTime.String()
		avgTime = &duration
	}

	return GetStatisticsResponse{
		TasksCreated:               stats.TasksCreated,
		TasksCompleted:             stats.TasksCompleted,
		TasksCompletedRate:         stats.TasksCompletedRate,
		TasksAverageCompletionTime: avgTime,
	}
}
