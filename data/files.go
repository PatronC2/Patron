package data

import (
	"database/sql"
    "fmt"
    "path/filepath"

	"github.com/PatronC2/Patron/types"	
	"github.com/PatronC2/Patron/lib/logger"	
	_ "github.com/lib/pq"
)

func FetchNextFileTransfer(uuid string) types.FileResponse {
    var info types.FileResponse
    query := `
        SELECT 
            "files"."FileID",
            "files"."UUID",
            "files"."Type",
            "files"."Path",
            "files"."Content"
        FROM "files" 
        INNER JOIN "agents" ON "files"."UUID" = "agents"."UUID" 
        WHERE "files"."UUID" = $1
        AND "files"."Status" = 'Pending' 
        LIMIT 1;
    `

    row := db.QueryRow(query, uuid)
    var content []byte
    err := row.Scan(
        &info.FileID,
        &info.AgentID,
        &info.Type,
        &info.Path,
        &content,
    )
    if err == sql.ErrNoRows {
        logger.Logf(logger.Info, "No pending file transfers for agent: %s\n", uuid)
        return info
    } else if err != nil {
        logger.Logf(logger.Error, "Error fetching file transfers for agent: %v\n", err)
        return info
    }

    info.Chunk = content
    logger.Logf(logger.Info, "Fetched file transfer %s for agent %s\n", info.FileID, uuid)
    return info
}

func UpdateFileTransfer(fileData types.FileToServer) error {
	var query string
	var args []interface{}

	if fileData.Type == "Download" {
		query = `UPDATE files SET "Status" = $1 WHERE "FileID" = $2 AND "UUID" = $3 AND "Type" = $4`
		args = append(args, fileData.Status, fileData.FileID, fileData.AgentID, fileData.Type)
	} else if fileData.Type == "Upload" {
		query = `UPDATE files SET "Status" = $1, "Content" = $2 WHERE "FileID" = $3 AND "UUID" = $4 AND "Type" = $5`
		args = append(args, fileData.Status, fileData.Chunk, fileData.FileID, fileData.AgentID, fileData.Type)
	} else {
		return fmt.Errorf("unknown transfer type: %s", fileData.Type)
	}

	_, err := db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update file transfer: %v", err)
	}

	fmt.Printf("File transfer with ID %s updated successfully\n", fileData.FileID)
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
