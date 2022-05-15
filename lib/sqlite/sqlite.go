// Database functions.
package sqlite

import (
	"database/sql"
	"os"

	"github.com/PatronC2/Patron/lib/logger"
)

const (
	DatabaseFilename string = "patron.sqlite3"

	ERR_DATABASE_INVALID int = 20
	ERR_STATEMENT        int = 21
	ERR_QUERY            int = 22
	ERR_SCAN             int = 23
)

/*
	Return a handle to the application database.

	Currently hardcoded to open utils.sqlite.DatabaseFilepath.
*/
func GetDatabaseHandle() *sql.DB {
	db, err := sql.Open("sqlite3", DatabaseFilename)
	if err != nil {
		logger.LogError(err)
		os.Exit(ERR_DATABASE_INVALID)
	}

	if db == nil {
		logger.Log(logger.Error, "db == nil, this should never happen")
		os.Exit(logger.ERR_UNKNOWN)
	} else {
		logger.Log(logger.Done, "Opened database file")
	}

	return db
}
