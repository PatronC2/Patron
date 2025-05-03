package data

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/types"
	_ "github.com/lib/pq"
)

func CreateAgent(uuid, ServerIP, ServerPort, CallBackFreq, CallBackJitter, Ip, User, Hostname, OSType, OSBuild, OSArch, CPUS, MEMORY string, NextCallback time.Time) error {
	CreateAgentSQL := `
	INSERT INTO agents (
		uuid, server_ip, server_port, callback_freq, callback_jitter,
		ip, agent_user, hostname, os_type, os_build, os_arch, cpus, memory, next_callback
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	_, err := db.Exec(CreateAgentSQL, uuid, ServerIP, ServerPort, CallBackFreq, CallBackJitter, Ip, User, Hostname, OSType, OSBuild, OSArch, CPUS, MEMORY, NextCallback)
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
		uuid,
		server_ip,
		server_port,
		callback_freq,
		callback_jitter,
		ip,
		agent_user,
		hostname,
		os_type,
		os_arch,
		os_build,
		cpus,
		memory,
		next_callback,
		status
	FROM agents_status
	WHERE uuid = $1
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
		&info.NextCallback,
		&info.Status,
	)

	if err == sql.ErrNoRows {
		logger.Logf(logger.Error, "No agent found with UUID: %s", uuid)
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
		uuid, server_ip, server_port, callback_freq, callback_jitter
	FROM "agents" WHERE "uuid"=$1
	`
	row, err := db.Query(FetchSQL, uuid)
	if err != nil {
		logger.Logf(logger.Error, "Error Fetching one agent: %v", err)
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

func UpdateAgentConfig(UUID, ServerIP, ServerPort, CallbackFrequency, CallbackJitter string, NextCallback time.Time) {
	updateAgentConfigSQL := `
	UPDATE agents 
	SET server_ip= $1, server_port= $2, callback_freq= $3, callback_jitter= $4, next_callback=$5
	WHERE "uuid"= $6`

	statement, err := db.Prepare(updateAgentConfigSQL)
	if err != nil {
		logger.Logf(logger.Error, "Error while updating agent config: %v", err)
	}

	_, err = statement.Exec(ServerIP, ServerPort, CallbackFrequency, CallbackJitter, NextCallback, UUID)
	if err != nil {
		logger.Logf(logger.Error, "Error while updating agent config: %v", err)
	}
	logger.Logf(logger.Info, "Agent %s Reveived Config Update  \n", UUID)
}

func UpdateAgentConfigNoNext(UUID, ServerIP, ServerPort, CallbackFrequency, CallbackJitter string) {
	updateSQL := `
	UPDATE agents 
	SET server_ip = $1, server_port = $2, callback_freq = $3, callback_jitter = $4
	WHERE uuid = $5`

	_, err := db.Exec(updateSQL, ServerIP, ServerPort, CallbackFrequency, CallbackJitter, UUID)
	if err != nil {
		logger.Logf(logger.Error, "Error while updating agent config: %v", err)
	}
	logger.Logf(logger.Info, "Agent %s received config update (without next_callback)", UUID)
}

func UpdateAgentCheckIn(confreq types.ConfigurationRequest) error {
	UpdateSQL := `
        UPDATE agents
        SET last_callback = NOW(), next_callback = $1
        WHERE uuid = $2`

	_, err := db.Exec(UpdateSQL, confreq.NextCallback.UTC(), confreq.AgentID)
	if err != nil {
		logger.Logf(logger.Error, "Error updating agent check-in for UUID %s: %v", confreq.AgentID, err)
		return err
	}

	logger.Logf(logger.Info, "Agent %s check-in updated in DB", confreq.AgentID)
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
		uuid,
		server_ip, 
		server_port, 
		callback_freq,
		callback_jitter,
		ip, 
		agent_user, 
		hostname,
		os_type,
		os_arch,
		os_build,
		cpus,
		memory,
		next_callback,
		status
	FROM agents_status
	`
	row, err := db.Query(FetchSQL)
	if err != nil {
		logger.Logf(logger.Error, "Error while getting agents: %v", err)
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
			&agents.NextCallback,
			&agents.Status,
		)
		if err != nil {
			logger.Logf(logger.Error, "Error scanning agent row: %v", err)
			return nil, err
		}

		tags, err := GetAgentTags(agents.AgentID)
		if err != nil {
			logger.Logf(logger.Error, "Error fetching tags for agent: %v", err)
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
		uuid,
		server_ip, 
		server_port, 
		callback_freq,
		callback_jitter,
		ip, 
		agent_user, 
		hostname,
		os_type,
		os_arch,
		os_build,
		cpus,
		memory,
		next_callback,
		status
	FROM agents_status
	WHERE "Ip" = $1
	`
	row, err := db.Query(FetchSQL, Ip)
	if err != nil {
		logger.Logf(logger.Error, "Error fetching agents by ip: %v", err)
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
			&agents.NextCallback,
			&agents.Status,
		)
		agentAppend = append(agentAppend, agents)
	}
	return agentAppend, err
}

func GetAgentsMetrics() (agentsMetrics types.AgentMetrics, err error) {
	metricsSQL := `
	SELECT status, COUNT(*) AS count
	FROM agents_status
	WHERE status IN ('Online', 'Offline')
	GROUP BY status;
	`
	rows, err := db.Query(metricsSQL)
	if err != nil {
		logger.Logf(logger.Error, "Error fetching agents metrics: %v", err)
	}

	agentsMetrics.OnlineCount = "0"
	agentsMetrics.OfflineCount = "0"

	for rows.Next() {
		var status string
		var count int

		if err := rows.Scan(&status, &count); err != nil {
			return agentsMetrics, fmt.Errorf("failed to scan row: %w", err)
		}

		switch status {
		case "Online":
			agentsMetrics.OnlineCount = strconv.Itoa(count)
		case "Offline":
			agentsMetrics.OfflineCount = strconv.Itoa(count)
		}
	}

	if err := rows.Err(); err != nil {
		return agentsMetrics, fmt.Errorf("row iteration error: %w", err)
	}

	return agentsMetrics, nil
}
