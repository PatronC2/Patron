package data

import (
	"log"

	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/types"
	_ "github.com/lib/pq"
)

func GetAgentTags(uuid string) (infoAppend []types.Tag, err error) {
	var info types.Tag
	FetchSQL := `
	SELECT
		"tag_id",
		"key",
		"value"
	FROM "tags" WHERE "uuid"=$1
	`
	rows, err := db.Query(FetchSQL, uuid)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(
			&info.TagID,
			&info.Key,
			&info.Value,
		)
		if err != nil {
			log.Println("Error scanning row:", err)
			return nil, err
		}
		infoAppend = append(infoAppend, info)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error iterating over rows:", err)
		return nil, err
	}

	logger.Logf(logger.Info, "Tags for %v: %+v\n", uuid, infoAppend)
	return infoAppend, nil
}

func PutAgentTags(uuid string, key string, value string) error {
	PutTagsSQL := `
    INSERT INTO "tags" ("uuid", "key", "value")
    VALUES ($1, $2, $3)
    ON CONFLICT ("uuid", "key") DO UPDATE 
    SET "value" = EXCLUDED."value"
    `
	_, err := db.Exec(PutTagsSQL, uuid, key, value)
	if err != nil {
		log.Fatalln(err)
		return err
	}
	logger.Logf(logger.Info, "Tags for %v have been updated in DB\n", uuid)
	return nil
}

func DeleteTag(tagid string) error {
	DeleteTagsSQL := `
    DELETE FROM "tags"
	WHERE "tag_id" = $1
    `
	_, err := db.Exec(DeleteTagsSQL, tagid)
	if err != nil {
		log.Fatalln(err)
		return err
	}
	logger.Logf(logger.Info, "Tag %d has been deleted\n", tagid)
	return nil
}
