package data

import (
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/types"
	_ "github.com/lib/pq"
)

func GetRedirectors() ([]types.Redirector, error) {
	FetchSQL := `
        SELECT
            redirector_id,
            name,
            description,
            listen_ip,
            forward_ip,
            forward_port,
            COALESCE(listen_port, '') AS listen_port,
            COALESCE(transport_protocol::text, '') AS transport_protocol,
            status
        FROM redirector_status
    `
	rows, err := db.Query(FetchSQL)
	if err != nil {
		logger.Logf(logger.Error, "Error querying redirectors: %v", err)
		return nil, err
	}
	defer rows.Close()

	var redirectors []types.Redirector

	for rows.Next() {
		var r types.Redirector
		if err := rows.Scan(
			&r.RedirectorID,
			&r.Name,
			&r.Description,
			&r.ListenIP,
			&r.ForwardIP,
			&r.ForwardPort,
			&r.ListenPort,
			&r.TransportProtocol,
			&r.Status,
		); err != nil {
			logger.Logf(logger.Error, "Error scanning row: %v", err)
			return nil, err
		}
		redirectors = append(redirectors, r)
	}

	if err := rows.Err(); err != nil {
		logger.Logf(logger.Error, "Error iterating over rows: %v", err)
		return nil, err
	}

	logger.Logf(logger.Info, "Current redirectors: %+v", redirectors)
	return redirectors, nil
}

func CreateRedirector(
	RedirectorID,
	Name,
	Description,
	ForwardIP,
	ForwardPort,
	ListenIP string,
) error {
	InsertSQL := `
        INSERT INTO redirectors (redirector_id, name, description, forward_ip, forward_port, listen_ip)
        VALUES ($1, $2, $3, $4, $5, $6)
    `

	_, err := db.Exec(InsertSQL,
		RedirectorID,
		Name,
		Description,
		ForwardIP,
		ForwardPort,
		ListenIP,
	)
	if err != nil {
		logger.Logf(logger.Error, "Error creating redirector with RedirectorID %s: %v", RedirectorID, err)
		return err
	}

	logger.Logf(logger.Info, "Successfully created redirector with RedirectorID %s", RedirectorID)
	return nil
}

func SetRedirectorStatus(redirectorID, listenPort string, protocols []string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1) Update last_report for status view
	updateSQL := `
        UPDATE redirectors
        SET last_report = NOW()
        WHERE redirector_id = $1;
    `
	if _, err := tx.Exec(updateSQL, redirectorID); err != nil {
		logger.Logf(logger.Error, "Error updating redirector status for %s: %v", redirectorID, err)
		return err
	}

	// 2) Clear existing listeners for this redirector
	if _, err := tx.Exec(`DELETE FROM redirector_listeners WHERE redirector_id = $1`, redirectorID); err != nil {
		logger.Logf(logger.Error, "Error clearing listeners for %s: %v", redirectorID, err)
		return err
	}

	// 3) Insert one row per protocol
	insertSQL := `
        INSERT INTO redirector_listeners (redirector_id, listen_port, protocol)
        VALUES ($1, $2, $3)
    `
	for _, p := range protocols {
		if _, err := tx.Exec(insertSQL, redirectorID, listenPort, p); err != nil {
			logger.Logf(logger.Error, "Error inserting listener for %s (%s/%s): %v",
				redirectorID, listenPort, p, err)
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	logger.Logf(logger.Info, "Updated redirector status + listeners for %s", redirectorID)
	return nil
}
