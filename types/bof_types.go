package types

import "time"

// BOFInfo represents metadata for a BOF file
type BOFInfo struct {
	Uuid         string    `json:"uuid"`
	Platform     string    `json:"platform"`
	FunctionName string    `json:"function_name"`
	Description  string    `json:"description"`
	UploadedAt   time.Time `json:"uploaded_at"`
}

// TaskStatus represents the status of a BOF task
type TaskStatus string

const (
	TaskPending     TaskStatus = "pending"
	TaskExecuting   TaskStatus = "executing"
	TaskCompleted   TaskStatus = "completed"
	TaskFailed      TaskStatus = "failed"
)

// TaskInfo represents a BOF execution task
type TaskInfo struct {
	Id          string     `json:"id"`
	BofUuid     string     `json:"bof_uuid"`
	AgentUuid   string     `json:"agent_uuid"`
	Status      TaskStatus `json:"status"`
	Output      string     `json:"output,omitempty"`
	Error       string     `json:"error,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}