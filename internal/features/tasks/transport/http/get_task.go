package tasks_transport_http

import (
	"net/http"

	core_logger "github.com/Kosench/golang-todoapp/internal/core/logger"
	core_http_request "github.com/Kosench/golang-todoapp/internal/core/transport/http/request"
	core_http_response "github.com/Kosench/golang-todoapp/internal/core/transport/http/response"
)

type GetTaskResponse TaskDTOResponse

// GetTask godoc
// @Summary Get a task by ID
// @Description Retrieves a single task by its unique identifier
// @Tags tasks
// @Produce json
// @Param id path int true "Task ID" minimum(1)
// @Success 200 {object} GetTaskResponse "Task found"
// @Failure 400 {object} core_http_response.ErrorResponse "Invalid task ID"
// @Failure 404 {object} core_http_response.ErrorResponse "Task not found"
// @Failure 500 {object} core_http_response.ErrorResponse "Internal server error"
// @Router /tasks/{id} [get]
func (h *TasksHTTPHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, w)

	taskID, err := core_http_request.GetIntPathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get taskId path value")
		return
	}

	task, err := h.tasksService.GetTask(ctx, taskID)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get task")
		return
	}

	response := GetTaskResponse(taskDTOFromDomain(task))

	responseHandler.JSONResponse(response, http.StatusOK)
}
