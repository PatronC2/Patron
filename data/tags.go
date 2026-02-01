package data

import (
	"github.com/PatronC2/Patron/types"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

func GetAgentTags(uuid string) ([]types.Tag, error) {
	const q = `
		SELECT tag_id, key, value
		FROM agent_tags
		WHERE uuid = $1
		ORDER BY tag_id ASC;
	`

	rows, err := db.Query(q, uuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]types.Tag, 0, 16)
	for rows.Next() {
		var t types.Tag
		if err := rows.Scan(&t.TagID, &t.Key, &t.Value); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func PutAgentTags(uuid, key, value string) error {
	const q = `
		INSERT INTO agent_tags (uuid, key, value, updated_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (uuid, key) DO UPDATE
		SET value = EXCLUDED.value,
		    updated_at = now();
	`
	_, err := db.Exec(q, uuid, key, value)
	return err
}

func DeleteTag(tagID int64) error {
	const q = `DELETE FROM agent_tags WHERE tag_id = $1;`
	_, err := db.Exec(q, tagID)
	return err
}
func GetTagKeyValues() ([]types.TagKeyValues, error) {
	const q = `
		SELECT key, array_agg(DISTINCT value) AS values
		FROM agent_tags
		GROUP BY key
		ORDER BY key;
	`

	rows, err := db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []types.TagKeyValues
	for rows.Next() {
		var kv types.TagKeyValues
		if err := rows.Scan(&kv.Key, pq.Array(&kv.Values)); err != nil {
			return nil, err
		}
		results = append(results, kv)
	}
	return results, rows.Err()
}
