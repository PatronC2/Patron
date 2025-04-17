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
		agent_id SERIAL PRIMARY KEY,
		uuid TEXT NOT NULL UNIQUE,
		server_ip TEXT NOT NULL DEFAULT 'Unknown',
		server_port TEXT NOT NULL DEFAULT 'Unknown',
		callback_freq TEXT NOT NULL DEFAULT 'Unknown',
		callback_jitter TEXT NOT NULL DEFAULT 'Unknown',
		ip TEXT NOT NULL DEFAULT 'Unknown',
		agent_user TEXT NOT NULL DEFAULT 'Unknown',
		hostname TEXT NOT NULL DEFAULT 'Unknown',
		os_type TEXT NOT NULL DEFAULT 'Unknown',
		os_arch TEXT NOT NULL DEFAULT 'Unknown',
		os_build TEXT NOT NULL DEFAULT 'Unkown',
		cpus TEXT NOT NULL DEFAULT 'Unknown',
		memory TEXT NOT NULL DEFAULT 'Unknown',
		last_callback TIMESTAMPTZ,
		next_callback TIMESTAMPTZ
	);
	CREATE OR REPLACE VIEW agents_status AS
	SELECT 
		agent_id,
		uuid,
		server_ip,
		server_port,
		callback_freq,
		callback_jitter,
		ip,
		agent_user,
		hostname,
		os_type,
		os_arch,
		os_build,
		cpus,
		memory,
		last_callback,
		next_callback,
		CASE 
			WHEN next_callback IS NULL OR next_callback < NOW() - INTERVAL '5 seconds'
				THEN 'Offline'
			ELSE 'Online'
		END AS status
	FROM 
		agents;
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
		FOREIGN KEY ("UUID") REFERENCES "agents" ("uuid")
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
		FOREIGN KEY ("UUID") REFERENCES "agents" ("uuid")
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
		FOREIGN KEY ("UUID") REFERENCES "agents" ("uuid")
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
		FOREIGN KEY ("UUID") REFERENCES "agents" ("uuid"),
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
		FOREIGN KEY ("UUID") REFERENCES "agents" ("uuid"),
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
		"LastReport" TIMESTAMPTZ
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

	ConfigSQL := `
	CREATE TABLE IF NOT EXISTS configs (
		application TEXT PRIMARY KEY,
		log_level TEXT,
		log_file_max_size BIGINT
	);
	`
	_, err = db.Exec(ConfigSQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	insertDefaults := `
	INSERT INTO configs (application, log_level, log_file_max_size)
	VALUES 
		('api', 'info', 10485760),
		('server', 'info', 10485760)
	ON CONFLICT (application) DO NOTHING;
	`
	_, err = db.Exec(insertDefaults)
	if err != nil {
		log.Fatal(err)
	}
	logger.Logf(logger.Info, "configs table initialized\n")

}
