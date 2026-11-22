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

// StoreBOF stores a new BOF file with UUID-based organization
func StoreBOF(uuid, platform, functionName, description string, payload []byte) (string, error) {
	if platform != "windows" && platform != "linux" {
		return "", fmt.Errorf("invalid platform: %s", platform)
	}

	// Create a unique directory for this BOF using UUID
	uuidDir := filepath.Join(bofsDir, platform, uuid)
	err := os.MkdirAll(uuidDir, 0755)
	if err != nil && !os.IsExist(err) {
		logger.Logf(logger.Error, "Failed to create BOF directory: %v", err)
		return "", fmt.Errorf("failed to create BOF directory: %w", err)
	}

	// Use UUID as the filename for the BOF binary
	fileName := uuid + "." + strings.ToLower(platform) // e.g., "123e4567-e89b-12d3-a456-426614174000.so"
	filePath := filepath.Join(uuidDir, fileName)

	// Check if BOF already exists
	if _, err := os.Stat(filePath); err == nil {
		logger.Logf(logger.Warning, "BOF file already exists: %s", filePath)
		return "", fmt.Errorf("BOF file already exists: %s", fileName)
	}

	err = os.WriteFile(filePath, payload, 0644)
	if err != nil {
		logger.Logf(logger.Error, "Failed to write BOF file: %v", err)
		return "", fmt.Errorf("failed to write BOF file: %w", err)
	}

	// Create metadata
	metadata := BOFInfo{
		Uuid:          uuid,
		Platform:      platform,
		FunctionName:  functionName,
		Description:   description,
		UploadedAt:    time.Now(),
	}

	metaPath := filePath + metadataExt
	metaJSON, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		logger.Logf(logger.Error, "Failed to marshal BOF metadata: %v", err)
		return "", fmt.Errorf("failed to marshal metadata: %w", err)
	}

	err = os.WriteFile(metaPath, metaJSON, 0644)
	if err != nil {
		logger.Logf(logger.Error, "Failed to write metadata file: %v", err)
		return "", fmt.Errorf("failed to write metadata: %w", err)
	}

	// Insert into PostgreSQL database
	insertSQL := `
	INSERT INTO bofs (uuid, platform, function_name, description, uploaded_at)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (uuid) DO UPDATE SET
		platform = EXCLUDED.platform,
		function_name = EXCLUDED.function_name,
		description = EXCLUDED.description,
		uploaded_at = EXCLUDED.uploaded_at
	`
	_, err = db.Exec(insertSQL, uuid, platform, functionName, description, time.Now())
	if err != nil {
		logger.Logf(logger.Error, "Failed to insert BOF into database: %v", err)
		// Clean up files on disk if database insert fails
		os.Remove(filePath)
		os.Remove(metaPath)
		return "", fmt.Errorf("failed to register BOF in database: %w", err)
	}

	logger.Logf(logger.Info, "BOF stored successfully: %s (UUID: %s)", fileName, uuid)
	return fileName, nil
}

// DeleteBOF deletes a BOF file and its metadata by UUID
func DeleteBOF(uuid string, platform string) error {
	// Build path based on UUID (same structure as StoreBOF)
	payloadPath := filepath.Join(platform, uuid, uuid+"."+strings.ToLower(platform))
	metaPath := payloadPath + metadataExt

	if err := os.Remove(payloadPath); err != nil && !os.IsNotExist(err) {
		logger.Logf(logger.Error, "Failed to delete BOF file: %v", err)
		return err
	}

	if err := os.Remove(metaPath); err != nil && !os.IsNotExist(err) {
		logger.Logf(logger.Error, "Failed to delete metadata: %v", err)
		return err
	}

	logger.Logf(logger.Info, "BOF deleted successfully: UUID=%s, Platform=%s", uuid, platform)
	return nil
}

// DeleteBOFFromDatabase removes a BOF entry from the database
func DeleteBOFFromDatabase(uuid string) error {
	_, err := db.Exec(`DELETE FROM bofs WHERE uuid = $1`, uuid)
	if err != nil {
		logger.Logf(logger.Error, "Failed to delete BOF from database: %v", err)
		return fmt.Errorf("failed to delete BOF from database: %w", err)
	}

	logger.Logf(logger.Info, "BOF deleted from database: UUID=%s", uuid)
	return nil
}

