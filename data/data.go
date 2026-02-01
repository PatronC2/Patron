package data

import (
	"database/sql"
	"fmt"
	"os"
	"time"

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

	logger.Logf(logger.Info, "Got environment variables host=%s, port=%s, user=%s, dbname=%s (password not shown)", host, port, user, dbname)

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	for {
		db, err = sql.Open("postgres", psqlInfo)
		if err != nil {
			logger.Logf(logger.Error, "Failed to connect to the database: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		err = db.Ping()
		if err != nil {
			logger.Logf(logger.Error, "Failed to ping the database: %v", err)
			db.Close()
			time.Sleep(5 * time.Second)
			continue
		}
		logger.Logf(logger.Info, "Postgres DB connected")
		break
	}
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
		next_callback TIMESTAMPTZ,
		transport_protocol TEXT
	);
	CREATE OR REPLACE VIEW agents_status AS
	SELECT
	a.*,
	CASE
		WHEN a.next_callback IS NULL THEN 'Offline'
		WHEN a.next_callback < NOW() - (
		make_interval(secs =>
			5
			+ COALESCE(NULLIF(a.callback_freq,'Unknown'),'0')::int
			* (COALESCE(NULLIF(a.callback_jitter,'Unknown'),'0')::numeric / 100.0)
		)
		) THEN 'Offline'
		ELSE 'Online'
	END AS status
	FROM agents a;
	`
	_, err := db.Exec(AgentSQL)
	if err != nil {
		logger.Logf(logger.Error, "Failed to create agents table, %v", err)
	}
	logger.Logf(logger.Info, "agents table initialized")

	CommandSQL := `
	CREATE TABLE IF NOT EXISTS commands (
		command_id     BIGSERIAL PRIMARY KEY,
		uuid           TEXT NOT NULL REFERENCES agents(uuid) ON DELETE CASCADE,

		command_uuid   TEXT NOT NULL UNIQUE,
		command_type   TEXT NOT NULL,
		command        TEXT NOT NULL,

		result         TEXT NOT NULL DEFAULT '0',
		output         TEXT NOT NULL DEFAULT 'Pending',

		created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
		updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
	);

	CREATE INDEX IF NOT EXISTS idx_commands_uuid_command_id
		ON commands (uuid, command_id ASC);

	CREATE INDEX IF NOT EXISTS idx_commands_uuid_result_command_id
		ON commands (uuid, result, command_id ASC);
	`
	_, err = db.Exec(CommandSQL)
	if err != nil {
		logger.Logf(logger.Error, "Failed to create Commands table: %v", err)
	}
	logger.Logf(logger.Info, "Commands table initialized")

	FilesSQL := `
	CREATE TABLE IF NOT EXISTS "files" (
		file_id          SERIAL PRIMARY KEY,
		uuid             TEXT NOT NULL REFERENCES agents(uuid),
		type             TEXT NOT NULL,          -- "Upload" or "Download"
		path             TEXT NOT NULL,          -- agent-side path (what agent reads/writes)
		status           TEXT NOT NULL DEFAULT 'Pending'
	);
	`
	_, err = db.Exec(FilesSQL)
	if err != nil {
		logger.Logf(logger.Error, "Failed to create files table: %v", err)
	}
	logger.Logf(logger.Info, "Files table initialized")

	KeylogSQL := `
	CREATE TABLE IF NOT EXISTS keylogs (
		keylog_id		SERIAL PRIMARY KEY,
		uuid			TEXT REFERENCES agents(uuid) ON DELETE CASCADE,
		created_at		TIMESTAMPTZ NOT NULL DEFAULT now(),
		contents		TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_logs_uuid_created_at
		ON keylogs (uuid, created_at DESC);
	`
	_, err = db.Exec(KeylogSQL)
	if err != nil {
		logger.Logf(logger.Error, "Failed to create Keylog table: %v", err)
	}
	logger.Logf(logger.Info, "Keylog table initialized")

	PayloadSQL := `
	CREATE TABLE IF NOT EXISTS payloads (
		payload_id          BIGSERIAL PRIMARY KEY,
		uuid                TEXT,
		name                TEXT NOT NULL,
		description         TEXT,

		server_ip           TEXT NOT NULL,
		server_port         INTEGER NOT NULL,
		callback_frequency  INTEGER NOT NULL,
		callback_jitter     INTEGER NOT NULL,

		concat              TEXT,
		transport_protocol  TEXT,

		deleted_at          TIMESTAMPTZ,
		created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
	);

	CREATE INDEX IF NOT EXISTS idx_payloads_active_id
		ON payloads (payload_id)
		WHERE deleted_at IS NULL;
	`
	_, err = db.Exec(PayloadSQL)
	if err != nil {
		logger.Logf(logger.Error, "Failed to create Payloads table: %v", err)
	}
	logger.Logf(logger.Info, "Payloads table initialized")

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
		logger.Logf(logger.Error, "Failed to create users table: %v", err)
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
		logger.Logf(logger.Error, "Failed to create notes table: %v", err)
	}
	logger.Logf(logger.Info, "Notes table initialized")

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
		logger.Logf(logger.Error, "Failed to create tags table: %v", err)
	}
	logger.Logf(logger.Info, "tags table initialized")

	RedirectorsSQL := `
	CREATE TABLE IF NOT EXISTS redirectors (
		redirector_id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		forward_ip TEXT,
		forward_port TEXT,
		is_teamserver BOOLEAN NOT NULL DEFAULT FALSE,
		last_report TIMESTAMPTZ
	);

	DO $$
	BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'listener_protocol') THEN
			CREATE TYPE listener_protocol AS ENUM ('tcp', 'quic', 'https');
		END IF;

		IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'ip_family') THEN
			CREATE TYPE ip_family AS ENUM ('ipv4', 'ipv6');
		END IF;
	END$$;

	CREATE TABLE IF NOT EXISTS redirector_listeners (
		id BIGSERIAL PRIMARY KEY,
		redirector_id TEXT NOT NULL REFERENCES redirectors(redirector_id) ON DELETE CASCADE,
		listen_ip TEXT NOT NULL,
		listen_port TEXT NOT NULL,
		protocol listener_protocol NOT NULL,
		ip_family ip_family NOT NULL,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		last_report TIMESTAMPTZ,
		UNIQUE (redirector_id, listen_ip, listen_port, protocol)
	);

	CREATE OR REPLACE VIEW redirector_status AS
	SELECT
		r.redirector_id,
		r.name,
		r.description,
		r.forward_ip,
		r.forward_port,

		l.listen_ip,
		l.listen_port,
		l.protocol AS transport_protocol,
		l.ip_family,
		l.last_report AS listener_last_report,

		r.last_report AS redirector_last_report,
		r.is_teamserver,

		CASE
			WHEN r.is_teamserver = TRUE THEN 'Online'
			WHEN l.last_report IS NULL 
				OR l.last_report < NOW() - INTERVAL '10 minutes'
			THEN 'Offline'
			ELSE 'Online'
		END AS status

	FROM redirectors r
	LEFT JOIN redirector_listeners l
		ON r.redirector_id = l.redirector_id;
	`
	_, err = db.Exec(RedirectorsSQL)
	if err != nil {
		logger.Logf(logger.Error, "Failed to create Redirectors table: %v", err)
	}
	logger.Logf(logger.Info, "Redirectors table initialized")

	teamserverID := "11111111-1111-1111-1111-111111111111"
	tcp_listener_ip := os.Getenv("REACT_APP_NGINX_IP")
	tcp_listener_port := os.Getenv("TCP_LISTENER_PORT")
	quic_listener_port := os.Getenv("QUIC_LISTENER_PORT")

	upsertRedirectorSQL := `
		INSERT INTO redirectors (
			redirector_id,
			name,
			description,
			forward_ip,
			forward_port,
			last_report,
			is_teamserver
		)
		VALUES ($1, $2, $3, $4, $5, NOW(), TRUE)
		ON CONFLICT (redirector_id)
		DO UPDATE SET
			name         = EXCLUDED.name,
			description  = EXCLUDED.description,
			forward_ip   = EXCLUDED.forward_ip,
			forward_port = EXCLUDED.forward_port,
			last_report  = NOW(),
			is_teamserver = TRUE;
	`
	_, err = db.Exec(
		upsertRedirectorSQL,
		teamserverID,
		"teamserver",
		"default listener",
		"127.0.0.1",
		tcp_listener_port,
	)
	if err != nil {
		logger.Logf(logger.Error, "Failed to initialize teamserver listener: %v", err)
	}
	logger.Logf(logger.Info, "Teamserver initialized")

	insertTeamserverSQL := `
		INSERT INTO redirector_listeners (
			redirector_id,
			listen_ip,
			listen_port,
			protocol,
			ip_family,
			last_report
		)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (redirector_id, listen_ip, listen_port, protocol)
		DO UPDATE SET last_report = EXCLUDED.last_report;
	`

	if tcp_listener_port != "" {
		if _, err := db.Exec(
			insertTeamserverSQL,
			teamserverID,
			tcp_listener_ip,
			tcp_listener_port,
			"tcp",
			"ipv4",
		); err != nil {
			logger.Logf(logger.Error, "Failed to insert teamserver tcp listener: %v", err)
		} else {
			logger.Logf(logger.Info, "Added teamserver tcp listener")
		}
	}

	if quic_listener_port != "" {
		if _, err := db.Exec(
			insertTeamserverSQL,
			teamserverID,
			tcp_listener_ip,
			quic_listener_port,
			"quic",
			"ipv4",
		); err != nil {
			logger.Logf(logger.Error, "Failed to insert teamserver quic listener: %v", err)
		} else {
			logger.Logf(logger.Info, "Added teamserver quic listener")
		}
	}

	ConfigSQL := `
	CREATE TABLE IF NOT EXISTS configs (
		application TEXT PRIMARY KEY,
		log_level TEXT,
		log_file_max_size BIGINT
	);
	`
	_, err = db.Exec(ConfigSQL)
	if err != nil {
		logger.Logf(logger.Error, "Failed to create configs table: %v", err)
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
		logger.Logf(logger.Error, "Failed to insert defaults: %v", err)
	}
	logger.Logf(logger.Info, "configs table initialized\n")

}
