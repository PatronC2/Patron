package main

import (
    "log"
	"fmt"
	"os"
	"time"

	"github.com/PatronC2/Patron/lib/logger"	
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
    "github.com/jmoiron/sqlx"
)

var db *sqlx.DB

func goDotEnvVariable(key string) string {

	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}

func OpenDatabase(){ 
	var err error
	var port int
	host := goDotEnvVariable("DB_HOST")
	fmt.Sscan(goDotEnvVariable("DB_PORT"), &port)
	user := goDotEnvVariable("DB_USER")
	password := goDotEnvVariable("DB_PASS")
	dbname := goDotEnvVariable("DB_NAME")

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
    "password=%s dbname=%s sslmode=disable",
    host, port, user, password, dbname)
	for {

		db, err = sqlx.Open("postgres", psqlInfo)
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
	default_user_name := "patron"
	default_user_pass := goDotEnvVariable("ADMIN_AUTH_PASS")

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

    passwordHash, err := bcrypt.GenerateFromPassword([]byte(default_user_pass), bcrypt.DefaultCost)
    if err != nil {
        log.Fatal(err.Error())
    }

    CreateAdminUserSQL := `
	INSERT INTO users (username, password_hash, role)
	VALUES ($1, $2, 'admin')
	ON CONFLICT (username) DO NOTHING;
	`

    _, err = db.Exec(CreateAdminUserSQL, default_user_name, string(passwordHash))
    if err != nil {
        log.Fatal(err.Error())
    }
    log.Println("Admin user created")
}
