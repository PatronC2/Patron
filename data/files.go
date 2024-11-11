package data

import (
	"database/sql"

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
