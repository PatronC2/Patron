package data

import (
	"database/sql"

	_ "github.com/lib/pq"
)

func GetLogLevel(appName string) (string, error) {
	var level string
	query := `
		SELECT log_level
		FROM configs
		WHERE application = $1
	`
	err := db.QueryRow(query, appName).Scan(&level)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return level, err
}

func SetLogLevel(appName string, level string) error {
	query := `
        INSERT INTO configs (application, log_level)
        VALUES ($1, $2)
        ON CONFLICT (application)
        DO UPDATE SET log_level = EXCLUDED.log_level;
    `
	_, err := db.Exec(query, appName, level)
	return err
}

func GetLogFileMaxSize(app string) (int64, error) {
	var size int64
	query := `
		SELECT log_file_max_size
		FROM configs
		WHERE application = $1
	`
	err := db.QueryRow(query, app).Scan(&size)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return size, err
}

func SetLogFileMaxSize(appName string, size int64) error {
	query := `
        INSERT INTO configs (application, log_file_max_size)
        VALUES ($1, $2)
        ON CONFLICT (application)
        DO UPDATE SET log_file_max_size = EXCLUDED.log_file_max_size;
    `
	_, err := db.Exec(query, appName, size)
	return err
}
