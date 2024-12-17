package data

import (
	"log"

	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/types"
	_ "github.com/lib/pq"
)

func GetAgentNotes(uuid string) (infoAppend []types.Note, err error) {
	var info types.Note
	FetchSQL := `
	SELECT 
		"note_id",
		"note"
	FROM "notes" WHERE "uuid"=$1
	`
	row, err := db.Query(FetchSQL, uuid)
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()
	for row.Next() {
		row.Scan(
			&info.NoteID,
			&info.Note,
		)
	}
	infoAppend = append(infoAppend, info)
	logger.Logf(logger.Info, "%v\n", info)
	return infoAppend, err
}

func PutAgentNotes(uuid string, note string) error {
	UpsertSQL := `
    INSERT INTO "notes" ("uuid", "note")
    VALUES ($1, $2)
    ON CONFLICT ("uuid")
    DO UPDATE SET "note" = $2;
    `
	_, err := db.Exec(UpsertSQL, uuid, note)
	if err != nil {
		log.Fatalln(err)
		return err
	}
	logger.Logf(logger.Info, "Notes for UUID %v have been updated in DB\n", uuid)
	return nil
}
