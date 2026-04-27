package tasks_transport_http

import (
	"net/http"
	"time"

	"github.com/Kosench/golang-todoapp/internal/core/domain"
	core_logger "github.com/Kosench/golang-todoapp/internal/core/logger"
	core_http_request "github.com/Kosench/golang-todoapp/internal/core/transport/http/request"
	core_http_response "github.com/Kosench/golang-todoapp/internal/core/transport/http/response"
)

type CreateTaskRequest struct {
	Title        string  `json:"title" validate:"required,min=1,max=100"`
	Description  *string `json:"description" validate:"omitempty,min=1,max=1000"`
	AuthorUserID int     `json:"author_user_id" validate:"required"`
}

type CreateTaskResponse TaskDTOResponse

// CreateTask godoc
// @Summary Create a new task
// @Description Creates a new task with the provided title, optional description, and author user ID
// @Tags tasks
// @Accept json
// @Produce json
// @Param request body CreateTaskRequest true "Task creation data"
// @Success 201 {object} CreateTaskResponse "Task created successfully"
// @Failure 400 {object} core_http_response.ErrorResponse "Invalid request body or validation error"
// @Failure 500 {object} core_http_response.ErrorResponse "Internal server error"
// @Router /tasks [post]
func (h *TasksHTTPHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, w)

	var req CreateTaskRequest
	if err := core_http_request.DecodeAndValidator(r, &req); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode and validate HTTP request")
		return
	}

	taskDomain := domain.NewTaskUninitialized(
		req.Title,
		req.Description,
		req.AuthorUserID,
		time.Now(),
	)

	taskDomain, err := h.tasksService.CreateTask(ctx, taskDomain)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to create task")
		return
	}

	response := CreateTaskResponse(taskDTOFromDomain(taskDomain))

	responseHandler.JSONResponse(response, http.StatusCreated)

}
