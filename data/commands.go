package data

import (
	"database/sql"
	"log"

	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/types"
	_ "github.com/lib/pq"
)

func GetAgentCommands(uuid string) (infoAppend []types.AgentCommands, err error) {
	var info types.AgentCommands
	FetchSQL := `
	SELECT 
		"uuid", 
		"command_type", 
		"command", 
		"command_uuid", 
		"output"
	FROM "commands"
	WHERE "uuid"= $1 AND "command_type" = 'shell'
	ORDER BY "command_id" asc;
	`
	row, err := db.Query(FetchSQL, uuid)
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()
	for row.Next() {
		row.Scan(
			&info.Uuid,
			&info.CommandType,
			&info.Command,
			&info.CommandUUID,
			&info.Output,
		)
		infoAppend = append(infoAppend, info)
	}

	return infoAppend, err
}

func FetchNextCommand(uuid string) types.CommandResponse {
	var info types.CommandResponse
	query := `
        SELECT 
            "commands"."uuid", 
            "commands"."command_type", 
            "commands"."command", 
            "commands"."command_uuid"
        FROM "commands" 
        INNER JOIN "agents" ON "commands"."uuid" = "agents"."uuid" 
        WHERE "commands"."uuid" = $1 
        AND "commands"."result" = '0' 
        LIMIT 1;
    `

	row := db.QueryRow(query, uuid)
	err := row.Scan(
		&info.AgentID,
		&info.CommandType,
		&info.Command,
		&info.CommandID,
	)
	if err == sql.ErrNoRows {
		logger.Logf(logger.Info, "No commands available for agent: %s\n", uuid)
		return info
	} else if err != nil {
		logger.Logf(logger.Error, "Error fetching command for agent: %v\n", err)
		return info
	}

	logger.Logf(logger.Info, "Fetched command %s for agent %s\n", info.Command, uuid)
	return info
}

func SendAgentCommand(uuid string, result string, CommandType string, Command string, CommandUUID string) {
	SendAgentCommandSQL := `INSERT INTO "commands" ("uuid", "result", "command_type", "command", "command_uuid")
	VALUES ($1, $2, $3, $4, $5)`

	statement, err := db.Prepare(SendAgentCommandSQL)
	if err != nil {

		log.Fatalln(err)
		logger.Logf(logger.Info, "Error in DB\n")
	}

	_, err = statement.Exec(uuid, result, CommandType, Command, CommandUUID)
	if err != nil {

		log.Fatalln(err)
	}
	logger.Logf(logger.Info, "Agent %s Reveived New Command \n", uuid)
}
