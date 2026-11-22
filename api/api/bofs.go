package api

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/PatronC2/Patron/data"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/gin-gonic/gin"
)

// getTaskString converts a *string to string, returning empty string if nil
func getTaskString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

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
	// Handle multipart form data (file upload)
	err := c.Request.ParseMultipartForm(32 << 20) // 32 MB max memory
	if err != nil {
		logger.Logf(logger.Error, "Failed to parse multipart form: %v", err)
		c.JSON(http.StatusBadRequest, BOFResponse{
			Success: false,
			Error:   "Invalid request format",
		})
		return
	}

	// Get file from form
	file, handlerErr := c.FormFile("payload")
	if handlerErr != nil {
		logger.Logf(logger.Error, "Failed to get uploaded file: %v", handlerErr)
		c.JSON(http.StatusBadRequest, BOFResponse{
			Success: false,
			Error:   "BOF file is required",
		})
		return
	}

	// Get form fields
	uuid := c.PostForm("uuid")
	platform := c.PostForm("platform")
	functionName := c.PostForm("function_name")
	description := c.PostForm("description")

	if uuid == "" {
		c.JSON(http.StatusBadRequest, BOFResponse{
			Success: false,
			Error:   "BOF UUID is required",
		})
		return
	}

	if platform != "windows" && platform != "linux" {
		c.JSON(http.StatusBadRequest, BOFResponse{
			Success: false,
			Error:   "Invalid platform. Must be 'windows' or 'linux'",
		})
		return
	}

	if functionName == "" {
		c.JSON(http.StatusBadRequest, BOFResponse{
			Success: false,
			Error:   "Function name is required",
		})
		return
	}

	// Open the uploaded file to get its contents
	uploadedFile, err := file.Open()
	if err != nil {
		logger.Logf(logger.Error, "Failed to open uploaded file: %v", err)
		c.JSON(http.StatusInternalServerError, BOFResponse{
			Success: false,
			Error:   "Failed to process uploaded file",
		})
		return
	}
	defer uploadedFile.Close()

	// Read file contents
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(uploadedFile)
	if err != nil {
		logger.Logf(logger.Error, "Failed to read uploaded file: %v", err)
		c.JSON(http.StatusInternalServerError, BOFResponse{
			Success: false,
			Error:   "Failed to process uploaded file",
		})
		return
	}

	// Store the BOF
	fileName, err := data.StoreBOF(uuid, platform, functionName, description, buf.Bytes())
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

// DeleteBOFHandler handles DELETE /api/bofs/:fileid - Delete a BOF file by UUID
func DeleteBOFHandler(c *gin.Context) {
	uuid := c.Param("fileid")

	// First, try to get the BOF from the database to determine platform
	metadata, _, err := data.GetBOFByUuid(uuid, "windows")
	if err != nil {
		// Try linux if windows fails
		metadata, _, err = data.GetBOFByUuid(uuid, "linux")
	}

	if err != nil {
		c.JSON(http.StatusNotFound, BOFResponse{
			Success: false,
			Error:   fmt.Sprintf("BOF not found in database: %s", uuid),
		})
		return
	}

	// Delete using the UUID from metadata
	if err := data.DeleteBOF(metadata.Uuid, metadata.Platform); err != nil {
		c.JSON(http.StatusInternalServerError, BOFResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to delete BOF: %v", err),
		})
		return
	}

	// Also remove from database
	if err := data.DeleteBOFFromDatabase(uuid); err != nil {
		logger.Logf(logger.Error, "Failed to delete BOF from database: %v", err)
		// Don't fail the entire operation if DB delete fails
	}

	c.JSON(http.StatusOK, BOFResponse{
		Success: true,
		Message: fmt.Sprintf("BOF deleted successfully: UUID=%s", uuid),
	})
}

// CreateTaskHandler handles POST /api/bofs/task - Create a BOF execution task
func CreateTaskHandler(c *gin.Context) {
	var req struct {
		AgentUuid  string `json:"agent_uuid"`
		BofUuid    string `json:"bof_uuid"`
		FunctionName string `json:"function_name,omitempty"` // Optional: override function name
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

	// Determine function name (use provided or default to function_name from metadata)
	functionName := req.FunctionName
	if functionName == "" {
		// Try to get the function name from BOF metadata
		metadata, _, metaErr := data.GetBOFByUuid(req.BofUuid, "windows")
		if metaErr != nil {
			metadata, _, metaErr = data.GetBOFByUuid(req.BofUuid, "linux")
		}
		if metadata != nil && metadata.FunctionName != "" {
			functionName = metadata.FunctionName
		} else if payload != nil && len(payload) > 0 {
			// Default to 'go' for BOFs without explicit function name
			functionName = "go"
		}
	}

	logger.Logf(logger.Info, "Creating task %s for agent %s, BOF %s, function=%s", 
		req.AgentUuid, req.BofUuid, functionName)

	// Create task with function name included
	task, err := data.CreateTaskWithFunction(req.AgentUuid, req.BofUuid, payload, functionName)
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

	// Get completed/failed tasks from database too
	completedTasks, err := data.GetCompletedTasks(agentUuid)
	if err != nil {
		logger.Logf(logger.Error, "Failed to get completed tasks for agent %s: %v", agentUuid, err)
		// Don't fail the entire response if completed tasks query fails
	}

	response := make([]map[string]interface{}, len(pendingTasks)+len(completedTasks))
	for i, task := range pendingTasks {
		response[i] = map[string]interface{}{
			"id":          task.Id,
			"bof_uuid":    task.BofUuid,
			"agent_uuid":  task.AgentUuid,
			"status":      string(task.Status),
			"output":      getTaskString(task.Output),
			"error":       getTaskString(task.Error),
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
			"output":      getTaskString(task.Output),
			"error":       getTaskString(task.Error),
			"created_at":  task.CreatedAt.Format(time.RFC3339),
			"completed_at": fmt.Sprintf("%v", task.CompletedAt),
		}
	}

	c.JSON(http.StatusOK, response)
}
