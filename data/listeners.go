package data

import (
	"log"

	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/types"
)

func GetListeners() ([]types.Listener, error) {
	var listeners []types.Listener

	rows, err := db.Query(`
		SELECT listener_id, name, description, listen_ip, listen_port, transport_protocol
		FROM listeners
	`)
	if err != nil {
		log.Println("Error querying listeners:", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var l types.Listener
		err := rows.Scan(
			&l.ListenerID,
			&l.Name,
			&l.Description,
			&l.ListenIP,
			&l.ListenPort,
			&l.TransportProtocol,
		)
		if err != nil {
			log.Println("Error scanning listener:", err)
			return nil, err
		}
		listeners = append(listeners, l)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error after iterating listener rows:", err)
		return nil, err
	}

	logger.Logf(logger.Info, "Retrieved listeners: %+v", listeners)
	return listeners, nil
}

func GetListenerByID(id int) (*types.Listener, error) {
	var l types.Listener

	row := db.QueryRow(`
		SELECT listener_id, name, description, listen_ip, listen_port, transport_protocol
		FROM listeners
		WHERE listener_id = $1
	`, id)

	err := row.Scan(
		&l.ListenerID,
		&l.Name,
		&l.Description,
		&l.ListenIP,
		&l.ListenPort,
		&l.TransportProtocol,
	)
	if err != nil {
		logger.Logf(logger.Error, "Error getting listener %d: %v", id, err)
		return nil, err
	}

	return &l, nil
}

func CreateListener(name, description, listenIP string, listenPort int, protocol string) error {
	_, err := db.Exec(`
		INSERT INTO listeners (name, description, listen_ip, listen_port, transport_protocol)
		VALUES ($1, $2, $3, $4, $5)
	`, name, description, listenIP, listenPort, protocol)

	if err != nil {
		logger.Logf(logger.Error, "Error creating listener: %v", err)
		return err
	}

	logger.Logf(logger.Info, "Created listener: %s:%d [%s]", listenIP, listenPort, protocol)
	return nil
}

func UpdateListener(id int, name, description, listenIP string, listenPort int, protocol string) error {
	_, err := db.Exec(`
		UPDATE listeners
		SET name = $1,
		    description = $2,
		    listen_ip = $3,
		    listen_port = $4,
		    transport_protocol = $5
		WHERE listener_id = $6
	`, name, description, listenIP, listenPort, protocol, id)

	if err != nil {
		logger.Logf(logger.Error, "Error updating listener %d: %v", id, err)
		return err
	}

	logger.Logf(logger.Info, "Updated listener %d", id)
	return nil
}
