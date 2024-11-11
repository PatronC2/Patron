package data

import (
	"database/sql"
    "fmt"

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

