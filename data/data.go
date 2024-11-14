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

	FilesSQL := `
	CREATE TABLE IF NOT EXISTS "files" (
		"FileID" SERIAL PRIMARY KEY,
		"UUID" TEXT,
		"Type" TEXT,
		"Path" TEXT,
		"Content" BYTEA,
		"Status" TEXT DEFAULT 'Pending',
		FOREIGN KEY ("UUID") REFERENCES "agents" ("UUID")
	);
	`
	_, err = db.Exec(FilesSQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	logger.Logf(logger.Info, "Files table initialized\n")

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

