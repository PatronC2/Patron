package data

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/PatronC2/Patron/Patronobuf/go/patronobuf"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/types"
	_ "github.com/lib/pq"
)

type AgentCounts struct {
	Total   int
	Online  int
	Offline int
}

func CreateAgent(req *patronobuf.ConfigurationRequest) error {
	CreateAgentSQL := `
	INSERT INTO agents (
		uuid, server_ip, server_port, callback_freq, callback_jitter,
		ip, agent_user, hostname, os_type, os_build, os_arch, cpus, memory, next_callback, transport_protocol
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`

	nextCallback := time.Unix(req.GetNextcallbackUnix(), 0)

	_, err := db.Exec(CreateAgentSQL,
		req.GetUuid(), req.GetServerip(), req.GetServerport(),
		req.GetCallbackfrequency(), req.GetCallbackjitter(),
		req.GetAgentip(), req.GetUsername(), req.GetHostname(),
		req.GetOstype(), req.GetOsbuild(), req.GetArch(),
		req.GetCpus(), req.GetMemory(), nextCallback, req.GetTransportprotocol(),
	)

	if err != nil {
		logger.Logf(logger.Error, "Error creating agent in DB: %v", err)
		return err
	}

	uuid := req.GetUuid()
	tags := map[string]string{
		"os_type":      req.GetOstype(),
		"os_build":     req.GetOsbuild(),
		"architecture": req.GetArch(),
		"username":     req.GetUsername(),
		"hostname":     req.GetHostname(),
	}

	for k, v := range tags {
		if err := PutAgentTags(uuid, k, v); err != nil {
			logger.Logf(logger.Error, "Error applying tag %q to agent %v: %v", k, uuid, err)
			return err
		}
	}

	err = PutAgentNotes(uuid, "")
	if err != nil {
		logger.Logf(logger.Error, "Error initializing notes for %v: %v", uuid, err)
	}

	logger.Logf(logger.Info, "New agent created in DB: %s", req.GetUuid())
	return nil
}

