package data

import (
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/types"
	_ "github.com/lib/pq"
)

func GetAgentNotes(uuid string) (types.Note, error) {
	const q = `
		SELECT uuid, note, updated_at
		FROM agent_notes
		WHERE uuid = $1;
	`

	var n types.Note
	err := db.QueryRow(q, uuid).Scan(&n.Uuid, &n.Note, &n.UpdatedAt)
	if err != nil {
		logger.Logf(logger.Error, "Error fetching notes for agent %v: %v", uuid, err)
	}
	return n, err
}

func PutAgentNotes(uuid, note string) error {
	const q = `
		INSERT INTO agent_notes (uuid, note, updated_at)
		VALUES ($1, $2, now())
		ON CONFLICT (uuid)
		DO UPDATE SET note = EXCLUDED.note, updated_at = now();
	`
	_, err := db.Exec(q, uuid, note)
	if err != nil {
		logger.Logf(logger.Error, "Error putting notes for agent %v: %v", uuid, note)
	}
	return err
}
