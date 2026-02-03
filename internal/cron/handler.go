package cron

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// APIResponse represents the standard API response format
type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Handler handles cron-related HTTP requests
type Handler struct {
	manager *CronManager
}

// NewHandler creates a new cron handler
func NewHandler(manager *CronManager) *Handler {
	return &Handler{
		manager: manager,
	}
}

// RegisterRoutes registers the cron API routes
func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/cron/tasks", h.ListTasks).Methods("GET")
	router.HandleFunc("/api/cron/tasks", h.CreateTask).Methods("POST")
	router.HandleFunc("/api/cron/tasks/{id}", h.GetTask).Methods("GET")
	router.HandleFunc("/api/cron/tasks/{id}", h.UpdateTask).Methods("PUT")
	router.HandleFunc("/api/cron/tasks/{id}", h.DeleteTask).Methods("DELETE")
	router.HandleFunc("/api/cron/tasks/{id}/execute", h.ExecuteTaskNow).Methods("POST")
}

// ListTasks returns all scheduled tasks
func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	tasks := h.manager.ListTasks()

	response := APIResponse{
		Status: "ok",
		Data:   tasks,
	}

	h.writeJSON(w, response, http.StatusOK)
}

// CreateTask creates a new scheduled task
func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		h.writeJSON(w, APIResponse{
			Status: "error",
			Error:  "Invalid request body",
		}, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if task.Name == "" {
		h.writeJSON(w, APIResponse{
			Status: "error",
			Error:  "Task name is required",
		}, http.StatusBadRequest)
		return
	}

	if task.Schedule == "" {
		h.writeJSON(w, APIResponse{
			Status: "error",
			Error:  "Cron schedule is required",
		}, http.StatusBadRequest)
		return
	}

	if task.Command == "" {
		h.writeJSON(w, APIResponse{
			Status: "error",
			Error:  "Task command is required",
		}, http.StatusBadRequest)
		return
	}

	id, err := h.manager.AddTask(&task)
	if err != nil {
		h.writeJSON(w, APIResponse{
			Status: "error",
			Error:  err.Error(),
		}, http.StatusInternalServerError)
		return
	}

	// Get the created task to return full details
	createdTask, _ := h.manager.GetTask(id)

	response := APIResponse{
		Status: "ok",
		Data:   createdTask,
	}

	h.writeJSON(w, response, http.StatusCreated)
}

// GetTask returns a specific task
func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	task, exists := h.manager.GetTask(taskID)
	if !exists {
		h.writeJSON(w, APIResponse{
			Status: "error",
			Error:  "Task not found",
		}, http.StatusNotFound)
		return
	}

	response := APIResponse{
		Status: "ok",
		Data:   task,
	}

	h.writeJSON(w, response, http.StatusOK)
}

// UpdateTask updates an existing task
func (h *Handler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	var updatedTask Task
	if err := json.NewDecoder(r.Body).Decode(&updatedTask); err != nil {
		h.writeJSON(w, APIResponse{
			Status: "error",
			Error:  "Invalid request body",
		}, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if updatedTask.Name == "" {
		h.writeJSON(w, APIResponse{
			Status: "error",
			Error:  "Task name is required",
		}, http.StatusBadRequest)
		return
	}

	if updatedTask.Schedule == "" {
		h.writeJSON(w, APIResponse{
			Status: "error",
			Error:  "Cron schedule is required",
		}, http.StatusBadRequest)
		return
	}

	if updatedTask.Command == "" {
		h.writeJSON(w, APIResponse{
			Status: "error",
			Error:  "Task command is required",
		}, http.StatusBadRequest)
		return
	}

	err := h.manager.UpdateTask(taskID, &updatedTask)
	if err != nil {
		h.writeJSON(w, APIResponse{
			Status: "error",
			Error:  err.Error(),
		}, http.StatusInternalServerError)
		return
	}

	// Get the updated task to return full details
	updated, _ := h.manager.GetTask(taskID)

	response := APIResponse{
		Status: "ok",
		Data:   updated,
	}

	h.writeJSON(w, response, http.StatusOK)
}

// DeleteTask deletes a scheduled task
func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	err := h.manager.RemoveTask(taskID)
	if err != nil {
		h.writeJSON(w, APIResponse{
			Status: "error",
			Error:  err.Error(),
		}, http.StatusInternalServerError)
		return
	}

	response := APIResponse{
		Status:  "ok",
		Message: "Task deleted successfully",
	}

	h.writeJSON(w, response, http.StatusOK)
}

// ExecuteTaskNow executes a task immediately
func (h *Handler) ExecuteTaskNow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	task, exists := h.manager.GetTask(taskID)
	if !exists {
		h.writeJSON(w, APIResponse{
			Status: "error",
			Error:  "Task not found",
		}, http.StatusNotFound)
		return
	}

	// Execute the task immediately
	go h.manager.executeTask(task)

	response := APIResponse{
		Status:  "ok",
		Message: "Task executed successfully",
		Data: map[string]interface{}{
			"taskId":     taskID,
			"executedAt": time.Now().Format(time.RFC3339),
		},
	}

	h.writeJSON(w, response, http.StatusOK)
}

// writeJSON writes a JSON response
func (h *Handler) writeJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.Encode(data)
}

// TaskRequest represents a task creation/update request
type TaskRequest struct {
	Name        string                 `json:"name"`
	Schedule    string                 `json:"schedule"` // Cron expression
	Command     string                 `json:"command"`  // Command to execute
	Payload     map[string]interface{} `json:"payload"`
	Enabled     *bool                  `json:"enabled,omitempty"`
	Description string                 `json:"description"`
}

// ConvertTaskRequest converts a TaskRequest to a Task
func (h *Handler) ConvertTaskRequest(req *TaskRequest) *Task {
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	return &Task{
		Name:        req.Name,
		Schedule:    req.Schedule,
		Command:     req.Command,
		Payload:     req.Payload,
		Enabled:     enabled,
		Description: req.Description,
	}
}
