package data

import (
	"database/sql"

	"github.com/PatronC2/Patron/Patronobuf/go/patronobuf"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/types"
	_ "github.com/lib/pq"
)

func GetAgentCommands(uuid string) ([]types.AgentCommands, error) {
	// Returns a list of the commands for a given agent, only used by API
	const q = `
		SELECT uuid, command_type, command, command_uuid, output
		FROM commands
		WHERE uuid = $1
		ORDER BY command_id ASC;
	`

	rows, err := db.Query(q, uuid)
	if err != nil {
		logger.Logf(logger.Error, "Error getting agent commands for agent %v: %v", uuid, err)
		return nil, err
	}
	defer rows.Close()

	out := make([]types.AgentCommands, 0, 32)
	for rows.Next() {
		var row types.AgentCommands
		if err := rows.Scan(
			&row.Uuid,
			&row.CommandType,
			&row.Command,
			&row.CommandUUID,
			&row.Output,
		); err != nil {
			logger.Logf(logger.Error, "Error fetching a command for agent %v: %v", uuid, err)
			return nil, err
		}
		out = append(out, row)
	}

	return out, rows.Err()
}

func UpdateAgentCommand(resp *patronobuf.CommandStatusRequest) error {
	// Updates the command's row once its been executed by the agent
	const q = `
		UPDATE commands
		SET result = $1,
		    output = $2,
		    updated_at = now()
		WHERE command_uuid = $3;
	`

	res, err := db.Exec(q, resp.GetResult(), resp.GetOutput(), resp.GetCommandid())
	if err != nil {
		logger.Logf(logger.Error, "Error updating command uuid %v: %v", resp.GetCommandid(), err)
		return err
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		logger.Logf(logger.Warning, "No rows affected, expected 1 for commandUUID %v", resp.GetCommandid())
		return sql.ErrNoRows
	}
	return nil
}

func FetchNextCommand(uuid string) (*patronobuf.CommandResponse, error) {
	// Grabs the next command for the agent to execute
	const q = `
		SELECT uuid, command_type, command, command_uuid
		FROM commands
		WHERE uuid = $1 AND result = '0'
		ORDER BY command_id ASC
		LIMIT 1;
	`

	var resp patronobuf.CommandResponse
	err := db.QueryRow(q, uuid).Scan(
		&resp.Uuid,
		&resp.Commandtype,
		&resp.Command,
		&resp.Commandid,
	)
	if err == sql.ErrNoRows {
		logger.Logf(logger.Debug, "No commands to execute for agent %v: %v", uuid, err)
		return nil, nil
	}
	if err != nil {
		logger.Logf(logger.Error, "Error fetching commands for agent %v: %v", uuid, err)
		return nil, err
	}
	return &resp, nil
}

func SendAgentCommand(uuid, result, commandType, command, commandUUID string) error {
	// Used by the API for users to send commands to the agents
	const q = `
		INSERT INTO commands (uuid, result, command_type, command, command_uuid)
		VALUES ($1, $2, $3, $4, $5);
	`

	_, err := db.Exec(q, uuid, result, commandType, command, commandUUID)
	if err != nil {
		logger.Logf(logger.Error, "Error inserting command for agent %v: %v", uuid, err)
	}
	return err
}
