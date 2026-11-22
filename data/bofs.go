package data

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/types"
)

const (
	bofsDir     = "bofs"
	windowsDir  = "bofs/windows"
	linuxDir    = "bofs/linux"
	tasksDir    = "tasks"
	metadataExt = ".meta.json"
	taskExt     = ".json"
)

var (
	fileCache sync.Map
)

// BOFInfo represents metadata for a BOF file
type BOFInfo = types.BOFInfo

// TaskStatus represents the status of a BOF task
type TaskStatus = types.TaskStatus

const (
	TaskPending     TaskStatus = types.TaskPending
	TaskExecuting   TaskStatus = types.TaskExecuting
	TaskCompleted   TaskStatus = types.TaskCompleted
	TaskFailed      TaskStatus = types.TaskFailed
)

// TaskInfo represents a BOF execution task
type TaskInfo = types.TaskInfo

// InitializeBOFDir creates the bofs directory structure if it doesn't exist
func InitializeBOFDir() error {
	err := os.MkdirAll(windowsDir, 0755)
	if err != nil && !os.IsExist(err) {
		logger.Logf(logger.Error, "Failed to create BOFs windows directory: %v", err)
		return err
	}

	err = os.MkdirAll(linuxDir, 0755)
	if err != nil && !os.IsExist(err) {
		logger.Logf(logger.Error, "Failed to create BOFs linux directory: %v", err)
		return err
	}

	logger.Logf(logger.Info, "BOFs directory initialized")
	return nil
}

// StoreBOF stores a new BOF file (flat file approach)
func StoreBOF(uuid, platform, functionName, description string, payload []byte) (string, error) {
	if platform != "windows" && platform != "linux" {
		return "", fmt.Errorf("invalid platform: %s", platform)
	}

	var dirPath string
	var fileName string

	switch platform {
	case "windows":
		dirPath = windowsDir
		fileName = functionName + ".dll"
	default:
		dirPath = linuxDir
		fileName = functionName + ".so"
	}

	filePath := filepath.Join(dirPath, fileName)

	if _, err := os.Stat(filePath); err == nil {
		logger.Logf(logger.Warning, "BOF file already exists: %s", filePath)
		return "", fmt.Errorf("BOF file already exists: %s", fileName)
	}

	err := os.WriteFile(filePath, payload, 0644)
	if err != nil {
		logger.Logf(logger.Error, "Failed to write BOF file: %v", err)
		return "", err
	}

	metadata := BOFInfo{
		Uuid:         uuid,
		Platform:     platform,
		FunctionName: functionName,
		Description:  description,
		UploadedAt:   time.Now(),
	}

	metaPath := filePath + metadataExt
	metaJSON, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		logger.Logf(logger.Error, "Failed to marshal BOF metadata: %v", err)
		return "", err
	}

	err = os.WriteFile(metaPath, metaJSON, 0644)
	if err != nil {
		logger.Logf(logger.Error, "Failed to write metadata file: %v", err)
		return "", err
	}

	logger.Logf(logger.Info, "BOF stored successfully: %s (UUID: %s)", fileName, uuid)
	return fileName, nil
}

// DeleteBOF deletes a BOF file and its metadata
func DeleteBOF(functionName string, platform string) error {
	var dirPath string

	switch platform {
	case "windows":
		dirPath = windowsDir
	default:
		dirPath = linuxDir
	}

	fileExt := ".dll"
	if platform == "linux" {
		fileExt = ".so"
	}

	filePath := filepath.Join(dirPath, functionName+fileExt)
	metaPath := filePath + metadataExt

	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		logger.Logf(logger.Error, "Failed to delete BOF file: %v", err)
		return err
	}

	if err := os.Remove(metaPath); err != nil && !os.IsNotExist(err) {
		logger.Logf(logger.Error, "Failed to delete metadata: %v", err)
		return err
	}

	logger.Logf(logger.Info, "BOF deleted successfully: %s/%s", platform, functionName)
	return nil
}

// GetAllBOFs returns all BOFs from the file system (global repository)
func GetAllBOFs() ([]*BOFInfo, error) {
	var bofs []*BOFInfo

	cacheKeys := make([]string, 0)
	fileCache.Range(func(key, _ interface{}) bool {
		cacheKeys = append(cacheKeys, key.(string))
		return true
	})

	if len(cacheKeys) > 0 {
		for _, key := range cacheKeys {
			if info, ok := fileCache.Load(key); ok {
				bofs = append(bofs, info.(*BOFInfo))
			}
		}
	} else {
		bofs, err := loadBOFsFromDisk()
		if err != nil {
			return nil, err
		}

		for _, bof := range bofs {
			fileCache.Store(bof.Platform+"_"+bof.FunctionName, bof)
		}
	}

	return bofs, nil
}

