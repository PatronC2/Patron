package api

import (
    "log"
	"fmt"
	"time"

	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/data"
	_ "github.com/lib/pq"
    "github.com/jmoiron/sqlx"
)

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
    log.Println("Users table initialized")

}
