package data

import (
	"log"

	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/types"
	_ "github.com/lib/pq"
)

func CreatePayload(uuid string, name string, description string, ServerIP string, ServerPort string, CallBackFreq string, CallBackJitter string, Concat string) {
	CreateAgentSQL := `INSERT INTO "Payloads" ("UUID", "Name", "Description", "ServerIP", "ServerPort", "CallbackFrequency", "CallbackJitter", "Concat")
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	statement, err := db.Prepare(CreateAgentSQL)
	if err != nil {

		logger.Logf(logger.Info, "Error in DB\n")
	}

	_, err = statement.Exec(uuid, name, description, ServerIP, ServerPort, CallBackFreq, CallBackJitter, Concat)
	if err != nil {
		logger.Logf(logger.Error, "Error in DB: %s", err)
	}
	logger.Logf(logger.Info, "New Payload created in DB\n")
}

func Payloads() []types.Payload {
	var payloads types.Payload
	FetchSQL := `
	SELECT
		"PayloadID",
		"UUID", 
		"Name",
		"Description",
		"ServerIP", 
		"ServerPort", 
		"CallbackFrequency", 
		"CallbackJitter",
		"Concat" 
	FROM "Payloads"
	WHERE "isDeleted"='0'
	`
	row, err := db.Query(FetchSQL)
	if err != nil {
		logger.Logf(logger.Error, "Error in DB: %s", err)
	}
	defer row.Close()
	var payloadAppend []types.Payload
	for row.Next() {
		row.Scan(
			&payloads.PayloadID,
			&payloads.Uuid,
			&payloads.Name,
			&payloads.Description,
			&payloads.ServerIP,
			&payloads.ServerPort,
			&payloads.CallbackFrequency,
			&payloads.CallbackJitter,
			&payloads.Concat,
		)
		payloadAppend = append(payloadAppend, payloads)
	}
	return payloadAppend
}

func DeletePayload(payloadid string) error {
	DeleteSQL := `
    UPDATE "Payloads"
    SET "isDeleted" = 1
    WHERE "PayloadID" = $1`

	statement, err := db.Prepare(DeleteSQL)
	if err != nil {
		logger.Logf(logger.Error, "Error preparing statement: %s", err)
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(payloadid)
	if err != nil {
		logger.Logf(logger.Error, "Error executing statement: %s", err)
		return err
	}

	logger.Logf(logger.Info, "Payload with ID %s marked as deleted in DB", payloadid)
	return nil
}

func GetPayloadConcat(payloadID string) (string, error) {
	var payloadConcat string
	FetchNameSQL := `
    SELECT "Concat"
    FROM "Payloads"
    WHERE "PayloadID" = $1 AND "isDeleted" = 0
    `

	err := db.QueryRow(FetchNameSQL, payloadID).Scan(&payloadConcat)
	if err != nil {
		log.Println("Error fetching payload name:", err)
		return "", err
	}

	return payloadConcat, nil
}
