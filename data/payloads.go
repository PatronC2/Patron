package data

import (
	"database/sql"
	"fmt"

	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/types"
)

func CreatePayload(p types.Payload) error {
	const q = `
		INSERT INTO payloads
			(uuid, name, description, server_ip, server_port, callback_frequency, callback_jitter, concat, transport_protocol)
		VALUES
			($1,$2,$3,$4,$5,$6,$7,$8,$9);
	`

	_, err := db.Exec(q,
		p.Uuid,
		p.Name,
		p.Description,
		p.ServerIP,
		p.ServerPort,
		p.CallbackFrequency,
		p.CallbackJitter,
		p.Concat,
		p.TransportProtocol,
	)
	if err != nil {
		logger.Logf(logger.Error, "CreatePayload failed: %v", err)
		return err
	}

	logger.Logf(logger.Info, "New payload created in DB")
	return nil
}

func GetPayloads() ([]types.Payload, error) {
	const q = `
		SELECT
			payload_id,
			uuid,
			name,
			description,
			server_ip,
			server_port,
			callback_frequency,
			callback_jitter,
			concat,
			transport_protocol
		FROM payloads
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC;
	`

	rows, err := db.Query(q)
	if err != nil {
		logger.Logf(logger.Error, "Payloads query failed: %v", err)
		return nil, err
	}
	defer rows.Close()

	out := make([]types.Payload, 0, 32)
	for rows.Next() {
		var p types.Payload
		if err := rows.Scan(
			&p.PayloadID,
			&p.Uuid,
			&p.Name,
			&p.Description,
			&p.ServerIP,
			&p.ServerPort,
			&p.CallbackFrequency,
			&p.CallbackJitter,
			&p.Concat,
			&p.TransportProtocol,
		); err != nil {
			logger.Logf(logger.Error, "Payloads scan failed: %v", err)
			return nil, err
		}
		out = append(out, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func DeletePayload(payloadID int64) error {
	const q = `
		UPDATE payloads
		SET deleted_at = now()
		WHERE payload_id = $1
		  AND deleted_at IS NULL;
	`

	res, err := db.Exec(q, payloadID)
	if err != nil {
		logger.Logf(logger.Error, "DeletePayload failed: %v", err)
		return err
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}

	logger.Logf(logger.Info, "Payload %d marked deleted", payloadID)
	return nil
}

func GetPayloadConcat(payloadID int64) (string, error) {
	const q = `
		SELECT concat
		FROM payloads
		WHERE payload_id = $1
		  AND deleted_at IS NULL;
	`

	var concat sql.NullString
	err := db.QueryRow(q, payloadID).Scan(&concat)
	if err != nil {
		return "", err
	}
	if !concat.Valid {
		return "", nil
	}
	return concat.String, nil
}

func ParsePayloadID(s string) (int64, error) {
	var id int64
	_, err := fmt.Sscanf(s, "%d", &id)
	return id, err
}
