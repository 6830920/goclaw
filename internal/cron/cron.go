package cron

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// Task represents a scheduled task
type Task struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Schedule    string                 `json:"schedule"` // Cron expression
	Command     string                 `json:"command"`  // Command to execute
	Payload     map[string]interface{} `json:"payload"`
	CreatedAt   time.Time              `json:"createdAt"`
	LastRun     *time.Time             `json:"lastRun,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Enabled     bool                   `json:"enabled"`
	Description string                 `json:"description"`
}

// CronManager manages scheduled tasks
type CronManager struct {
	cron      *cron.Cron
	tasks     map[string]*Task
	taskMutex sync.RWMutex
	logger    *log.Logger
}

// NewCronManager creates a new cron manager
func NewCronManager(logger *log.Logger) *CronManager {
	if logger == nil {
		logger = log.Default()
	}

	cm := &CronManager{
		cron:   cron.New(cron.WithChain(cron.Recover(cron.DefaultLogger))),
		tasks:  make(map[string]*Task),
		logger: logger,
	}

	return cm
}

// AddTask adds a new scheduled task
func (cm *CronManager) AddTask(task *Task) (string, error) {
	cm.taskMutex.Lock()
	defer cm.taskMutex.Unlock()

	if task.ID == "" {
		task.ID = fmt.Sprintf("task_%d", time.Now().UnixNano()) // Use Nano to ensure uniqueness
	}

	if _, exists := cm.tasks[task.ID]; exists {
		return "", fmt.Errorf("task with ID %s already exists", task.ID)
	}

	// Only schedule the task if it's enabled
	if task.Enabled {
		_, err := cm.cron.AddFunc(task.Schedule, func() {
			cm.executeTask(task)
		})
		if err != nil {
			return "", fmt.Errorf("failed to schedule task: %w", err)
		}
	}

	task.CreatedAt = time.Now()
	cm.tasks[task.ID] = task

	status := "scheduled"
	if !task.Enabled {
		status = "added (not scheduled - disabled)"
	}
	cm.logger.Printf("Task %s: %s (cron: %s) - %s", task.ID, task.Name, task.Schedule, status)

	return task.ID, nil // Return the actual task ID
}

// RemoveTask removes a scheduled task
func (cm *CronManager) RemoveTask(taskID string) error {
	cm.taskMutex.Lock()
	defer cm.taskMutex.Unlock()

	task, exists := cm.tasks[taskID]
	if !exists {
		return fmt.Errorf("task %s not found", taskID)
	}

	// For now, we'll recreate the cron scheduler
	// In a production system, we'd store the cron.EntryID

	delete(cm.tasks, taskID)

	// For now, we'll recreate the cron scheduler
	// In a production system, we'd store the cron.EntryID
	cm.cron.Stop()
	cm.cron = cron.New(cron.WithChain(cron.Recover(cron.DefaultLogger)))

	// Re-add remaining tasks
	for id, t := range cm.tasks {
		if t.Enabled {
			_, err := cm.cron.AddFunc(t.Schedule, func() {
				cm.executeTask(t)
			})
			if err != nil {
				cm.logger.Printf("Failed to reschedule task %s: %v", id, err)
			}
		}
	}

	if cm.cron.Entries() != nil {
		cm.cron.Start()
	}

	cm.logger.Printf("Removed task %s: %s", taskID, task.Name)
	return nil
}

// Start starts the cron scheduler
func (cm *CronManager) Start() {
	cm.cron.Start()
	cm.logger.Println("Cron scheduler started")
}

// Stop stops the cron scheduler
func (cm *CronManager) Stop() context.Context {
	ctx := cm.cron.Stop()
	cm.logger.Println("Cron scheduler stopped")
	return ctx
}

// ListTasks returns all scheduled tasks
func (cm *CronManager) ListTasks() []*Task {
	cm.taskMutex.RLock()
	defer cm.taskMutex.RUnlock()

	tasks := make([]*Task, 0, len(cm.tasks))
	for _, task := range cm.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// GetTask returns a specific task
func (cm *CronManager) GetTask(taskID string) (*Task, bool) {
	cm.taskMutex.RLock()
	defer cm.taskMutex.RUnlock()

	task, exists := cm.tasks[taskID]
	return task, exists
}

// executeTask executes a scheduled task
func (cm *CronManager) executeTask(task *Task) {
	startTime := time.Now()

	cm.logger.Printf("Executing task %s: %s", task.ID, task.Name)

	// Here you would implement the actual task execution logic
	// For now, we'll just log the execution
	result := cm.runTaskCommand(task)

	// Update task status
	cm.taskMutex.Lock()
	if task.LastRun == nil {
		task.LastRun = &startTime
	} else {
		*task.LastRun = startTime
	}

	if result != nil {
		task.Error = result.Error()
	} else {
		task.Error = ""
	}
	cm.taskMutex.Unlock()

	duration := time.Since(startTime)
	cm.logger.Printf("Task %s completed in %v", task.ID, duration)
}

// runTaskCommand executes the actual command for the task
func (cm *CronManager) runTaskCommand(task *Task) error {
	// This is where you'd implement the actual task logic
	// For example:
	// - Send a notification/reminders
	// - Execute API calls
	// - Process data
	// - etc.

	switch task.Command {
	case "reminder":
		return cm.handleReminder(task)
	case "notification":
		return cm.handleNotification(task)
	default:
		return cm.handleGenericTask(task)
	}
}

// handleReminder handles reminder tasks
func (cm *CronManager) handleReminder(task *Task) error {
	message, ok := task.Payload["message"].(string)
	if !ok {
		message = "提醒: 任务已触发"
	}

	user, ok := task.Payload["user"].(string)
	if !ok {
		user = "system"
	}

	cm.logger.Printf("REMINDER for %s: %s", user, message)
	// In a real implementation, this would send the reminder to the user
	return nil
}

// handleNotification handles notification tasks
func (cm *CronManager) handleNotification(task *Task) error {
	title, ok := task.Payload["title"].(string)
	if !ok {
		title = "通知"
	}

	body, ok := task.Payload["body"].(string)
	if !ok {
		body = "您有一个新的通知"
	}

	cm.logger.Printf("NOTIFICATION: %s - %s", title, body)
	// In a real implementation, this would send the notification
	return nil
}

// handleGenericTask handles generic tasks
func (cm *CronManager) handleGenericTask(task *Task) error {
	cm.logger.Printf("Executing generic task: %s", task.Command)
	// Implement generic task execution
	return nil
}

// UpdateTask updates an existing task
func (cm *CronManager) UpdateTask(taskID string, updatedTask *Task) error {
	cm.taskMutex.Lock()
	defer cm.taskMutex.Unlock()

	existingTask, exists := cm.tasks[taskID]
	if !exists {
		return fmt.Errorf("task %s not found", taskID)
	}

	// Update fields
	existingTask.Name = updatedTask.Name
	existingTask.Schedule = updatedTask.Schedule
	existingTask.Command = updatedTask.Command
	existingTask.Payload = updatedTask.Payload
	existingTask.Enabled = updatedTask.Enabled
	existingTask.Description = updatedTask.Description

	// Remove and re-add the task with new schedule
	cm.cron.Stop()
	cm.cron = cron.New(cron.WithChain(cron.Recover(cron.DefaultLogger)))

	// Re-add all tasks
	for id, t := range cm.tasks {
		if t.Enabled {
			_, err := cm.cron.AddFunc(t.Schedule, func() {
				cm.executeTask(t)
			})
			if err != nil {
				cm.logger.Printf("Failed to reschedule task %s: %v", id, err)
			}
		}
	}

	if cm.cron.Entries() != nil {
		cm.cron.Start()
	}

	return nil
}

// ExecuteTaskNow executes a task immediately (outside of the scheduled time)
func (cm *CronManager) ExecuteTaskNow(taskID string) (map[string]interface{}, error) {
	cm.taskMutex.RLock()
	task, exists := cm.tasks[taskID]
	cm.taskMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("task %s not found", taskID)
	}

	// Execute the task directly
	cm.executeTask(task)

	result := map[string]interface{}{
		"executedAt": time.Now(),
		"taskId":     taskID,
		"taskName":   task.Name,
	}

	return result, nil
}
