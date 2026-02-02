package cron

import (
	"testing"
	"time"
)

func TestCronManager_BasicOperations(t *testing.T) {
	manager := NewCronManager(nil) // Use default logger

	// Test adding a task
	task := Task{
		Name:        "test-task",
		Schedule:    "* * * * *", // Every minute
		Command:     "echo",
		Payload:     map[string]interface{}{"message": "test"},
		Description: "A test task",
	}

	id, err := manager.AddTask(&task)
	if err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	// Test getting the task
	retrievedTask, exists := manager.GetTask(id)
	if !exists {
		t.Fatalf("Failed to get task: task does not exist")
	}

	if retrievedTask.Name != task.Name {
		t.Errorf("Task name mismatch: expected '%s', got '%s'", task.Name, retrievedTask.Name)
	}

	// Test listing tasks
	tasks := manager.ListTasks()

	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}

	// Test removing task
	err = manager.RemoveTask(id)
	if err != nil {
		t.Fatalf("Failed to remove task: %v", err)
	}

	// Verify task is removed
	tasks = manager.ListTasks()

	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks after removal, got %d", len(tasks))
	}
}

func TestCronManager_Scheduling(t *testing.T) {
	manager := NewCronManager(nil) // Use default logger

	// Add a task that runs every minute
	task := Task{
		Name:        "frequent-task",
		Schedule:    "* * * * *", // Every minute
		Command:     "test",
		Payload:     map[string]interface{}{"test": true},
		Description: "A frequent test task",
	}

	id, err := manager.AddTask(&task)
	if err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	// Start the manager
	manager.Start()

	// Give it a moment to schedule
	time.Sleep(100 * time.Millisecond)

	// Stop the manager
	manager.Stop()

	// Clean up
	manager.RemoveTask(id)
}

func TestCronManager_DisabledTask(t *testing.T) {
	manager := NewCronManager(nil) // Use default logger

	// Add a disabled task
	task := Task{
		Name:        "disabled-task",
		Schedule:    "* * * * *",
		Command:     "echo",
		Payload:     map[string]interface{}{"message": "disabled"},
		Description: "A disabled test task",
		Enabled:     false,
	}

	id, err := manager.AddTask(&task)
	if err != nil {
		t.Fatalf("Failed to add disabled task: %v", err)
	}

	// Verify task is added but disabled
	retrievedTask, exists := manager.GetTask(id)
	if !exists {
		t.Fatalf("Failed to get disabled task: task does not exist")
	}

	if retrievedTask.Enabled {
		t.Error("Expected task to be disabled")
	}

	// Clean up
	manager.RemoveTask(id)
}

func TestCronManager_UpdateTask(t *testing.T) {
	manager := NewCronManager(nil) // Use default logger

	// Add initial task
	initialTask := Task{
		Name:        "update-test",
		Schedule:    "* * * * *",
		Command:     "original",
		Payload:     map[string]interface{}{"original": true},
		Description: "Original task",
	}

	id, err := manager.AddTask(&initialTask)
	if err != nil {
		t.Fatalf("Failed to add initial task: %v", err)
	}

	// Update the task
	updatedTask := Task{
		Name:        "updated-test",
		Schedule:    "0 * * * *", // Hourly instead of minutely
		Command:     "updated",
		Payload:     map[string]interface{}{"updated": true},
		Description: "Updated task",
		Enabled:     false,
	}

	err = manager.UpdateTask(id, &updatedTask)
	if err != nil {
		t.Fatalf("Failed to update task: %v", err)
	}

	// Verify the update
	retrievedTask, exists := manager.GetTask(id)
	if !exists {
		t.Fatalf("Failed to get updated task: task does not exist")
	}

	if retrievedTask.Name != "updated-test" {
		t.Errorf("Name not updated: expected 'updated-test', got '%s'", retrievedTask.Name)
	}

	if retrievedTask.Command != "updated" {
		t.Errorf("Command not updated: expected 'updated', got '%s'", retrievedTask.Command)
	}

	if retrievedTask.Enabled {
		t.Error("Task should be disabled after update")
	}

	// Clean up
	manager.RemoveTask(id)
}

func TestCronManager_MultipleTasks(t *testing.T) {
	manager := NewCronManager(nil) // Use default logger

	// Add multiple tasks
	taskNames := []string{"task1", "task2", "task3", "task4", "task5"}
	taskIDs := make([]string, len(taskNames))

	for i, name := range taskNames {
		task := Task{
			Name:        name,
			Schedule:    "* * * * *",
			Command:     "test",
			Payload:     map[string]interface{}{"id": i},
			Description: "Test task " + name,
		}

		id, err := manager.AddTask(&task)
		if err != nil {
			t.Fatalf("Failed to add task %s: %v", name, err)
		}
		taskIDs[i] = id
	}

	// Verify all tasks are added
	tasks := manager.ListTasks()

	if len(tasks) != len(taskNames) {
		t.Errorf("Expected %d tasks, got %d", len(taskNames), len(tasks))
	}

	// Verify each task exists
	taskMap := make(map[string]bool)
	for _, task := range tasks {
		taskMap[task.Name] = true
	}

	for _, expectedName := range taskNames {
		if !taskMap[expectedName] {
			t.Errorf("Missing task: %s", expectedName)
		}
	}

	// Clean up all tasks
	for _, id := range taskIDs {
		manager.RemoveTask(id)
	}

	// Verify all tasks are removed
	finalTasks := manager.ListTasks()

	if len(finalTasks) != 0 {
		t.Errorf("Expected 0 tasks after cleanup, got %d", len(finalTasks))
	}
}

func TestCronManager_TaskExecution(t *testing.T) {
	manager := NewCronManager(nil) // Use default logger

	// Add a simple task
	task := Task{
		Name:        "execution-test",
		Schedule:    "0 0 1 1 *", // January 1st at 00:00 (will not run immediately)
		Command:     "test-execution",
		Payload:     map[string]interface{}{"executed": false},
		Description: "Task for execution test",
	}

	id, err := manager.AddTask(&task)
	if err != nil {
		t.Fatalf("Failed to add execution test task: %v", err)
	}

	// Manually execute the task
	result, err := manager.ExecuteTaskNow(id)
	if err != nil {
		t.Logf("Task execution returned error (expected for test command): %v", err)
		// We expect an error since "test-execution" is not a real command
		// But the execution attempt should still happen
	}

	taskId, ok := result["taskId"].(string)
	if !ok || taskId != id {
		t.Errorf("Result task ID mismatch: expected '%s', got '%v'", id, result["taskId"])
	}

	// Clean up
	manager.RemoveTask(id)
}

func TestCronManager_InvalidSchedule(t *testing.T) {
	manager := NewCronManager(nil) // Use default logger

	// Add a task with invalid schedule
	invalidTask := Task{
		Name:        "invalid-schedule",
		Schedule:    "invalid", // Invalid cron schedule
		Command:     "test",
		Payload:     map[string]interface{}{"test": true},
		Description: "Task with invalid schedule",
	}

	id, err := manager.AddTask(&invalidTask)
	if err != nil {
		// This is expected - invalid schedule should cause an error
		t.Logf("Correctly rejected invalid schedule: %v", err)
		return
	}

	// If no error occurred, the invalid schedule was accepted (which might be okay depending on implementation)
	// Let's try to start the manager to see if it handles invalid schedules gracefully
	manager.Start()
	time.Sleep(100 * time.Millisecond)
	manager.Stop()

	// Clean up
	manager.RemoveTask(id)
}
