package data

import (
	"database/sql"
	"log"

	"github.com/PatronC2/Patron/Patronobuf/go/patronobuf"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/types"
	_ "github.com/lib/pq"
)

func GetAgentCommands(uuid string) (infoAppend []types.AgentCommands, err error) {
	var info types.AgentCommands
	FetchSQL := `
	SELECT 
		"UUID", 
		"CommandType", 
		"Command", 
		"CommandUUID", 
		"Output"
	FROM "Commands"
	WHERE "UUID"= $1
	ORDER BY "CommandID" asc;
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

func UpdateAgentCommand(commandUUID, result, output, uuid string) error {
	updateSQL := `
		UPDATE "Commands"
		SET "Result" = $1, "Output" = $2
		WHERE "CommandUUID" = $3
	`

	_, err := db.Exec(updateSQL, result, output, commandUUID)
	if err != nil {
		logger.Logf(logger.Error, "Error updating command %s for agent %s: %v", commandUUID, uuid, err)
		return err
	}

	logger.Logf(logger.Info, "Command %s updated for agent %s", commandUUID, uuid)
	return nil
}

func FetchNextCommand(uuid string) *patronobuf.CommandResponse {
	query := `
		SELECT 
			"UUID", 
			"CommandType", 
			"Command", 
			"CommandUUID"
		FROM "Commands"
		WHERE "UUID" = $1 AND "Result" = '0'
		ORDER BY "CommandID" ASC
		LIMIT 1;
	`

	var resp patronobuf.CommandResponse
	err := db.QueryRow(query, uuid).Scan(
		&resp.Uuid,
		&resp.Commandtype,
		&resp.Command,
		&resp.Commandid,
	)
	if err == sql.ErrNoRows {
		logger.Logf(logger.Info, "No commands available for agent: %s", uuid)
		return &resp
	} else if err != nil {
		logger.Logf(logger.Error, "Error fetching command for agent %s: %v", uuid, err)
		return &resp
	}

	logger.Logf(logger.Info, "Fetched command for agent %s: %s", uuid, resp.Command)
	return &resp
}

func SendAgentCommand(uuid string, result string, CommandType string, Command string, CommandUUID string) {
	SendAgentCommandSQL := `INSERT INTO "Commands" ("UUID", "Result", "CommandType", "Command", "CommandUUID")
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
