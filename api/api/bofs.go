package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/PatronC2/Patron/data"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/gin-gonic/gin"
)

// TaskResponse represents API response for task operations
type TaskResponse struct {
	Success bool   `json:"success"`
	TaskID  string `json:"task_id,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// BOFListResponse represents a BOF in the list
type BOFListItem struct {
	FileId       int       `json:"file_id"`
	Uuid         string    `json:"uuid"`
	Platform     string    `json:"platform"`
	FunctionName string    `json:"function_name"`
	Description  string    `json:"description"`
	UploadedAt   time.Time `json:"uploaded_at"`
}

type BOFResponse struct {
	Success bool   `json:"success"`
	ID      string `json:"id,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// UploadBOFHandler handles POST /api/bofs - Upload a new BOF file
func UploadBOFHandler(c *gin.Context) {
	var req struct {
		Uuid          string `json:"uuid"`
		Platform      string `json:"platform"`
		FunctionName  string `json:"function_name"`
		Description   string `json:"description,omitempty"`
		BofPayload    []byte `json:"payload"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Logf(logger.Error, "Failed to parse BOF request: %v", err)
		c.JSON(http.StatusBadRequest, BOFResponse{
			Success: false,
			Error:   "Invalid request format",
		})
		return
	}

	if req.Uuid == "" {
		c.JSON(http.StatusBadRequest, BOFResponse{
			Success: false,
			Error:   "BOF UUID is required",
		})
		return
	}

	if len(req.BofPayload) == 0 {
		c.JSON(http.StatusBadRequest, BOFResponse{
			Success: false,
			Error:   "BOF payload cannot be empty",
		})
		return
	}

	validPlatforms := map[string]bool{"windows": true, "linux": true}
	if !validPlatforms[req.Platform] {
		c.JSON(http.StatusBadRequest, BOFResponse{
			Success: false,
			Error:   "Invalid platform. Must be 'windows' or 'linux'",
		})
		return
	}

	fileName, err := data.StoreBOF(req.Uuid, req.Platform, req.FunctionName, req.Description, req.BofPayload)
	if err != nil {
		logger.Logf(logger.Error, "Failed to store BOF: %v", err)
		c.JSON(http.StatusInternalServerError, BOFResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to store BOF: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, BOFResponse{
		Success: true,
		ID:      fileName,
		Message: "BOF uploaded successfully",
	})
}

// ListBOFsHandler handles GET /api/bofs/:agt - Legacy endpoint (kept for backward compatibility)
func ListBOFsHandler(c *gin.Context) {
	uuid := c.Param("agt")
	if uuid == "" {
		c.JSON(http.StatusBadRequest, BOFResponse{
			Success: false,
			Error:   "Agent UUID is required",
		})
		return
	}

	// Legacy behavior - fetch BOFs associated with agent UUID
	bofs, err := data.FetchBOFsForAgent(uuid)
	if err != nil {
		logger.Logf(logger.Error, "Failed to fetch BOFs for agent %v: %v", uuid, err)
		c.JSON(http.StatusInternalServerError, BOFResponse{
			Success: false,
			Error:   "Failed to fetch BOFs",
		})
		return
	}

	response := make([]map[string]interface{}, len(bofs))
	for i, bof := range bofs {
		response[i] = map[string]interface{}{
			"file_id":      i + 1,
			"uuid":         bof.Uuid,
			"platform":     bof.Platform,
			"function_name": bof.FunctionName,
			"description":  bof.Description,
			"uploaded_at":  bof.UploadedAt.Format(time.RFC3339),
		}
	}

	c.JSON(http.StatusOK, response)
}

// ListAllBOFsHandler handles GET /api/bofs/list - List all BOFs from global repository
func ListAllBOFsHandler(c *gin.Context) {
	bofs, err := data.ListAllBOFs()
	if err != nil {
		logger.Logf(logger.Error, "Failed to fetch all BOFs: %v", err)
		c.JSON(http.StatusInternalServerError, BOFResponse{
			Success: false,
			Error:   "Failed to fetch BOFs",
		})
		return
	}

	response := make([]BOFListItem, len(bofs))
	for i, bof := range bofs {
		response[i] = BOFListItem{
			FileId:       i + 1,
			Uuid:         bof.Uuid,
			Platform:     bof.Platform,
			FunctionName: bof.FunctionName,
			Description:  bof.Description,
			UploadedAt:   bof.UploadedAt,
		}
	}

	c.JSON(http.StatusOK, response)
}

// DeleteBOFHandler handles DELETE /api/bofs/:fileid - Delete a BOF file
func DeleteBOFHandler(c *gin.Context) {
	fileIDStr := c.Param("fileid")

	functionName := fileIDStr

	var deleted bool
	if err := data.DeleteBOF(functionName, "windows"); err == nil {
		deleted = true
	} else if err := data.DeleteBOF(functionName, "linux"); err == nil {
		deleted = true
	} else {
		c.JSON(http.StatusNotFound, BOFResponse{
			Success: false,
			Error:   fmt.Sprintf("BOF not found: %s", functionName),
		})
		return
	}

	if !deleted {
		c.JSON(http.StatusNotFound, BOFResponse{
			Success: false,
			Error:   "BOF not found",
		})
		return
	}

	c.JSON(http.StatusOK, BOFResponse{
		Success: true,
		Message: "BOF deleted successfully",
	})
}

// CreateTaskHandler handles POST /api/bofs/task - Create a BOF execution task
func CreateTaskHandler(c *gin.Context) {
	var req struct {
		AgentUuid string `json:"agent_uuid"`
		BofUuid   string `json:"bof_uuid"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Logf(logger.Error, "Failed to parse task request: %v", err)
		c.JSON(http.StatusBadRequest, TaskResponse{
			Success: false,
			Error:   "Invalid request format",
		})
		return
	}

	if req.AgentUuid == "" {
		c.JSON(http.StatusBadRequest, TaskResponse{
			Success: false,
			Error:   "Agent UUID is required",
		})
		return
	}

	if req.BofUuid == "" {
		c.JSON(http.StatusBadRequest, TaskResponse{
			Success: false,
			Error:   "BOF UUID is required",
		})
		return
	}

	// Get BOF payload from file system (try both platforms)
	_, payload, err := data.GetBOFByUuid(req.BofUuid, "windows")
	if err != nil {
		// Try the other platform if first one fails
		_, payload, err = data.GetBOFByUuid(req.BofUuid, "linux")
		if err != nil {
			logger.Logf(logger.Error, "Failed to get BOF %s: %v", req.BofUuid, err)
			c.JSON(http.StatusNotFound, TaskResponse{
				Success: false,
				Error:   fmt.Sprintf("BOF not found: %s", req.BofUuid),
			})
			return
		}
		
		// BOF should be found by now if we got here
	}

	// Create task
	task, err := data.CreateTask(req.AgentUuid, req.BofUuid, payload)
	if err != nil {
		logger.Logf(logger.Error, "Failed to create task: %v", err)
		c.JSON(http.StatusInternalServerError, TaskResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to create task: %v", err),
		})
		return
	}

	logger.Logf(logger.Info, "Created task %s for agent %s, BOF %s", task.Id, req.AgentUuid, req.BofUuid)
	c.JSON(http.StatusOK, TaskResponse{
		Success: true,
		TaskID:  task.Id,
		Message: fmt.Sprintf("Task created: %s", task.Id),
	})
}

