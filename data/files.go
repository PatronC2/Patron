package data

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/PatronC2/Patron/Patronobuf/go/patronobuf"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/types"
)

// Where file bytes live (mount a Docker volume here).
const agentFilesDir = "/app/agent_files"

func ensureAgentFilesDir() error {
	return os.MkdirAll(agentFilesDir, 0o750)
}

func fileDiskPath(fileID int64) string {
	// Store as /app/agent_files/<file_id>
	return filepath.Join(agentFilesDir, fmt.Sprintf("%d", fileID))
}

func atomicWriteFile(dst string, data []byte, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o750); err != nil {
		return err
	}
	tmp := dst + ".tmp"
	if err := os.WriteFile(tmp, data, perm); err != nil {
		logger.Logf(logger.Error, "Error writing file to disk: %v", err)
		return err
	}
	logger.Logf(logger.Info, "Successfully wrote %v to disk", dst)
	return os.Rename(tmp, dst)
}

// FetchNextFileTransfer returns the next pending transfer for an agent.
// If the transfer type is "Download", it reads bytes from disk and sets resp.Chunk.
func FetchNextFileTransfer(uuid string) *patronobuf.FileResponse {
	query := `
		SELECT file_id, uuid, type, path
		FROM files
		WHERE uuid = $1
		  AND status = 'Pending'
		ORDER BY file_id ASC
		LIMIT 1;
	`

	var (
		resp   patronobuf.FileResponse
		fileID int64
	)

	err := db.QueryRow(query, uuid).Scan(
		&fileID,
		&resp.Uuid,
		&resp.Transfertype,
		&resp.Filepath,
	)
	if err == sql.ErrNoRows {
		logger.Logf(logger.Info, "No pending file transfers for agent: %s", uuid)
		return nil
	}
	if err != nil {
		logger.Logf(logger.Error, "Error fetching file transfer: %v", err)
		return nil
	}

	resp.Fileid = strconv.FormatInt(fileID, 10)

	// Only include bytes when the agent is supposed to download from server
	if resp.GetTransfertype() == "Download" {
		if err := ensureAgentFilesDir(); err != nil {
			logger.Logf(logger.Error, "Failed to ensure agent files dir: %v", err)
			return nil
		}

		diskPath := fileDiskPath(fileID)
		b, err := os.ReadFile(diskPath)
		if err != nil {
			logger.Logf(logger.Error, "Failed to read file bytes for file_id=%d (%s): %v", fileID, diskPath, err)
			return nil
		}
		resp.Chunk = b
	}

	logger.Logf(logger.Info, "Fetched file transfer file_id=%s type=%s for agent=%s", resp.Fileid, resp.Transfertype, uuid)
	return &resp
}

// UpdateFileTransfer updates status for a transfer. For "Upload" it also writes bytes to disk.
// For "Download" it only updates status (agent confirms it downloaded / completed).
func UpdateFileTransfer(file *patronobuf.FileToServer) error {
	fileID, err := strconv.ParseInt(file.GetFileid(), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid file_id %q: %w", file.GetFileid(), err)
	}

	switch file.GetTransfertype() {
	case "Download":
		// Agent reports status for a download transfer (no bytes to store).
		query := `
			UPDATE files
			SET status = $1
			WHERE file_id = $2 AND uuid = $3 AND type = $4;
		`
		if _, err := db.Exec(query, file.GetStatus(), fileID, file.GetUuid(), file.GetTransfertype()); err != nil {
			return fmt.Errorf("failed to update download transfer: %w", err)
		}

	case "Upload":
		// Agent sends bytes to server -> store bytes on disk, then update status.
		if err := ensureAgentFilesDir(); err != nil {
			return fmt.Errorf("ensure agent files dir: %w", err)
		}
		diskPath := fileDiskPath(fileID)
		if err := atomicWriteFile(diskPath, file.GetChunk(), 0o640); err != nil {
			return fmt.Errorf("write uploaded bytes: %w", err)
		}

		query := `
			UPDATE files
			SET status = $1
			WHERE file_id = $2 AND uuid = $3 AND type = $4;
		`
		if _, err := db.Exec(query, file.GetStatus(), fileID, file.GetUuid(), file.GetTransfertype()); err != nil {
			return fmt.Errorf("failed to update upload transfer: %w", err)
		}

	default:
		return fmt.Errorf("unknown transfer type: %s", file.GetTransfertype())
	}

	logger.Logf(logger.Debug, "File transfer file_id=%d updated successfully", fileID)
	return nil
}

func ListFilesForUUID(uuid string) ([]types.FileToServer, error) {
	query := `
		SELECT file_id, path, status
		FROM files
		WHERE uuid = $1
		ORDER BY file_id DESC;
	`

	rows, err := db.Query(query, uuid)
	if err != nil {
		logger.Logf(logger.Error, "Error listing files for UUID %s: %v", uuid, err)
		return nil, err
	}
	defer rows.Close()

	var files []types.FileToServer
	for rows.Next() {
		var f types.FileToServer
		if err := rows.Scan(&f.FileID, &f.Path, &f.Status); err != nil {
			logger.Logf(logger.Error, "Error scanning file for UUID %s: %v", uuid, err)
			return nil, err
		}
		files = append(files, f)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return files, nil
}

// DownloadFile is your HTTP API download endpoint helper.
// It reads bytes from disk for file_id and uses basename(path) as filename.
func DownloadFile(fileIDStr string) ([]byte, string, error) {
	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		return nil, "", fmt.Errorf("invalid file_id %q: %w", fileIDStr, err)
	}

	var agentPath string
	query := `SELECT path FROM files WHERE file_id = $1;`

	err = db.QueryRow(query, fileID).Scan(&agentPath)
	if err == sql.ErrNoRows {
		logger.Logf(logger.Info, "No file found with file_id: %s", fileIDStr)
		return nil, "", nil
	}
	if err != nil {
		return nil, "", err
	}

	if err := ensureAgentFilesDir(); err != nil {
		return nil, "", err
	}

	diskPath := fileDiskPath(fileID)
	b, err := os.ReadFile(diskPath)
	if err != nil {
		return nil, "", err
	}

	return b, filepath.Base(agentPath), nil
}

// UploadFile queues a transfer request.
// If transfertype == "Download" (agent downloads from server), content must be present and will be written to disk.
// If transfertype == "Upload" (agent uploads to server), content is ignored (no bytes yet).
func UploadFile(path string, uuid string, transfertype string, content []byte) error {
	if err := ensureAgentFilesDir(); err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var fileID int64
	insert := `
		INSERT INTO files (uuid, type, path, status)
		VALUES ($1, $2, $3, 'Pending')
		RETURNING file_id;
	`
	if err := tx.QueryRow(insert, uuid, transfertype, path).Scan(&fileID); err != nil {
		return err
	}

	// Only write bytes at request creation time if the agent will DOWNLOAD them from us.
	if transfertype == "Download" {
		if len(content) == 0 {
			return fmt.Errorf("Download transfer requires non-empty content")
		}
		diskPath := fileDiskPath(fileID)
		if err := atomicWriteFile(diskPath, content, 0o640); err != nil {
			return err
		}
	}

	// For "Upload", we just queue; agent will send bytes later via UpdateFileTransfer.

	if err := tx.Commit(); err != nil {
		return err
	}

	logger.Logf(logger.Info, "Queued file transfer file_id=%d type=%s uuid=%s path=%s", fileID, transfertype, uuid, path)
	return nil
}