func FetchOneAgent(uuid string) (*patronobuf.ConfigurationRequest, error) {
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
		os_build,
		os_arch,
		cpus,
		memory,
		next_callback,
		status,
		transport_protocol
	FROM agents_status
	WHERE uuid = $1
	`

	var req patronobuf.ConfigurationRequest
	var nextCallback time.Time

	err := db.QueryRow(query, uuid).Scan(
		&req.Uuid,
		&req.Serverip,
		&req.Serverport,
		&req.Callbackfrequency,
		&req.Callbackjitter,
		&req.Agentip,
		&req.Username,
		&req.Hostname,
		&req.Ostype,
		&req.Osbuild,
		&req.Arch,
		&req.Cpus,
		&req.Memory,
		&nextCallback,
		&req.Status,
		&req.Transportprotocol,
	)

	if err == sql.ErrNoRows {
		logger.Logf(logger.Error, "No agent found with UUID: %s", uuid)
		return nil, nil
	} else if err != nil {
		logger.Logf(logger.Error, "Error fetching agent with UUID: %s - %v", uuid, err)
		return nil, err
	}

	req.NextcallbackUnix = nextCallback.Unix()
	logger.Logf(logger.Info, "Fetched agent: %s", req.Uuid)

	return &req, nil
}

func UpdateAgentConfig(UUID, ServerIP, ServerPort, CallbackFrequency, CallbackJitter string, NextCallback time.Time, TransportProtocol string) {
	updateAgentConfigSQL := `
	UPDATE agents 
	SET server_ip= $1, server_port= $2, callback_freq= $3, callback_jitter= $4, next_callback=$5, transport_protocol=$6
	WHERE "uuid"= $7`

	statement, err := db.Prepare(updateAgentConfigSQL)
	if err != nil {
		logger.Logf(logger.Error, "Error while updating agent config: %v", err)
	}

	_, err = statement.Exec(ServerIP, ServerPort, CallbackFrequency, CallbackJitter, NextCallback, TransportProtocol, UUID)
	if err != nil {
		logger.Logf(logger.Error, "Error while updating agent config: %v", err)
	}
	logger.Logf(logger.Info, "Agent %s Reveived Config Update  \n", UUID)
}

func UpdateAgentConfigNoNext(UUID, ServerIP, ServerPort, CallbackFrequency, CallbackJitter string, TransportProtocol string) {
	updateSQL := `
	UPDATE agents 
	SET server_ip = $1, server_port = $2, callback_freq = $3, callback_jitter = $4, transport_protocol = $5
	WHERE uuid = $6`

	_, err := db.Exec(updateSQL, ServerIP, ServerPort, CallbackFrequency, CallbackJitter, TransportProtocol, UUID)
	if err != nil {
		logger.Logf(logger.Error, "Error while updating agent config: %v", err)
	}
	logger.Logf(logger.Info, "Agent %s received config update (without next_callback)", UUID)
}

func UpdateAgentCheckIn(req *patronobuf.ConfigurationRequest) error {
	UpdateSQL := `
        UPDATE agents
        SET last_callback = NOW(), next_callback = $1
        WHERE uuid = $2`

	_, err := db.Exec(UpdateSQL, time.Unix(req.GetNextcallbackUnix(), 0), req.GetUuid())
	if err != nil {
		logger.Logf(logger.Error, "Error updating agent check-in for UUID %s: %v", req.GetUuid(), err)
		return err
	}

	logger.Logf(logger.Info, "Agent %s check-in updated in DB", req.GetUuid())
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
		t.tag_id,
		t.key,
		t.value
	FROM agents_status a
	LEFT JOIN agent_tags t ON a.uuid = t.uuid
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
			status, transportProtocol                                                         string
			tagID                                                                             sql.NullInt64
			tagKey, tagValue                                                                  sql.NullString
		)

		err := rows.Scan(&uuid, &serverIP, &serverPort, &callbackFreq, &callbackJitter, &ip, &agentUser,
			&hostname, &osType, &osArch, &osBuild, &cpus, &memory, &nextCallback, &status, &transportProtocol,
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
				TransportProtocol: transportProtocol,
				Tags:              []types.Tag{},
			}
		}

		if tagID.Valid && tagKey.Valid && tagValue.Valid {
			agentMap[uuid].Tags = append(agentMap[uuid].Tags, types.Tag{
				TagID: int64(tagID.Int64),
				Key:   tagKey.String,
				Value: tagValue.String,
			})
		}
	}

	// Flatten map into slice
	var agentList []types.ConfigurationRequest
	for _, agent := range agentMap {
		agentList = append(agentList, *agent)
	}

	return agentList, nil
}

func GetAgentCounts() (AgentCounts, error) {
	var c AgentCounts
	const q = `
		SELECT
		  COUNT(*)::int AS total,
		  SUM(CASE WHEN status = 'Online'  THEN 1 ELSE 0 END)::int AS online,
		  SUM(CASE WHEN status = 'Offline' THEN 1 ELSE 0 END)::int AS offline
		FROM agents_status;
	`
	err := db.QueryRow(q).Scan(&c.Total, &c.Online, &c.Offline)
	return c, err
}

func FilterAgents(filters map[string]string, tagFilters []string, logic string, limit, offset int, sort string) ([]types.ConfigurationRequest, int, int, error) {
	baseSelect := `
		SELECT 
			a.uuid, a.server_ip, a.server_port, a.callback_freq, a.callback_jitter,
			a.ip, a.agent_user, a.hostname, a.os_type, a.os_arch, a.os_build,
			a.cpus, a.memory, a.next_callback, a.status, a.transport_protocol
		FROM agents_status a`

	var (
		args       []interface{}
		conditions []string
	)

	// Basic filters
	if v := filters["hostname"]; v != "" {
		args = append(args, "%"+v+"%")
		conditions = append(conditions, fmt.Sprintf("a.hostname ILIKE $%d", len(args)))
	}
	if v := filters["ip"]; v != "" {
		args = append(args, "%"+v+"%")
		conditions = append(conditions, fmt.Sprintf("a.ip LIKE $%d", len(args)))
	}
	if v := filters["status"]; v != "" {
		args = append(args, v)
		conditions = append(conditions, fmt.Sprintf("a.status = $%d", len(args)))
	}

	// Tag filters (EXISTS-based; no JOIN => no duplicates)
	// Parse tagFilters into (key,val) pairs first
	type kv struct{ k, v string }
	pairs := make([]kv, 0, len(tagFilters))
	for _, tf := range tagFilters {
		parts := strings.SplitN(tf, ":", 2)
		if len(parts) != 2 {
			continue
		}
		k := strings.TrimSpace(parts[0])
		v := strings.TrimSpace(parts[1])
		if k == "" || v == "" {
			continue
		}
		pairs = append(pairs, kv{k: k, v: v})
	}

	if len(pairs) > 0 {
		if strings.ToLower(logic) == "and" {
			// Require all tag pairs to exist
			for _, p := range pairs {
				args = append(args, p.k, p.v)
				kIdx := len(args) - 1
				vIdx := len(args)
				conditions = append(conditions, fmt.Sprintf(`
					EXISTS (
						SELECT 1 FROM tags t
						WHERE t.uuid = a.uuid
						  AND t.key = $%d
						  AND t.value = $%d
					)`, kIdx, vIdx))
			}
		} else {
			// OR logic: any tag pair matches
			orParts := make([]string, 0, len(pairs))
			for _, p := range pairs {
				args = append(args, p.k, p.v)
				kIdx := len(args) - 1
				vIdx := len(args)
				orParts = append(orParts, fmt.Sprintf(`(t.key = $%d AND t.value = $%d)`, kIdx, vIdx))
			}
			conditions = append(conditions, `
				EXISTS (
					SELECT 1 FROM tags t
					WHERE t.uuid = a.uuid
					  AND (`+strings.Join(orParts, " OR ")+`)
				)`)
		}
	}

	// WHERE
	query := baseSelect
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Count query (now safe because query returns one row per agent)
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
		logger.Logf(logger.Error, "Failed to run query: %v. Error: %v", countQuery, err)
		return nil, 0, 0, fmt.Errorf("failed to count agents: %w", err)
	}

	// Execute main query
	rows, err := db.Query(query, args...)
	if err != nil {
		logger.Logf(logger.Error, "Failed to run query: %v. Error: %v", query, err)
		return nil, 0, 0, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	agents := make([]types.ConfigurationRequest, 0, limit)
	for rows.Next() {
		var agent types.ConfigurationRequest
		err := rows.Scan(
			&agent.AgentID, &agent.ServerIP, &agent.ServerPort, &agent.CallbackFrequency,
			&agent.CallbackJitter, &agent.AgentIP, &agent.Username, &agent.Hostname,
			&agent.OSType, &agent.OSArch, &agent.OSBuild, &agent.CPUS, &agent.MEMORY,
			&agent.NextCallback, &agent.Status, &agent.TransportProtocol,
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
		status,
		transport_protocol,
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
			&agents.TransportProtocol,
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
