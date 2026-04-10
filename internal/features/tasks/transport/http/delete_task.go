package tasks_transport_http

import (
	"net/http"

	core_logger "github.com/Kosench/golang-todoapp/internal/core/logger"
	core_http_request "github.com/Kosench/golang-todoapp/internal/core/transport/http/request"
	core_http_response "github.com/Kosench/golang-todoapp/internal/core/transport/http/response"
)

// DeleteTask godoc
// @Summary Delete a task
// @Description Deletes a task by its unique identifier
// @Tags tasks
// @Param id path int true "Task ID" minimum(1)
// @Success 204 "Task deleted successfully"
// @Failure 400 {object} core_http_response.ErrorResponse "Invalid task ID"
// @Failure 404 {object} core_http_response.ErrorResponse "Task not found"
// @Failure 500 {object} core_http_response.ErrorResponse "Internal server error"
// @Router /tasks/{id} [delete]
func (h *TasksHTTPHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, w)

	taskID, err := core_http_request.GetIntPathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed do get taskID path value",
		)
		return
	}

	if err = h.tasksService.DeleteTask(ctx, taskID); err != nil {
		responseHandler.ErrorResponse(err, "failed to delete task")
		return
	}

	responseHandler.NoContentResponse()
}
