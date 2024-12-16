package data

import (
	"database/sql"
	"fmt"
	"log"
	"os"

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
		"agent_id" SERIAL PRIMARY KEY,
		"uuid" TEXT NOT NULL UNIQUE,
		"server_ip" TEXT NOT NULL DEFAULT 'Unknown',
		"server_port" TEXT NOT NULL DEFAULT 'Unknown',
		"call_back_freq" TEXT NOT NULL DEFAULT 'Unknown',
		"call_back_jitter" TEXT NOT NULL DEFAULT 'Unknown',
		"ip" TEXT NOT NULL DEFAULT 'Unknown',
		"user" TEXT NOT NULL DEFAULT 'Unknown',
		"hostname" TEXT NOT NULL DEFAULT 'Unknown',
		"os_type" TEXT NOT NULL DEFAULT 'Unknown',
		"os_arch" TEXT NOT NULL DEFAULT 'Unknown',
		"os_build" TEXT NOT NULL DEFAULT 'Unkown',
		"cpus" TEXT NOT NULL DEFAULT 'Unknown',
		"memory" TEXT NOT NULL DEFAULT 'Unknown',
		"last_call_back" TIMESTAMP
	);
	CREATE OR REPLACE VIEW agents_status AS
	SELECT 
		"agent_id",
		"uuid",
		"server_ip",
		"server_port",
		"call_back_freq",
		"call_back_jitter",
		"ip",
		"user",
		"hostname",
		"os_type",
		"os_arch",
		"os_build",
		"cpus",
		"memory",
		"last_call_back",
		CASE 
			WHEN "last_call_back" IS NULL OR "last_call_back" < NOW() - INTERVAL '1 second' * 2 * CAST("call_back_freq" AS INTEGER) THEN 'Offline'
			ELSE 'Online'
		END AS "status"
	FROM 
		"agents";
	`
	_, err := db.Exec(AgentSQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	logger.Logf(logger.Info, "agents table initialized\n")

	CommandSQL := `
	CREATE TABLE IF NOT EXISTS "commands" (
		"command_id" SERIAL PRIMARY KEY,
		"uuid" TEXT,
		"result" TEXT,
		"command_type" TEXT,
		"command" TEXT,
		"command_uuid" TEXT,
		"output" TEXT DEFAULT 'Pending',
		FOREIGN KEY ("uuid") REFERENCES "agents" ("uuid")
	);
	`
	_, err = db.Exec(CommandSQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	logger.Logf(logger.Info, "Commands table initialized\n")

	FilesSQL := `
	CREATE TABLE IF NOT EXISTS "files" (
		"file_id" SERIAL PRIMARY KEY,
		"uuid" TEXT,
		"type" TEXT,
		"path" TEXT,
		"content" BYTEA,
		"status" TEXT DEFAULT 'Pending',
		FOREIGN KEY ("uuid") REFERENCES "agents" ("uuid")
	);
	`
	_, err = db.Exec(FilesSQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	logger.Logf(logger.Info, "Files table initialized\n")

	KeylogSQL := `
	CREATE TABLE IF NOT EXISTS "keylog" (
		"keylogID" SERIAL PRIMARY KEY,
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
	CREATE TABLE IF NOT EXISTS "payloads" (
		"payloadID" SERIAL PRIMARY KEY,
		"uuid" TEXT,
		"name" TEXT,
		"description" TEXT,
		"server_ip" TEXT,
		"server_port" TEXT,
		"callback_frequency" TEXT,
		"callback_jitter" TEXT,
		"doncat" TEXT,
		"is_deleted" INTEGER NOT NULL DEFAULT 0
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
		"note_id" SERIAL PRIMARY KEY,
		"uuid" TEXT NOT NULL,
		"note" TEXT,
		FOREIGN KEY ("uuid") REFERENCES "agents" ("uuid"),
		UNIQUE ("uuid")
	);
	`
	_, err = db.Exec(NotesSQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	logger.Logf(logger.Info, "Notes table initialized\n")

	TagsSQL := `
	CREATE TABLE IF NOT EXISTS "tags" (
		"tag_id" SERIAL PRIMARY KEY,
		"uuid" TEXT NOT NULL,
		"key" TEXT NOT NULL,
		"value" TEXT,
		FOREIGN KEY ("uuid") REFERENCES "agents" ("uuid"),
		UNIQUE ("uuid", "key")
	);
	`
	_, err = db.Exec(TagsSQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	logger.Logf(logger.Info, "tags table initialized\n")

	RedirectorsSQL := `
	CREATE TABLE IF NOT EXISTS "redirectors" (
		"redirector_id" TEXT PRIMARY KEY,
		"name" TEXT NOT NULL,
		"description" TEXT,
		"forward_ip" TEXT,
		"forward_port" TEXT,
		"listen_port" TEXT NOT NULL,
		"last_report" TIMESTAMP
	);
	CREATE OR REPLACE VIEW redirector_status AS
	SELECT 
		"redirector_id",
		"name",
		"description",
		"forward_ip",
		"forward_port",
		"listen_port",
		"last_report",
		CASE 
			WHEN "last_report" IS NULL OR "last_report" < NOW() - INTERVAL '10 minutes' THEN 'Offline'
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

	EventsSQL := `
	CREATE TABLE IF NOT EXISTS events (
		event_id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		script BYTEA,
		schedule TEXT NOT NULL,
		status TEXT NOT NULL CHECK (Status IN ('NOTRUN', 'RUNNABLE', 'RUNNING', 'COMPLETE')) DEFAULT 'UNKNOWN',
		lastrun TIMESTAMP
	);
	`
	_, err = db.Exec(EventsSQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	logger.Logf(logger.Info, "Events table initialized")

	ActionsSQL := `
	CREATE TABLE IF NOT EXISTS "actions" (
		"action_id" SERIAL PRIMARY KEY,
		"name" TEXT NOT NULL,
		"description" TEXT,
		"file" BYTEA NOT NULL
	);
	`
	_, err = db.Exec(ActionsSQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	logger.Logf(logger.Info, "Actions table initialized")

	TriggersSQL := `
	CREATE TABLE IF NOT EXISTS "triggers" (
		"id" SERIAL PRIMARY KEY,
		"event_id" INT REFERENCES events("event_id") ON DELETE CASCADE,
		"action_id" INT REFERENCES actions("action_id") ON DELETE CASCADE
	);
	`
	_, err = db.Exec(TriggersSQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	logger.Logf(logger.Info, "Triggers table initialized")
}
