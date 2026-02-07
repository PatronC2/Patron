package data

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/PatronC2/Patron/lib/logger"
)

type TagKV struct {
	Key   string  `json:"key"`
	Value *string `json:"value"`
}

type KeylogDocument struct {
	KeylogID  int64     `json:"keylog_id"`
	Contents  *string   `json:"contents"`
	CreatedAt time.Time `json:"created_at"`
	UUID      string    `json:"uuid"`
	IP        string    `json:"ip"`
	Tags      []TagKV   `json:"tags"`
}

func GetIndexerCheckpoint(name string) (int64, error) {
	var lastID int64
	err := db.QueryRow(`SELECT last_keylog_id FROM opensearch_checkpoints WHERE name = $1`, name).Scan(&lastID)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return lastID, nil
}

func UpdateIndexerCheckpoint(name string, lastID int64) error {
	_, err := db.Exec(`
		INSERT INTO opensearch_checkpoints (name, last_keylog_id, updated_at)
		VALUES ($1, $2, now())
		ON CONFLICT (name)
		DO UPDATE SET last_keylog_id = EXCLUDED.last_keylog_id, updated_at = now()
	`, name, lastID)
	return err
}

func FetchKeylogDocumentsSince(lastKeylogID int64, limit int) ([]KeylogDocument, int64, error) {
	const q = `
	SELECT
		k.keylog_id,
		k.contents,
		k.created_at,
		a.uuid,
		a.ip,
		COALESCE(t.tags, '[]'::jsonb) AS tags
	FROM keylogs k
	JOIN agents a ON a.uuid = k.uuid
	LEFT JOIN (
		SELECT
			uuid,
			jsonb_agg(jsonb_build_object('key', key, 'value', value) ORDER BY key) AS tags
		FROM agent_tags
		GROUP BY uuid
	) t ON t.uuid = k.uuid
	WHERE k.keylog_id > $1
	ORDER BY k.keylog_id
	LIMIT $2
	`

	rows, err := db.Query(q, lastKeylogID, limit)
	if err != nil {
		return nil, lastKeylogID, err
	}
	defer rows.Close()

	var (
		docs  []KeylogDocument
		maxID = lastKeylogID
	)

	for rows.Next() {
		var (
			doc       KeylogDocument
			contents  sql.NullString
			tagsBytes []byte
		)
		if err := rows.Scan(&doc.KeylogID, &contents, &doc.CreatedAt, &doc.UUID, &doc.IP, &tagsBytes); err != nil {
			return nil, maxID, err
		}
		if contents.Valid {
			doc.Contents = &contents.String
		}
		if len(tagsBytes) > 0 {
			if err := json.Unmarshal(tagsBytes, &doc.Tags); err != nil {
				return nil, maxID, fmt.Errorf("unmarshal tags for keylog_id=%d: %w", doc.KeylogID, err)
			}
		}
		docs = append(docs, doc)
		if doc.KeylogID > maxID {
			maxID = doc.KeylogID
		}
	}
	if err := rows.Err(); err != nil {
		return nil, maxID, err
	}

	logger.Logf(logger.Debug, "Fetched %d keylog docs after keylog_id=%d", len(docs), lastKeylogID)
	return docs, maxID, nil
}
