package data

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
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

func Agents() ([]types.ConfigurationRequest, error) {
	/* DEPRECATED
	USE FilterAgents() INSTEAD!
	This function slams the DB, network, and user's browser
	This was fine when only dealing with <50 agents
	Remains until PatronCLI is updated to use the new function.
	*/
	query := `
	SELECT 
		a.uuid,
		a.server_ip,
		a.server_port,
		a.callback_freq,
		a.callback_jitter,
		a.ip,
		a.agent_user,
		a.hostname,
		a.os_type,
		a.os_arch,
		a.os_build,
		a.cpus,
		a.memory,
		a.next_callback,
		a.status,
		t."TagID",
		t."Key",
		t."Value"
	FROM agents_status a
	LEFT JOIN tags t ON a.uuid = t."UUID"
	`

	rows, err := db.Query(query)
	if err != nil {
		logger.Logf(logger.Error, "Error while getting agents: %v", err)
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	agentMap := make(map[string]*types.ConfigurationRequest)

	for rows.Next() {
		var (
			uuid, serverIP, serverPort, callbackFreq, callbackJitter, ip, agentUser, hostname string
			osType, osArch, osBuild, cpus, memory                                             string
			nextCallback                                                                      time.Time
			status                                                                            string
			tagID, tagKey, tagValue                                                           sql.NullString
		)

		err := rows.Scan(&uuid, &serverIP, &serverPort, &callbackFreq, &callbackJitter, &ip, &agentUser,
			&hostname, &osType, &osArch, &osBuild, &cpus, &memory, &nextCallback, &status,
			&tagID, &tagKey, &tagValue)
		if err != nil {
			logger.Logf(logger.Error, "Error scanning row from Agents: %v", err)
			continue
		}

		if _, exists := agentMap[uuid]; !exists {
			agentMap[uuid] = &types.ConfigurationRequest{
				AgentID:           uuid,
				ServerIP:          serverIP,
				ServerPort:        serverPort,
				CallbackFrequency: callbackFreq,
				CallbackJitter:    callbackJitter,
				AgentIP:           ip,
				Username:          agentUser,
				Hostname:          hostname,
				OSType:            osType,
				OSArch:            osArch,
				OSBuild:           osBuild,
				CPUS:              cpus,
				MEMORY:            memory,
				NextCallback:      nextCallback,
				Status:            status,
				Tags:              []types.Tag{},
			}
		}

		if tagID.Valid && tagKey.Valid && tagValue.Valid {
			tagIDInt, err := strconv.Atoi(tagID.String)
			if err != nil {
				logger.Logf(logger.Error, "Invalid TagID for agent %s: %v", uuid, err)
			} else {
				agentMap[uuid].Tags = append(agentMap[uuid].Tags, types.Tag{
					TagID: tagIDInt,
					Key:   tagKey.String,
					Value: tagValue.String,
				})
			}
		}
	}

	// Flatten map into slice
	var agentList []types.ConfigurationRequest
	for _, agent := range agentMap {
		agentList = append(agentList, *agent)
	}

	return agentList, nil
}

func FilterAgents(filters map[string]string, tagFilters []string, logic string, limit, offset int, sort string) ([]types.ConfigurationRequest, int, int, error) {
	baseSelect := `
		SELECT 
			a.uuid, a.server_ip, a.server_port, a.callback_freq, a.callback_jitter,
			a.ip, a.agent_user, a.hostname, a.os_type, a.os_arch, a.os_build,
			a.cpus, a.memory, a.next_callback, a.status
		FROM agents_status a`

	var (
		args          []interface{}
		conditions    []string
		tagConditions []string
		joinTags      bool
	)

	// Process basic filters
	if v := filters["hostname"]; v != "" {
		args = append(args, "%"+v+"%")
		conditions = append(conditions, fmt.Sprintf("a.hostname ILIKE $%d", len(args)))
	}
	if v := filters["ip"]; v != "" {
		args = append(args, v)
		conditions = append(conditions, fmt.Sprintf("a.ip = $%d", len(args)))
	}
	if v := filters["status"]; v != "" {
		args = append(args, v)
		conditions = append(conditions, fmt.Sprintf("a.status = $%d", len(args)))
	}

	// Handle tag filters
	for _, tf := range tagFilters {
		parts := strings.SplitN(tf, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key, val := parts[0], parts[1]
		args = append(args, key, val)
		tagConditions = append(tagConditions, fmt.Sprintf("(t.\"Key\" = $%d AND t.\"Value\" = $%d)", len(args)-1, len(args)))
	}
	joinTags = len(tagConditions) > 0

	var whereClause string
	if joinTags {
		conditions = append(conditions, "("+strings.Join(tagConditions, " OR ")+")")
	}
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	query := baseSelect
	if joinTags {
		query += " JOIN tags t ON t.\"UUID\" = a.uuid"
	}
	query += whereClause

	// If AND logic on tags, add GROUP BY + HAVING
	if joinTags && logic == "and" {
		havingIndex := len(args) + 1
		query += `
		GROUP BY a.agent_id, a.uuid, a.server_ip, a.server_port, a.callback_freq, a.callback_jitter,
		         a.ip, a.agent_user, a.hostname, a.os_type, a.os_arch, a.os_build, a.cpus, a.memory,
		         a.next_callback, a.status
		HAVING COUNT(DISTINCT t."Key") = $` + strconv.Itoa(havingIndex)
		args = append(args, len(tagFilters))
	}

	// Build count query
	countQuery := "SELECT COUNT(*) FROM (" + query + ") AS sub"

	// Sorting
	sortableFields := map[string]bool{
		"hostname": true, "ip": true, "status": true, "callback_freq": true, "next_callback": true,
	}
	if sort != "" {
		parts := strings.SplitN(sort, ":", 2)
		if len(parts) == 2 && sortableFields[parts[0]] {
			direction := strings.ToUpper(parts[1])
			if direction != "ASC" && direction != "DESC" {
				direction = "ASC"
			}
			query += fmt.Sprintf(" ORDER BY a.%s %s", parts[0], direction)
		}
	}

	// Pagination
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)

	// Execute count query
	var totalCount int
	if err := db.QueryRow(countQuery, args...).Scan(&totalCount); err != nil {
		return nil, 0, 0, fmt.Errorf("failed to count agents: %w", err)
	}

	// Execute main query
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	agents := make([]types.ConfigurationRequest, 0)
	for rows.Next() {
		var agent types.ConfigurationRequest
		err := rows.Scan(
			&agent.AgentID, &agent.ServerIP, &agent.ServerPort, &agent.CallbackFrequency,
			&agent.CallbackJitter, &agent.AgentIP, &agent.Username, &agent.Hostname,
			&agent.OSType, &agent.OSArch, &agent.OSBuild, &agent.CPUS, &agent.MEMORY,
			&agent.NextCallback, &agent.Status,
		)
		if err != nil {
			logger.Logf(logger.Error, "Error scanning agent: %v", err)
			continue
		}
		agents = append(agents, agent)
	}

	nextOffset := offset + len(agents)
	return agents, totalCount, nextOffset, nil
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