// GetBOFsByPlatform returns BOFs filtered by platform
func GetBOFsByPlatform(platform string) ([]*BOFInfo, error) {
	allBofs, err := GetAllBOFs()
	if err != nil {
		return nil, err
	}

	var filtered []*BOFInfo
	for _, bof := range allBofs {
		if bof.Platform == platform {
			filtered = append(filtered, bof)
		}
	}

	return filtered, nil
}

// GetBOF loads a specific BOF by function name and platform
func GetBOF(functionName string, platform string) (*BOFInfo, []byte, error) {
	var dirPath string
	fileExt := ".dll"

	switch platform {
	case "windows":
		dirPath = windowsDir
	default:
		dirPath = linuxDir
		fileExt = ".so"
	}

	filePath := filepath.Join(dirPath, functionName+fileExt)
	metaPath := filePath + metadataExt

	payload, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("BOF file not found: %s", functionName)
	}

	var metadata BOFInfo
	metaJSON, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, nil, fmt.Errorf("metadata file not found: %s", metaPath)
	}

	err = json.Unmarshal(metaJSON, &metadata)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse metadata: %v", err)
	}

	cacheKey := platform + "_" + functionName
	fileCache.Store(cacheKey, &metadata)

	return &metadata, payload, nil
}

// loadBOFsFromDisk scans the file system and loads all BOF metadata
func loadBOFsFromDisk() ([]*BOFInfo, error) {
	var bofs []*BOFInfo

	windowsBofs, err := scanDirectory(windowsDir)
	if err != nil {
		logger.Logf(logger.Warning, "Error scanning Windows BOFs: %v", err)
	} else {
		bofs = append(bofs, windowsBofs...)
	}

	linuxBofs, err := scanDirectory(linuxDir)
	if err != nil {
		logger.Logf(logger.Warning, "Error scanning Linux BOFs: %v", err)
	} else {
		bofs = append(bofs, linuxBofs...)
	}

	return bofs, nil
}

// scanDirectory scans a directory for metadata files and loads BOF info
func scanDirectory(dirPath string) ([]*BOFInfo, error) {
	var bofs []*BOFInfo

	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), metadataExt) {
			metaPath := filepath.Join(dirPath, file.Name())

			var metadata BOFInfo
			metaJSON, err := os.ReadFile(metaPath)
			if err != nil {
				logger.Logf(logger.Warning, "Failed to read metadata %s: %v", file.Name(), err)
				continue
			}

			err = json.Unmarshal(metaJSON, &metadata)
			if err != nil {
				logger.Logf(logger.Warning, "Failed to parse metadata %s: %v", file.Name(), err)
				continue
			}

			bofs = append(bofs, &metadata)
		}
	}

	return bofs, nil
}

// FetchBOFsForAgent returns BOFs associated with a specific agent (for legacy UI display)
func FetchBOFsForAgent(agentID string) ([]*BOFInfo, error) {
	allBofs, err := GetAllBOFs()
	if err != nil {
		return nil, err
	}

	var filtered []*BOFInfo
	for _, bof := range allBofs {
		if strings.Contains(bof.Uuid, agentID) || bof.Uuid == agentID {
			filtered = append(filtered, bof)
		}
	}

	return filtered, nil
}

// GetBOFByUuid returns a BOF by its UUID and platform
func GetBOFByUuid(uuid string, platform string) (*BOFInfo, []byte, error) {
	allBofs, err := GetAllBOFs()
	if err != nil {
		return nil, nil, err
	}
	
	for _, bof := range allBofs {
		if bof.Uuid == uuid {
			// Return the BOF - platform should match what was stored
			metadata, payload, err := GetBOF(bof.FunctionName, bof.Platform)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to load BOF file: %w", err)
			}
			return metadata, payload, nil
		}
	}
	
	return nil, nil, fmt.Errorf("BOF with UUID %s not found", uuid)
}

// ListAllBOFs returns all BOFs from the global repository (not assigned to any agent)
func ListAllBOFs() ([]*BOFInfo, error) {
	bofs, err := GetAllBOFs()
	if err != nil {
		return nil, err
	}
	return bofs, nil
}

// CreateTask creates a new BOF execution task
func CreateTask(agentUuid, bofUuid string, payload []byte) (*TaskInfo, error) {
	task := TaskInfo{
		Id:          fmt.Sprintf("%d", time.Now().UnixNano()),
		BofUuid:     bofUuid,
		AgentUuid:   agentUuid,
		Status:      TaskPending,
		CreatedAt:   time.Now(),
	}

	// Save task to file
	taskPath := filepath.Join(tasksDir, fmt.Sprintf("%s%s", agentUuid, taskExt))
	taskJSON, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal task: %w", err)
	}

	if err := os.WriteFile(taskPath, taskJSON, 0644); err != nil {
		return nil, fmt.Errorf("failed to write task file: %w", err)
	}

	// Also store payload for later retrieval
	payloadPath := filepath.Join(tasksDir, fmt.Sprintf("%s_payload%s", agentUuid, taskExt))
	if err := os.WriteFile(payloadPath, payload, 0644); err != nil {
		return nil, fmt.Errorf("failed to write payload file: %w", err)
	}

	logger.Logf(logger.Info, "Created task %s for agent %s, BOF %s", task.Id, agentUuid, bofUuid)
	return &task, nil
}

