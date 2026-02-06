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
			COALESCE(listen_ip, '')             AS listen_ip,
			forward_ip,
			forward_port,
			COALESCE(listen_port, '')           AS listen_port,
			COALESCE(transport_protocol::text, '') AS transport_protocol,
			COALESCE(ip_family::text, '')       AS ip_family,
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
			&r.IPFamily,
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

type RedirectorCounts struct {
	Total   int
	Online  int
	Offline int
}

func GetRedirectorCounts() (RedirectorCounts, error) {
	var c RedirectorCounts
	const q = `
		WITH per_redirector AS (
			SELECT
				r.redirector_id,
				CASE
					WHEN r.is_teamserver THEN 'Online'
					WHEN MAX(
						CASE
							WHEN l.last_report IS NOT NULL
								AND l.last_report >= NOW() - INTERVAL '10 minutes'
							THEN 1 ELSE 0
						END
					) = 1 THEN 'Online'
					ELSE 'Offline'
				END AS status
			FROM redirectors r
			LEFT JOIN redirector_listeners l
				ON r.redirector_id = l.redirector_id
			GROUP BY r.redirector_id, r.is_teamserver
		)
		SELECT
			COUNT(*)::int AS total,
			COALESCE(SUM(CASE WHEN status = 'Online'  THEN 1 ELSE 0 END), 0)::int AS online,
			COALESCE(SUM(CASE WHEN status = 'Offline' THEN 1 ELSE 0 END), 0)::int AS offline
		FROM per_redirector;
	`
	err := db.QueryRow(q).Scan(&c.Total, &c.Online, &c.Offline)
	return c, err
}

func CreateRedirector(
	RedirectorID,
	Name,
	Description,
	ForwardIP,
	ForwardPort string,
) error {
	InsertSQL := `
        INSERT INTO redirectors (redirector_id, name, description, forward_ip, forward_port)
        VALUES ($1, $2, $3, $4, $5)
    `

	_, err := db.Exec(InsertSQL,
		RedirectorID,
		Name,
		Description,
		ForwardIP,
		ForwardPort,
	)
	if err != nil {
		logger.Logf(logger.Error, "Error creating redirector with RedirectorID %s: %v", RedirectorID, err)
		return err
	}

	logger.Logf(logger.Info, "Successfully created redirector with RedirectorID %s", RedirectorID)
	return nil
}

func SetRedirectorStatus(redirectorID string, listenIPv4 string, listenIPv6 string, listenPort string, protocols []string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	updateSQL := `
        UPDATE redirectors
        SET last_report = NOW()
        WHERE redirector_id = $1;
    `
	if _, err := tx.Exec(updateSQL, redirectorID); err != nil {
		logger.Logf(logger.Error, "Error updating redirector status for %s: %v", redirectorID, err)
		return err
	}

	insertSQL := `
		INSERT INTO redirector_listeners (redirector_id, listen_ip, listen_port, protocol, ip_family, last_report)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (redirector_id, listen_ip, listen_port, protocol)
		DO UPDATE SET last_report = EXCLUDED.last_report;
	`
	if listenIPv4 != "" {
		for _, p := range protocols {
			tx.Exec(insertSQL, redirectorID, listenIPv4, listenPort, p, "ipv4")
		}
	}
	if listenIPv6 != "" {
		for _, p := range protocols {
			tx.Exec(insertSQL, redirectorID, listenIPv6, listenPort, p, "ipv6")
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	logger.Logf(logger.Info, "Updated redirector status + listeners for %s", redirectorID)
	return nil
}
