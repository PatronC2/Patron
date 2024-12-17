package data

import (
	"log"

	"github.com/PatronC2/Patron/helper"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/types"
	_ "github.com/lib/pq"
)

func CreateKeys(uuid string) error {
	CreateKeysSQL := `
        INSERT INTO "keylog" ("uuid", "keys")
        VALUES ($1, $2)`

	_, err := db.Exec(CreateKeysSQL, uuid, "")
	if err != nil {
		logger.Logf(logger.Error, "Error creating keylog entry in DB for UUID %s: %v", uuid, err)
		return err
	}

	logger.Logf(logger.Info, "New keylog entry created for agent %s", uuid)
	return nil
}

func UpdateAgentKeys(UUID, Keys string) error {
	updateAgentKeylogSQL := `
        UPDATE "keylog"
        SET "keys" = "keys" || $1
        WHERE "uuid" = $2
    `
	_, err := db.Exec(updateAgentKeylogSQL, Keys, UUID)
	if err != nil {
		logger.Logf(logger.Error, "Error updating keys for agent with UUID %s: %v", UUID, err)
		return err
	}

	logger.Logf(logger.Info, "Successfully updated keys for agent with UUID %s", UUID)
	return nil
}

func Keylog(uuid string) []types.KeysRequest {
	var info types.KeysRequest
	FetchSQL := `
	SELECT 
		"uuid",
		"keys"
	FROM "keylog"
	WHERE "uuid"= $1
	ORDER BY "keylog_id" asc;
	`
	row, err := db.Query(FetchSQL, uuid)
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()
	var infoAppend []types.KeysRequest
	for row.Next() {
		row.Scan(
			&info.AgentID,
			&info.Keys,
		)
		info.Keys = helper.FormatKeyLogs(info.Keys)
		infoAppend = append(infoAppend, info)
	}
	return infoAppend
}
