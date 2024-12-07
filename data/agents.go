package data

import (
	"database/sql"
	"log"

	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/types"
	_ "github.com/lib/pq"
)

func CreateAgent(uuid, ServerIP, ServerPort, CallBackFreq, CallBackJitter, Ip, User, Hostname, OSType, OSBuild, OSArch, CPUS, MEMORY string) error {
	CreateAgentSQL := `
        INSERT INTO "agents" ("UUID", "ServerIP", "ServerPort", "CallBackFreq", "CallBackJitter", "Ip", "User", "Hostname", "OSType", "OSBuild", "OSArch", "CPUS", "MEMORY")
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err := db.Exec(CreateAgentSQL, uuid, ServerIP, ServerPort, CallBackFreq, CallBackJitter, Ip, User, Hostname, OSType, OSBuild, OSArch, CPUS, MEMORY)
	if err != nil {
		logger.Logf(logger.Error, "Error creating agent in DB: %v", err)
		return err
	}

	logger.Logf(logger.Info, "New agent created in DB: %s", uuid)
	return nil
}

func FetchOneAgent(uuid string) (info types.ConfigurationRequest, err error) {
	query := `
        SELECT 
            "UUID",
            "ServerIP",
            "ServerPort",
            "CallBackFreq",
            "CallBackJitter",
            "Ip",
            "User",
            "Hostname",
			"OSType",
			"OSArch",
			"OSBuild",
			"CPUS",
			"MEMORY",
            "Status"
        FROM "agents_status" WHERE "UUID"=$1
    `

	err = db.QueryRow(query, uuid).Scan(
		&info.AgentID,
		&info.ServerIP,
		&info.ServerPort,
		&info.CallbackFrequency,
		&info.CallbackJitter,
		&info.AgentIP,
		&info.Username,
		&info.Hostname,
		&info.OSType,
		&info.OSArch,
		&info.OSBuild,
		&info.CPUS,
		&info.MEMORY,
		&info.Status,
	)

	if err == sql.ErrNoRows {
		logger.Logf(logger.Info, "No agent found with UUID: %s", uuid)
		return info, nil
	} else if err != nil {
		logger.Logf(logger.Error, "Error fetching agent with UUID: %s - %v", uuid, err)
		return info, err
	}

	logger.Logf(logger.Info, "Fetched agent: %v", info)
	return info, nil
}

func FetchOne(uuid string) (infoAppend []types.ConfigurationResponse, err error) {
	var info types.ConfigurationResponse
	FetchSQL := `
	SELECT 
		"UUID","ServerIP", "ServerPort","CallBackFreq","CallBackJitter"
	FROM "agents" WHERE "UUID"=$1
	`
	row, err := db.Query(FetchSQL, uuid)
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()
	for row.Next() {
		row.Scan(
			&info.AgentID,
			&info.ServerIP,
			&info.ServerPort,
			&info.CallbackFrequency,
			&info.CallbackJitter,
		)
	}
	infoAppend = append(infoAppend, info)
	logger.Logf(logger.Info, "%v\n", info)
	return infoAppend, err
}

func UpdateAgentConfig(UUID string, ServerIP string, ServerPort string, CallbackFrequency string, CallbackJitter string) {
	updateAgentConfigSQL := `UPDATE "agents" SET "ServerIP"= $1, "ServerPort"= $2, "CallBackFreq"= $3, "CallBackJitter"= $4 WHERE "UUID"= $5`

	statement, err := db.Prepare(updateAgentConfigSQL)
	if err != nil {

		log.Fatalln(err)
		logger.Logf(logger.Info, "Error in DB\n")
	}

	_, err = statement.Exec(ServerIP, ServerPort, CallbackFrequency, CallbackJitter, UUID)
	if err != nil {

		log.Fatalln(err)
	}
	logger.Logf(logger.Info, "Agent %s Reveived Config Update  \n", UUID)
}

func UpdateAgentCheckIn(uuid string) error {
	UpdateSQL := `
        UPDATE "agents"
        SET "LastCallBack" = NOW()
        WHERE "UUID" = $1`

	_, err := db.Exec(UpdateSQL, uuid)
	if err != nil {
		logger.Logf(logger.Error, "Error updating agent check-in for UUID %s: %v", uuid, err)
		return err
	}

	logger.Logf(logger.Info, "Agent %s check-in updated in DB", uuid)
	return nil
}

func UpdateAgentCommand(CommandUUID, Result, Output, uuid string) error {
	updateAgentCommandSQL := `
        UPDATE "Commands"
        SET "Result" = $1, "Output" = $2
        WHERE "CommandUUID" = $3`

	_, err := db.Exec(updateAgentCommandSQL, Result, Output, CommandUUID)
	if err != nil {
		logger.Logf(logger.Error, "Error updating command for CommandUUID %s: %v", CommandUUID, err)
		return err
	}

	logger.Logf(logger.Info, "Command %s updated for agent %s", CommandUUID, uuid)
	return nil
}

func Agents() (agentAppend []types.ConfigurationRequest, err error) {
	var agents types.ConfigurationRequest
	FetchSQL := `
	SELECT 
		"UUID",
		"ServerIP", 
		"ServerPort", 
		"CallBackFreq",
		"CallBackJitter",
		"Ip", 
		"User", 
		"Hostname",
		"OSType",
		"OSArch",
		"OSBuild",
		"CPUS",
		"MEMORY",
		"Status"
	FROM "agents_status"
	`
	row, err := db.Query(FetchSQL)
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()

	for row.Next() {
		err := row.Scan(
			&agents.AgentID,
			&agents.ServerIP,
			&agents.ServerPort,
			&agents.CallbackFrequency,
			&agents.CallbackJitter,
			&agents.AgentIP,
			&agents.Username,
			&agents.Hostname,
			&agents.OSType,
			&agents.OSArch,
			&agents.OSBuild,
			&agents.CPUS,
			&agents.MEMORY,
			&agents.Status,
		)
		if err != nil {
			log.Println("Error scanning agent row:", err)
			return nil, err
		}

		tags, err := GetAgentTags(agents.AgentID)
		if err != nil {
			log.Println("Error fetching tags for agent:", agents.AgentID, err)
			return nil, err
		}

		agentWithTags := agents
		agentWithTags.Tags = tags

		agentAppend = append(agentAppend, agentWithTags)
	}
	logger.Logf(logger.Info, "Agents: %+v", agentAppend)
	return agentAppend, err
}

func AgentsByIp(Ip string) (agentAppend []types.ConfigurationRequest, err error) {
	var agents types.ConfigurationRequest
	FetchSQL := `
	SELECT 
		"UUID", 
		"ServerIP",
		"ServerPort",
		"CallBackFreq", 
		"CallBackJitter", 
		"Ip", 
		"User", 
		"Hostname",
		"OSType",
		"OSArch",
		"OSBuild",
		"CPUS",
		"MEMORY",
		"Status"
	FROM "agents_status"
	AND "Ip" = $1
	`
	row, err := db.Query(FetchSQL, Ip)
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()
	for row.Next() {
		row.Scan(
			&agents.AgentID,
			&agents.ServerIP,
			&agents.ServerPort,
			&agents.CallbackFrequency,
			&agents.CallbackJitter,
			&agents.AgentIP,
			&agents.Username,
			&agents.Hostname,
			&agents.OSType,
			&agents.OSArch,
			&agents.OSBuild,
			&agents.CPUS,
			&agents.MEMORY,
			&agents.Status,
		)
		agentAppend = append(agentAppend, agents)
	}
	return agentAppend, err
}