// GetPendingTasks returns pending tasks for a specific agent
func GetPendingTasks(agentUuid string) ([]*TaskInfo, error) {
	taskPath := filepath.Join(tasksDir, fmt.Sprintf("%s%s", agentUuid, taskExt))
	
	_, err := os.Stat(taskPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []*TaskInfo{}, nil
		}
		return nil, fmt.Errorf("failed to stat task file: %w", err)
	}

	var tasks []*TaskInfo
	taskJSON, err := os.ReadFile(taskPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read task file: %w", err)
	}

	err = json.Unmarshal(taskJSON, &tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to parse task JSON: %w", err)
	}

	// Filter only pending tasks
	var pending []*TaskInfo
	for _, task := range tasks {
		if task.Status == TaskPending {
			pending = append(pending, task)
		}
	}

	return pending, nil
}

// UpdateTaskStatus updates the status of a task
func MarkTaskCompleted(taskId, agentUuid, status string, output string) error {
	// Read existing tasks from file
	tasksFile := filepath.Join(TasksDir(), fmt.Sprintf("%s%s", agentUuid, TaskExt()))
	existingTasks, err := readTasksFromFile(tasksFile)
	if err != nil {
		return err
	}

	// Find and update the task
	var updatedTasks []*TaskInfo
	for _, task := range existingTasks {
		if task.Id == taskId && task.AgentUuid == agentUuid {
			task.Status = TaskStatus(status)
			task.Output = output
			now := time.Now()
			task.CompletedAt = &now
		}
		updatedTasks = append(updatedTasks, task)
	}

	// Write updated tasks back to file
	return writeTasksToFile(tasksFile, updatedTasks)
}

func UpdateTaskStatus(taskId string, status TaskStatus, output string, errStr string) error {
	taskPath := filepath.Join(tasksDir, fmt.Sprintf("%s%s", taskId, taskExt))
	_, err := os.Stat(taskPath)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	// Read current task
	var tasks []*TaskInfo
	taskJSON, err := os.ReadFile(taskPath)
	if err != nil {
		return fmt.Errorf("failed to read task file: %w", err)
	}

	err = json.Unmarshal(taskJSON, &tasks)
	if err != nil {
		return fmt.Errorf("failed to parse task JSON: %w", err)
	}

	// Update the task
	for _, task := range tasks {
		if task.Id == taskId {
			task.Status = status
			task.Output = output
			if errStr != "" {
				task.Error = errStr
			}
			if status == TaskCompleted || status == TaskFailed {
				now := time.Now()
				task.CompletedAt = &now
			}
			break
		}
	}

	// Write back
	taskJSON, err = json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	if err := os.WriteFile(taskPath, taskJSON, 0644); err != nil {
		return fmt.Errorf("failed to write task file: %w", err)
	}

	logger.Logf(logger.Info, "Updated task %s status to %s", taskId, status)
	return nil
}

// GetTaskPayload retrieves the payload for a specific task
func GetTaskPayload(taskId string) ([]byte, error) {
	payloadPath := filepath.Join(tasksDir, fmt.Sprintf("%s_payload%s", taskId, taskExt))
	return os.ReadFile(payloadPath)
}

// ClearCache clears the file cache
func ClearCache() {
	fileCache.Range(func(key, value interface{}) bool {
		fileCache.Delete(key)
		return true
	})
	logger.Logf(logger.Info, "BOF cache cleared")
}

// InitializeTasksDir creates the tasks directory if it doesn't exist
func InitializeTasksDir() error {
	err := os.MkdirAll(tasksDir, 0755)
	if err != nil && !os.IsExist(err) {
		logger.Logf(logger.Error, "Failed to create tasks directory: %v", err)
		return err
	}
	return nil
}

// TasksDir returns the tasks directory path
func TasksDir() string {
	return tasksDir
}

// TaskExt returns the task file extension
func TaskExt() string {
	return taskExt
}

// readTasksFromFile reads all tasks from a single task file
func readTasksFromFile(filePath string) ([]*TaskInfo, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read task file: %w", err)
	}

	var tasks []*TaskInfo
	err = json.Unmarshal(data, &tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to parse task JSON: %w", err)
	}

	return tasks, nil
}

// writeTasksToFile writes all tasks to a single task file
func writeTasksToFile(filePath string, tasks []*TaskInfo) error {
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write task file: %w", err)
	}

	return nil
}