// GetAgentTasksHandler handles GET /api/bofs/tasks/:agentUuid - Get tasks for a specific agent
func GetAgentTasksHandler(c *gin.Context) {
	agentUuid := c.Param("agentUuid")
	if agentUuid == "" {
		c.JSON(http.StatusBadRequest, TaskResponse{
			Success: false,
			Error:   "Agent UUID is required",
		})
		return
	}

	pendingTasks, err := data.GetPendingTasks(agentUuid)
	if err != nil {
		logger.Logf(logger.Error, "Failed to get tasks for agent %s: %v", agentUuid, err)
		c.JSON(http.StatusInternalServerError, TaskResponse{
			Success: false,
			Error:   "Failed to fetch tasks",
		})
		return
	}

	// Get completed tasks too
	completedTasks := make([]*data.TaskInfo, 0)
	taskPath := filepath.Join(data.TasksDir(), fmt.Sprintf("%s%s", agentUuid, data.TaskExt()))
	_, err = os.Stat(taskPath)
	if err == nil {
		var allTasks []*data.TaskInfo
		taskJSON, readErr := os.ReadFile(taskPath)
		if readErr == nil {
			json.Unmarshal(taskJSON, &allTasks)
			for _, task := range allTasks {
				if task.Status != data.TaskPending {
					completedTasks = append(completedTasks, task)
				}
			}
		}
	}

	response := make([]map[string]interface{}, len(pendingTasks)+len(completedTasks))
	for i, task := range pendingTasks {
		response[i] = map[string]interface{}{
			"id":          task.Id,
			"bof_uuid":    task.BofUuid,
			"agent_uuid":  task.AgentUuid,
			"status":      string(task.Status),
			"output":      task.Output,
			"error":       task.Error,
			"created_at":  task.CreatedAt.Format(time.RFC3339),
			"completed_at": fmt.Sprintf("%v", task.CompletedAt),
		}
	}

	for i, task := range completedTasks {
		response[len(pendingTasks)+i] = map[string]interface{}{
			"id":          task.Id,
			"bof_uuid":    task.BofUuid,
			"agent_uuid":  task.AgentUuid,
			"status":      string(task.Status),
			"output":      task.Output,
			"error":       task.Error,
			"created_at":  task.CreatedAt.Format(time.RFC3339),
			"completed_at": fmt.Sprintf("%v", task.CompletedAt),
		}
	}

	c.JSON(http.StatusOK, response)
}
