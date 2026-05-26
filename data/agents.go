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
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type AgentCounts struct {
	Total   int
	Online  int
	Offline int
}

func FindOrCreateAgentFromStartup(req *patronobuf.StartupRequest) (string, error) {
	existing, err := FindAgentByStartup(req)
	if err != nil {
		return "", err
	}
	if existing != "" {
		if err := UpdateAgentFromStartup(existing, req); err != nil {
			return "", err
		}
		return existing, nil
	}

	newUUID := uuid.New().String()
	if err := CreateAgentFromStartup(newUUID, req); err != nil {
		return "", err
	}
	return newUUID, nil
}

func FindAgentByStartup(req *patronobuf.StartupRequest) (string, error) {
	query := `
	SELECT uuid
	FROM agents
	WHERE file_path = $1
	  AND agent_user = $2
	  AND hostname = $3
	  AND os_type = $4
	  AND os_arch = $5
	  AND os_build = $6
	  AND cpus = $7
	  AND memory = $8
	ORDER BY agent_id ASC
	LIMIT 1`

	var agentUUID string
	err := db.QueryRow(query,
		req.GetFilepath(),
		req.GetUsername(),
		req.GetHostname(),
		req.GetOstype(),
		req.GetArch(),
		req.GetOsbuild(),
		req.GetCpus(),
		req.GetMemory(),
	).Scan(&agentUUID)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		logger.Logf(logger.Error, "Error finding agent from startup request: %v", err)
		return "", err
	}
	return agentUUID, nil
}

func UpdateAgentFromStartup(agentUUID string, req *patronobuf.StartupRequest) error {
	updateSQL := `
	UPDATE agents
	SET file_path = $1,
		ip = $2,
		agent_user = $3,
		hostname = $4,
		os_type = $5,
		os_build = $6,
		os_arch = $7,
		cpus = $8,
		memory = $9,
		capabilities = $10
	WHERE uuid = $11`

	_, err := db.Exec(updateSQL,
		req.GetFilepath(),
		req.GetAgentip(),
		req.GetUsername(),
		req.GetHostname(),
		req.GetOstype(),
		req.GetOsbuild(),
		req.GetArch(),
		req.GetCpus(),
		req.GetMemory(),
		pq.Array(req.GetCapabilities()),
		agentUUID,
	)
	if err != nil {
		logger.Logf(logger.Error, "Error updating agent startup fields for %s: %v", agentUUID, err)
		return err
	}
	return nil
}

