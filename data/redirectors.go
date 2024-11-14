package data

import (
	"log"

	"github.com/PatronC2/Patron/types"	
	"github.com/PatronC2/Patron/lib/logger"	
	_ "github.com/lib/pq"
)


func GetRedirectors() (redirectors []types.Redirector, err error) {
	var data types.Redirector
	FetchSQL := `
	SELECT
		"RedirectorID",
		"Name",
		"Description",
		"ForwardIP",
		"ForwardPort",
		"ListenPort",
		"Status"
	FROM "redirector_status"
	`
	rows, err := db.Query(FetchSQL)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(
			&data.RedirectorID,
			&data.Name,
			&data.Description,
			&data.ForwardIP,
			&data.ForwardPort,
			&data.ListenPort,
			&data.Status,
		)
		if err != nil {
			log.Println("Error scanning row:", err)
			return nil, err
		}
		redirectors = append(redirectors, data)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error iterating over rows:", err)
		return nil, err
	}

	logger.Logf(logger.Info, "Current redirectors: %+v\n", redirectors)
	return redirectors, nil
}

func CreateRedirector(RedirectorID, Name, Description, ForwardIP, ForwardPort, ListenPort string) error {
    InsertSQL := `
        INSERT INTO "redirectors" ("RedirectorID", "Name", "Description", "ForwardIP", "ForwardPort", "ListenPort")
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `

    _, err := db.Exec(InsertSQL, RedirectorID, Name, Description, ForwardIP, ForwardPort, ListenPort)
    if err != nil {
        logger.Logf(logger.Error, "Error creating redirector with RedirectorID %s: %v", RedirectorID, err)
        return err
    }

    logger.Logf(logger.Info, "Successfully created redirector with RedirectorID %s", RedirectorID)
    return nil
}

func SetRedirectorStatus(RedirectorID string) error {
    UpdateSQL := `
        UPDATE "redirectors"
        SET "LastReport" = NOW()
        WHERE "RedirectorID" = $1;
    `

    _, err := db.Exec(UpdateSQL, RedirectorID)
    if err != nil {
        logger.Logf(logger.Error, "Error updating redirector status for RedirectorID %s: %v", RedirectorID, err)
        return err
    }

    logger.Logf(logger.Info, "Updated redirector status for RedirectorID %s", RedirectorID)
    return nil
}