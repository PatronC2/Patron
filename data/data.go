package data

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/PatronC2/Patron/types"
	_ "github.com/mattn/go-sqlite3"
	"github.com/s-christian/gollehs/lib/logger"
)

var db *sql.DB

func OpenDatabase() error {
	var err error
	// os.Remove("./data/sqlite-database.db")
	db, err = sql.Open("sqlite3", "./data/sqlite-database.db")
	if err != nil {
		return err
	}
	return db.Ping()
}

func InitDatabase() {
	AgentSQL := `
	CREATE TABLE IF NOT EXISTS "Agents" (
		"AgentID"	INTEGER NOT NULL UNIQUE,
		"UUID"	TEXT NOT NULL UNIQUE,
		"CallBackUser"	TEXT NOT NULL DEFAULT 'Unknown',
		"CallBackToIP"	TEXT NOT NULL DEFAULT 'Unknown',
		"CallBackFeq"	TEXT NOT NULL DEFAULT 'Unknown',
		"CallBackJitter"	TEXT NOT NULL DEFAULT 'Unknown',
		"Ip"	TEXT NOT NULL DEFAULT 'Unknown',
		"User"	TEXT NOT NULL DEFAULT 'Unknown',
		"Hostname"	TEXT NOT NULL DEFAULT 'Unknown',
		"isDeleted"	INTEGER NOT NULL DEFAULT 0,
		"LastCallBack"	TEXT DEFAULT 'Unknown',
		PRIMARY KEY("AgentID" AUTOINCREMENT)
	);
	`
	AgentSQLstatement, err := db.Prepare(AgentSQL)
	if err != nil {
		log.Fatal(err.Error())
	}

	AgentSQLstatement.Exec()
	logger.Logf(logger.Info, "Agents table created\n")

	CommandSQL := `
	CREATE TABLE IF NOT EXISTS "Commands" (
		"CommandID"	INTEGER NOT NULL UNIQUE,
		"UUID"	TEXT,
		"Result"	TEXT,
		"CommandType"	TEXT,
		"Command"	TEXT,
		"CommandUUID"	TEXT,
		"Output" TEXT DEFAULT 'Unknown',
		PRIMARY KEY("CommandID" AUTOINCREMENT),
		FOREIGN KEY("UUID") REFERENCES "Agents"("UUID")
	);
	`
	CommandSQLstatement, err := db.Prepare(CommandSQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	CommandSQLstatement.Exec()
	logger.Logf(logger.Info, "Commands table created\n")
}

func CreateAgent(uuid string, CallBackToIP string, CallBackFeq string, CallBackJitter string, Ip string, User string, Hostname string) {
	CreateAgentSQL := `INSERT INTO Agents (UUID, CallBackToIP, CallBackFeq, CallBackJitter, Ip, User, Hostname)
	VALUES (?, ?, ?, ?, ?, ?, ?)`

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

func FetchOneAgent(uuid string) types.ConfigAgent {
	FetchSQL := `
	SELECT 
		UUID,CallBackToIP,CallbackFeq,CallBackJitter
	FROM Agents WHERE UUID=$1
	`
	agentStruct := types.ConfigAgent{}
	row := db.QueryRow(FetchSQL, uuid)

	err := row.Scan(
		&agentStruct.Uuid,
		&agentStruct.CallbackTo,
		&agentStruct.CallbackFrequency,
		&agentStruct.CallbackJitter,
	)
	switch err {
	case sql.ErrNoRows:
		logger.Logf(logger.Info, "No rows were returned! \n")
		return agentStruct
	case nil:
		fmt.Println(agentStruct)
	default:
		panic(err)
	}

	logger.Logf(logger.Info, "Agent %s Fetched \n", agentStruct.Uuid)
	return agentStruct
}

func FetchNextCommand(uuid string) types.GiveAgentCommand {
	FetchSQL := `
	SELECT 
		Commands.UUID, 
		Agents.CallBackToIP, 
		Agents.CallBackFeq, 
		Agents.CallBackJitter, 
		Commands.CommandType, 
		Commands.Command, 
		Commands.CommandUUID
	FROM Commands INNER JOIN 
		Agents ON Commands.UUID = Agents.UUID 
	WHERE Commands.UUID=$1 AND Result='0' LIMIT 1
	`
	agentStruct := types.GiveAgentCommand{}
	agentStruct.Binary = nil
	row := db.QueryRow(FetchSQL, uuid)

	err := row.Scan(
		&agentStruct.UpdateAgentConfig.Uuid,
		&agentStruct.UpdateAgentConfig.CallbackTo,
		&agentStruct.UpdateAgentConfig.CallbackFrequency,
		&agentStruct.UpdateAgentConfig.CallbackJitter,
		&agentStruct.CommandType,
		&agentStruct.Command,
		&agentStruct.CommandUUID,
		// &agentStruct.Binary,
	)
	switch err {
	case sql.ErrNoRows:
		logger.Logf(logger.Info, "No rows were returned! \n")
	case nil:
		fmt.Println(agentStruct)
	default:
		panic(err)
	}

	logger.Logf(logger.Info, "Agent %s Fetched Next Command %s \n", agentStruct.UpdateAgentConfig.Uuid, agentStruct.Command)
	return agentStruct
}
func SendAgentCommand(uuid string, result string, CommandType string, Command string, CommandUUID string) {
	SendAgentCommandSQL := `INSERT INTO Commands (UUID, Result, CommandType, Command, CommandUUID)
	VALUES (?, ?, ?, ?, ?)`

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

func UpdateAgentCommand(CommandUUID string, Output string, uuid string) {
	updateAgentCommandSQL := `UPDATE Commands SET Result='1', Output= ? WHERE CommandUUID= ?`

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

// WEB Functions

func Agents() []types.ConfigAgent {
	var agents types.ConfigAgent
	FetchSQL := `
	SELECT 
		UUID, 
		CallBackToIP, 
		CallBackFeq, 
		CallBackJitter, 
		Ip, 
		User, 
		Hostname
	FROM Agents
	WHERE isDeleted='0'
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
		)
		agentAppend = append(agentAppend, agents)
	}
	return agentAppend
}

func Agent(uuid string) []types.Agent {
	var info types.Agent
	FetchSQL := `
	SELECT 
		UUID, 
		CommandType, 
		Command, 
		CommandUUID, 
		Output
	FROM Commands
	WHERE UUID= ? AND CommandType = 'shell'
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
	// logger.Logf(logger.Info, "Agent %s Fetched Next Command %s \n", agentStruct.UpdateAgentConfig.Uuid, agentStruct.Command)
	// return agentStruct
}