// GetAllBOFs returns all BOFs from the file system (global repository)
func GetAllBOFs() ([]*BOFInfo, error) {
	// Query the database for all BOFs
	query := `SELECT uuid, platform, function_name, description, uploaded_at FROM bofs ORDER BY uploaded_at DESC`
	rows, err := db.Query(query)
	if err != nil {
		logger.Logf(logger.Error, "Failed to query BOFs from database: %v", err)
		return []*BOFInfo{}, err
	}
	defer rows.Close()

	var bofs []*BOFInfo
	for rows.Next() {
		var bof BOFInfo
		err := rows.Scan(&bof.Uuid, &bof.Platform, &bof.FunctionName, &bof.Description, &bof.UploadedAt)
		if err != nil {
			logger.Logf(logger.Error, "Failed to scan BOF row: %v", err)
			continue
		}
		bofs = append(bofs, &bof)
	}

	if err = rows.Err(); err != nil {
		logger.Logf(logger.Error, "Error iterating BOF rows: %v", err)
		return bofs, err
	}

	// Store in cache for performance (keyed by UUID for consistency)
	for _, bof := range bofs {
		fileCache.Store(bof.Platform+"_"+bof.Uuid, bof)
	}

	return bofs, nil
}

// GetBOFsByPlatform returns BOFs filtered by platform (database query)
func GetBOFsByPlatform(platform string) ([]*BOFInfo, error) {
	query := `SELECT uuid, platform, function_name, description, uploaded_at FROM bofs WHERE platform = $1 ORDER BY uploaded_at DESC`
	rows, err := db.Query(query, platform)
	if err != nil {
		logger.Logf(logger.Error, "Failed to query BOFs for platform %s: %v", platform, err)
		return []*BOFInfo{}, err
	}
	defer rows.Close()

	var bofs []*BOFInfo
	for rows.Next() {
		var bof BOFInfo
		err := rows.Scan(&bof.Uuid, &bof.Platform, &bof.FunctionName, &bof.Description, &bof.UploadedAt)
		if err != nil {
			logger.Logf(logger.Error, "Failed to scan BOF row: %v", err)
			continue
		}
		bofs = append(bofs, &bof)
	}

	if err = rows.Err(); err != nil {
		logger.Logf(logger.Error, "Error iterating BOF rows: %v", err)
		return bofs, err
	}

	// Store in cache
	for _, bof := range bofs {
		fileCache.Store(bof.Platform+"_"+bof.Uuid, bof)
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
	// Query database for the specific BOF
	query := `SELECT uuid, platform, function_name, description, uploaded_at FROM bofs WHERE uuid = $1`
	rows, err := db.Query(query, uuid)
	if err != nil {
		logger.Logf(logger.Error, "Failed to query BOF from database: %v", err)
		return nil, nil, fmt.Errorf("failed to query BOF from database: %w", err)
	}
	defer rows.Close()

	var bof BOFInfo
	if !rows.Next() {
		// BOF not found in database - check if file exists on disk (orphaned file)
		logger.Logf(logger.Warning, "BOF with UUID %s not found in database, checking file system", uuid)
		return getBOFFromFilesystem(uuid, platform)
	}
	err = rows.Scan(&bof.Uuid, &bof.Platform, &bof.FunctionName, &bof.Description, &bof.UploadedAt)
	if err != nil {
		logger.Logf(logger.Error, "Failed to scan BOF row: %v", err)
		return nil, nil, fmt.Errorf("failed to scan BOF row: %w", err)
	}

	// Verify platform matches
	if bof.Platform != platform {
		return nil, nil, fmt.Errorf("BOF for UUID %s exists only on platform %s, not %s", uuid, bof.Platform, platform)
	}

	// Build file path based on database metadata (UUID-based directory structure)
	bofsPath := filepath.Join(bofsDir, platform, uuid)
	fileName := uuid + "." + strings.ToLower(platform)
	payloadPath := filepath.Join(bofsPath, fileName)
	metaPath := payloadPath + metadataExt

	// Load BOF payload from file
	payload, err := os.ReadFile(payloadPath)
	if err != nil {
		logger.Logf(logger.Error, "Failed to read BOF file %s: %v", payloadPath, err)
		return nil, nil, fmt.Errorf("failed to read BOF file: %w", err)
	}

	// Load metadata from file
	metadataJSON, err := os.ReadFile(metaPath)
	if err != nil {
		logger.Logf(logger.Error, "Failed to read metadata file %s: %v", metaPath, err)
		return nil, nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var loadedMetadata BOFInfo
	err = json.Unmarshal(metadataJSON, &loadedMetadata)
	if err != nil {
		logger.Logf(logger.Error, "Failed to parse metadata: %v", err)
		return nil, nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	// Update cached entry
	fileCache.Store(platform+"_"+uuid, &loadedMetadata)

	return &loadedMetadata, payload, nil
}

// getBOFFromFilesystem attempts to load a BOF directly from the file system
// This is used as a fallback when the BOF exists on disk but not in the database
func getBOFFromFilesystem(uuid string, platform string) (*BOFInfo, []byte, error) {
	bofsPath := filepath.Join(bofsDir, platform, uuid)
	fileName := uuid + "." + strings.ToLower(platform)
	payloadPath := filepath.Join(bofsPath, fileName)
	metaPath := payloadPath + metadataExt

	// Check if files exist on disk
	_, err := os.Stat(payloadPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, fmt.Errorf("BOF file does not exist at %s", payloadPath)
		}
		return nil, nil, fmt.Errorf("failed to stat BOF file: %w", err)
	}

	_, err = os.Stat(metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, fmt.Errorf("BOF metadata does not exist at %s", metaPath)
		}
		return nil, nil, fmt.Errorf("failed to stat BOF metadata: %w", err)
	}

	// Load BOF payload from file
	payload, err := os.ReadFile(payloadPath)
	if err != nil {
		logger.Logf(logger.Error, "Failed to read BOF file %s: %v", payloadPath, err)
		return nil, nil, fmt.Errorf("failed to read BOF file: %w", err)
	}

	// Load metadata from file
	metadataJSON, err := os.ReadFile(metaPath)
	if err != nil {
		logger.Logf(logger.Error, "Failed to read metadata file %s: %v", metaPath, err)
		return nil, nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var loadedMetadata BOFInfo
	err = json.Unmarshal(metadataJSON, &loadedMetadata)
	if err != nil {
		logger.Logf(logger.Error, "Failed to parse metadata: %v", err)
		return nil, nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	// Register orphaned BOF in database with a warning
	logger.Logf(logger.Warning, "Orphaned BOF detected (exists on disk but not in database). UUID=%s, Platform=%s, Function=%s",
		loadedMetadata.Uuid, loadedMetadata.Platform, loadedMetadata.FunctionName)
	logger.Logf(logger.Info, "Auto-registering orphaned BOF in database...")

	// Insert/update in database
	insertSQL := `
	INSERT INTO bofs (uuid, platform, function_name, description, uploaded_at)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (uuid) DO UPDATE SET
		platform = EXCLUDED.platform,
		function_name = EXCLUDED.function_name,
		description = EXCLUDED.description,
		uploaded_at = EXCLUDED.uploaded_at
	`
	_, err = db.Exec(insertSQL, loadedMetadata.Uuid, loadedMetadata.Platform, loadedMetadata.FunctionName, 
		loadedMetadata.Description, time.Now())
	if err != nil {
		logger.Logf(logger.Error, "Failed to register orphaned BOF in database: %v", err)
		// Don't fail the operation, but warn that registration failed
		return &loadedMetadata, payload, fmt.Errorf("failed to register BOF in database (file will still be used): %w", err)
	}

	logger.Logf(logger.Info, "Orphaned BOF auto-registered successfully: UUID=%s", loadedMetadata.Uuid)

	// Update cached entry
	fileCache.Store(platform+"_"+uuid, &loadedMetadata)

	return &loadedMetadata, payload, nil
}

// ListAllBOFs returns all BOFs from the global repository (not assigned to any agent)
func ListAllBOFs() ([]*BOFInfo, error) {
	bofs, err := GetAllBOFs()
	if err != nil {
		return nil, err
	}
	return bofs, nil
}

// CreateTask creates a new BOF execution task (legacy, without function name)
func CreateTask(agentUuid, bofUuid string, payload []byte) (*TaskInfo, error) {
	return CreateTaskWithFunction(agentUuid, bofUuid, payload, "go")
}

// CreateTaskWithFunction creates a new BOF execution task with function name
func CreateTaskWithFunction(agentUuid, bofUuid string, payload []byte, functionName string) (*TaskInfo, error) {
	// Get agent's platform from database and normalize it
	var osType string
	err := db.QueryRow("SELECT os_type FROM agents WHERE uuid = $1", agentUuid).Scan(&osType)
	if err != nil {
		logger.Logf(logger.Error, "Failed to get agent platform for %s: %v", agentUuid, err)
		// Default to windows if we can't determine
		platform := "windows"
		task := TaskInfo{
			Id:            fmt.Sprintf("%d", time.Now().UnixNano()),
			BofUuid:       bofUuid,
			AgentUuid:     agentUuid,
			Platform:      platform,
			FunctionName:  functionName,
			Status:        TaskPending,
			CreatedAt:     time.Now(),
		}

		// Save task to database
		insertSQL := `
		INSERT INTO tasks (id, bof_uuid, agent_uuid, platform, status, output, error, created_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NULL)
		ON CONFLICT (id) DO UPDATE SET
			bof_uuid = EXCLUDED.bof_uuid,
			agent_uuid = EXCLUDED.agent_uuid,
			platform = EXCLUDED.platform,
			status = EXCLUDED.status,
			output = EXCLUDED.output,
			error = EXCLUDED.error,
			created_at = EXCLUDED.created_at,
			deleted_at = EXCLUDED.deleted_at
		`
		_, err = db.Exec(insertSQL, task.Id, bofUuid, agentUuid, platform, string(TaskPending), nil, nil, time.Now())
		if err != nil {
			logger.Logf(logger.Error, "Failed to insert task into database: %v", err)
			return nil, fmt.Errorf("failed to insert task into database: %w", err)
		}

	return &task, nil
	}

	// Normalize the platform name (handle case variations like "Windows" vs "windows")
	platform := strings.ToLower(osType)
	if platform != "windows" && platform != "linux" {
		logger.Logf(logger.Warning, "Unknown agent os_type '%s', defaulting to windows", osType)
		platform = "windows"
	}

	task := TaskInfo{
		Id:            fmt.Sprintf("%d", time.Now().UnixNano()),
		BofUuid:       bofUuid,
		AgentUuid:     agentUuid,
		Platform:      platform,
		FunctionName:  functionName,
		Status:        TaskPending,
		CreatedAt:     time.Now(),
	}

	// Save task to database
	insertSQL := `
	INSERT INTO tasks (id, bof_uuid, agent_uuid, platform, status, output, error, created_at, deleted_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NULL)
	ON CONFLICT (id) DO UPDATE SET
		bof_uuid = EXCLUDED.bof_uuid,
		agent_uuid = EXCLUDED.agent_uuid,
		platform = EXCLUDED.platform,
		status = EXCLUDED.status,
		output = EXCLUDED.output,
		error = EXCLUDED.error,
		created_at = EXCLUDED.created_at,
		deleted_at = EXCLUDED.deleted_at
	`
	_, err = db.Exec(insertSQL, task.Id, bofUuid, agentUuid, platform, string(TaskPending), nil, nil, time.Now())
	if err != nil {
		logger.Logf(logger.Error, "Failed to insert task into database: %v", err)
		return nil, fmt.Errorf("failed to insert task into database: %w", err)
	}

	logger.Logf(logger.Info, "Created task %s for agent %s, BOF %s, function=%s", task.Id, agentUuid, bofUuid, functionName)
	return &task, nil
}

// GetPendingTasks returns pending tasks for a specific agent from the database (excludes soft-deleted)
func GetPendingTasks(agentUuid string) ([]*TaskInfo, error) {
	query := `SELECT id, bof_uuid, agent_uuid, status, output, error, created_at FROM tasks WHERE agent_uuid = $1 AND status = 'pending' AND deleted_at IS NULL ORDER BY created_at DESC`
	rows, err := db.Query(query, agentUuid)
	if err != nil {
		logger.Logf(logger.Error, "Failed to query pending tasks from database: %v", err)
		return []*TaskInfo{}, fmt.Errorf("failed to query pending tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*TaskInfo
	for rows.Next() {
		var task TaskInfo
		err := rows.Scan(&task.Id, &task.BofUuid, &task.AgentUuid, &task.Status, &task.Output, &task.Error, &task.CreatedAt)
		if err != nil {
			logger.Logf(logger.Error, "Failed to scan task row: %v", err)
			continue
		}
		tasks = append(tasks, &task)
	}

	if err = rows.Err(); err != nil {
		return tasks, fmt.Errorf("error iterating task rows: %w", err)
	}

	logger.Logf(logger.Info, "Found %d pending tasks for agent %s", len(tasks), agentUuid)
	return tasks, nil
}

// GetCompletedTasks returns completed/failed tasks for a specific agent from the database (excludes soft-deleted)
func GetCompletedTasks(agentUuid string) ([]*TaskInfo, error) {
	query := `SELECT id, bof_uuid, agent_uuid, status, output, error, created_at, completed_at FROM tasks WHERE agent_uuid = $1 AND status != 'pending' AND deleted_at IS NULL ORDER BY completed_at DESC`
	rows, err := db.Query(query, agentUuid)
	if err != nil {
		logger.Logf(logger.Error, "Failed to query completed tasks from database: %v", err)
		return []*TaskInfo{}, fmt.Errorf("failed to query completed tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*TaskInfo
	for rows.Next() {
		var task TaskInfo
		err := rows.Scan(&task.Id, &task.BofUuid, &task.AgentUuid, &task.Status, &task.Output, &task.Error, &task.CreatedAt, &task.CompletedAt)
		if err != nil {
			logger.Logf(logger.Error, "Failed to scan task row: %v", err)
			continue
		}
		tasks = append(tasks, &task)
	}

	if err = rows.Err(); err != nil {
		return tasks, fmt.Errorf("error iterating task rows: %w", err)
	}

	logger.Logf(logger.Info, "Found %d completed tasks for agent %s", len(tasks), agentUuid)
	return tasks, nil
}

// GetAllTasks returns all tasks for an agent including soft-deleted (for admin purposes)
func GetAllTasks(agentUuid string) ([]*TaskInfo, error) {
	query := `SELECT id, bof_uuid, agent_uuid, status, output, error, created_at, completed_at, deleted_at FROM tasks WHERE agent_uuid = $1 ORDER BY created_at DESC`
	rows, err := db.Query(query, agentUuid)
	if err != nil {
		logger.Logf(logger.Error, "Failed to query all tasks from database: %v", err)
		return []*TaskInfo{}, fmt.Errorf("failed to query all tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*TaskInfo
	for rows.Next() {
		var task TaskInfo
		err := rows.Scan(&task.Id, &task.BofUuid, &task.AgentUuid, &task.Status, &task.Output, &task.Error, &task.CreatedAt, &task.CompletedAt, &task.DeletedAt)
		if err != nil {
			logger.Logf(logger.Error, "Failed to scan task row: %v", err)
			continue
		}
		tasks = append(tasks, &task)
	}

	if err = rows.Err(); err != nil {
		return tasks, fmt.Errorf("error iterating task rows: %w", err)
	}

	logger.Logf(logger.Info, "Found %d total tasks (including soft-deleted) for agent %s", len(tasks), agentUuid)
	return tasks, nil
}

// MarkTaskCompleted updates the status of a task in the database (uses soft delete pattern)
func MarkTaskCompleted(taskId, agentUuid, status string, output string) error {
	updateSQL := `
	UPDATE tasks 
	SET status = $1, output = $2, completed_at = NOW(), deleted_at = NOW()
	WHERE id = $3 AND agent_uuid = $4 AND deleted_at IS NULL
	`
	_, err := db.Exec(updateSQL, status, output, taskId, agentUuid)
	if err != nil {
		logger.Logf(logger.Error, "Failed to update task in database: %v", err)
		return fmt.Errorf("failed to update task in database: %w", err)
	}

	logger.Logf(logger.Info, "Updated task %s status to %s (soft-deleted at %s)", taskId, status, time.Now().Format(time.RFC3339))
	return nil
}

// UpdateTaskStatus updates the status of a task in the database
func UpdateTaskStatus(taskId string, status TaskStatus, output string, errStr string) error {
	updateSQL := `
	UPDATE tasks 
	SET status = $1, output = $2, error = $3, completed_at = NOW()
	WHERE id = $4 AND deleted_at IS NULL
	`
	_, err := db.Exec(updateSQL, status, output, errStr, taskId)
	if err != nil {
		logger.Logf(logger.Error, "Failed to update task in database: %v", err)
		return fmt.Errorf("failed to update task in database: %w", err)
	}

	logger.Logf(logger.Info, "Updated task %s status to %s", taskId, status)
	return nil
}

// GetTaskPayload retrieves the payload for a specific task from the database
func GetTaskPayload(taskId string) ([]byte, error) {
	query := `SELECT payload FROM tasks WHERE id = $1`
	var payload []byte
	err := db.QueryRow(query, taskId).Scan(&payload)
	if err != nil {
		return nil, fmt.Errorf("failed to get task payload from database: %w", err)
	}
	return payload, nil
}

// ClearCache clears the file cache
func ClearCache() {
	fileCache.Range(func(key, value interface{}) bool {
		fileCache.Delete(key)
		return true
	})
	logger.Logf(logger.Info, "BOF cache cleared")
}