func CreateAgentFromStartup(agentUUID string, req *patronobuf.StartupRequest) error {
	CreateAgentSQL := `
	INSERT INTO agents (
		uuid, file_path, ip, agent_user, hostname, os_type, os_build, os_arch, cpus, memory, capabilities, transport_protocol
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 'Unknown')`

	_, err := db.Exec(CreateAgentSQL,
		agentUUID, req.GetFilepath(), req.GetAgentip(), req.GetUsername(), req.GetHostname(),
		req.GetOstype(), req.GetOsbuild(), req.GetArch(), req.GetCpus(), req.GetMemory(),
		pq.Array(req.GetCapabilities()),
	)

	if err != nil {
		logger.Logf(logger.Error, "Error creating agent in DB: %v", err)
		return err
	}

	tags := map[string]string{
		"os_type":      req.GetOstype(),
		"os_build":     req.GetOsbuild(),
		"architecture": req.GetArch(),
		"username":     req.GetUsername(),
		"hostname":     req.GetHostname(),
	}

	for k, v := range tags {
		if err := PutAgentTags(agentUUID, k, v); err != nil {
			logger.Logf(logger.Error, "Error applying tag %q to agent %v: %v", k, agentUUID, err)
			return err
		}
	}

	err = PutAgentNotes(agentUUID, "")
	if err != nil {
		logger.Logf(logger.Error, "Error initializing notes for %v: %v", agentUUID, err)
	}

	logger.Logf(logger.Info, "New agent created in DB: %s", agentUUID)
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
		next_callback,
		status,
		COALESCE(transport_protocol, 'Unknown') AS transport_protocol
	FROM agents_status
	WHERE uuid = $1
	`

	var req patronobuf.ConfigurationRequest
	var nextCallback sql.NullTime

	err := db.QueryRow(query, uuid).Scan(
		&req.Uuid,
		&req.Serverip,
		&req.Serverport,
		&req.Callbackfrequency,
		&req.Callbackjitter,
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

	if nextCallback.Valid {
		req.NextcallbackUnix = nextCallback.Time.Unix()
	}
	logger.Logf(logger.Info, "Fetched agent: %s", req.Uuid)

	return &req, nil
}

func FetchOneAgentDetails(uuid string) (*types.ConfigurationRequest, error) {
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
		capabilities,
		next_callback,
		status,
		COALESCE(transport_protocol, 'Unknown') AS transport_protocol
	FROM agents_status
	WHERE uuid = $1
	`

	var agent types.ConfigurationRequest
	var nextCallback sql.NullTime

	err := db.QueryRow(query, uuid).Scan(
		&agent.AgentID,
		&agent.ServerIP,
		&agent.ServerPort,
		&agent.CallbackFrequency,
		&agent.CallbackJitter,
		&agent.AgentIP,
		&agent.Username,
		&agent.Hostname,
		&agent.OSType,
		&agent.OSArch,
		&agent.OSBuild,
		&agent.CPUS,
		&agent.MEMORY,
		pq.Array(&agent.Capabilities),
		&nextCallback,
		&agent.Status,
		&agent.TransportProtocol,
	)
	if err == sql.ErrNoRows {
		logger.Logf(logger.Error, "No agent found with UUID: %s", uuid)
		return nil, nil
	}
	if err != nil {
		logger.Logf(logger.Error, "Error fetching agent details with UUID: %s - %v", uuid, err)
		return nil, err
	}

	if nextCallback.Valid {
		agent.NextCallback = nextCallback.Time
	}

	return &agent, nil
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

func UpdateAgentConfigIfMissing(UUID, ServerIP, ServerPort, CallbackFrequency, CallbackJitter string, TransportProtocol string) {
	updateSQL := `
	UPDATE agents
	SET server_ip = CASE WHEN server_ip IS NULL OR server_ip = '' OR server_ip = 'Unknown' THEN $1 ELSE server_ip END,
		server_port = CASE WHEN server_port IS NULL OR server_port = '' OR server_port = 'Unknown' THEN $2 ELSE server_port END,
		callback_freq = CASE WHEN callback_freq IS NULL OR callback_freq = '' OR callback_freq = 'Unknown' THEN $3 ELSE callback_freq END,
		callback_jitter = CASE WHEN callback_jitter IS NULL OR callback_jitter = '' OR callback_jitter = 'Unknown' THEN $4 ELSE callback_jitter END,
		transport_protocol = CASE WHEN transport_protocol IS NULL OR transport_protocol = '' OR transport_protocol = 'Unknown' THEN $5 ELSE transport_protocol END
	WHERE uuid = $6`

	_, err := db.Exec(updateSQL, ServerIP, ServerPort, CallbackFrequency, CallbackJitter, TransportProtocol, UUID)
	if err != nil {
		logger.Logf(logger.Error, "Error while seeding missing agent config: %v", err)
	}
}

func UpdateAgentNextCallback(UUID string, NextCallback time.Time) {
	updateSQL := `
	UPDATE agents
	SET next_callback = $1
	WHERE uuid = $2`

	_, err := db.Exec(updateSQL, NextCallback, UUID)
	if err != nil {
		logger.Logf(logger.Error, "Error while updating agent next callback: %v", err)
	}
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
        SET last_callback = NOW()
        WHERE uuid = $1`

	_, err := db.Exec(UpdateSQL, req.GetUuid())
	if err != nil {
		logger.Logf(logger.Error, "Error updating agent check-in for UUID %s: %v", req.GetUuid(), err)
		return err
	}

	logger.Logf(logger.Info, "Agent %s check-in updated in DB", req.GetUuid())
	return nil
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
						SELECT 1 FROM agent_tags t
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
					SELECT 1 FROM agent_tags t
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
