package data

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"github.com/PatronC2/Patron/Patronobuf/go/patronobuf"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/types"
	_ "github.com/lib/pq"
)

func FetchNextFileTransfer(uuid string) *patronobuf.FileResponse {
	query := `
        SELECT 
            "files"."FileID",
            "files"."UUID",
            "files"."Type",
            "files"."Path",
            "files"."Content"
        FROM "files" 
        INNER JOIN "agents" ON "files"."UUID" = agents.uuid 
        WHERE "files"."UUID" = $1
        AND "files"."Status" = 'Pending' 
        LIMIT 1;
    `

	var (
		resp    patronobuf.FileResponse
		content []byte
	)

	err := db.QueryRow(query, uuid).Scan(
		&resp.Fileid,
		&resp.Uuid,
		&resp.Transfertype,
		&resp.Filepath,
		&content,
	)
	if err == sql.ErrNoRows {
		logger.Logf(logger.Info, "No pending file transfers for agent: %s", uuid)
		return nil
	}
	if err != nil {
		logger.Logf(logger.Error, "Error fetching file transfer: %v", err)
		return nil
	}

	resp.Chunk = content
	logger.Logf(logger.Info, "Fetched file transfer %s for agent %s", resp.Fileid, uuid)
	return &resp
}

func UpdateFileTransfer(file *patronobuf.FileToServer) error {
	var query string
	var args []interface{}

	switch file.GetTransfertype() {
	case "Download":
		query = `UPDATE files SET "Status" = $1 WHERE "FileID" = $2 AND "UUID" = $3 AND "Type" = $4`
		args = append(args, file.GetStatus(), file.GetFileid(), file.GetUuid(), file.GetTransfertype())

	case "Upload":
		query = `UPDATE files SET "Status" = $1, "Content" = $2 WHERE "FileID" = $3 AND "UUID" = $4 AND "Type" = $5`
		args = append(args, file.GetStatus(), file.GetChunk(), file.GetFileid(), file.GetUuid(), file.GetTransfertype())

	default:
		return fmt.Errorf("unknown transfer type: %s", file.GetTransfertype())
	}

	_, err := db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update file transfer: %w", err)
	}

	logger.Logf(logger.Debug, "File transfer with ID %s updated successfully", file.GetFileid())
	return nil
}

func ListFilesForUUID(uuid string) ([]types.FileToServer, error) {
	var files []types.FileToServer
	query := `
        SELECT 
            "FileID",
            "Path",
            "Status"
        FROM "files"
        WHERE "UUID" = $1;
    `
	rows, err := db.Query(query, uuid)
	if err != nil {
		logger.Logf(logger.Error, "Error listing files for UUID %s: %v\n", uuid, err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var file types.FileToServer
		err := rows.Scan(&file.FileID, &file.Path, &file.Status)
		if err != nil {
			logger.Logf(logger.Error, "Error scanning file for UUID %s: %v\n", uuid, err)
			return nil, err
		}
		files = append(files, file)
	}

	if err = rows.Err(); err != nil {
		logger.Logf(logger.Error, "Error with result rows for UUID %s: %v\n", uuid, err)
		return nil, err
	}

	return files, nil
}

func DownloadFile(fileID string) ([]byte, string, error) {
	var content []byte
	var path string
	query := `
        SELECT "Content", "Path"
        FROM "files"
        WHERE "FileID" = $1;
    `

	err := db.QueryRow(query, fileID).Scan(&content, &path)
	if err == sql.ErrNoRows {
		logger.Logf(logger.Info, "No file found with FileID: %s\n", fileID)
		return nil, "", nil
	} else if err != nil {
		logger.Logf(logger.Error, "Error downloading file with FileID %s: %v\n", fileID, err)
		return nil, "", err
	}

	return content, filepath.Base(path), nil
}

func UploadFile(path string, uuid string, transfertype string, content []byte) error {
	// "Download" is from the agent's perspective, not the API.
	// If it's an "Upload", we don't need to store content
	if transfertype == "Upload" {
		// Insert file with no content for uploads
		query := `
			INSERT INTO "files" ("UUID", "Type", "Path")
			VALUES ($1, $2, $3);
		`
		_, err := db.Exec(query, uuid, transfertype, path)
		if err != nil {
			logger.Logf(logger.Error, "Error uploading file for UUID %s to path %s: %v\n", uuid, path, err)
			return err
		}
		logger.Logf(logger.Info, "File uploaded for UUID %s to path %s\n", uuid, path)
	} else if transfertype == "Download" {
		// Insert file with content for downloads
		query := `
			INSERT INTO "files" ("UUID", "Type", "Path", "Content")
			VALUES ($1, $2, $3, $4);
		`
		_, err := db.Exec(query, uuid, transfertype, path, content)
		if err != nil {
			logger.Logf(logger.Error, "Error uploading file for UUID %s to path %s: %v\n", uuid, path, err)
			return err
		}
		logger.Logf(logger.Info, "File uploaded for UUID %s to path %s\n", uuid, path)
	}

	return nil
}
