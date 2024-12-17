package data

import (
	"database/sql"

	"github.com/PatronC2/Patron/types"
)

func ListEvents(db *sql.DB) ([]types.Event, error) {
	query := `SELECT "event_id", "name", "description", "script", "schedule", "last_run" FROM "events";`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []types.Event
	for rows.Next() {
		var e types.Event
		err := rows.Scan(&e.EventID, &e.Name, &e.Description, &e.Script, &e.Schedule, &e.LastRun)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}

	return events, nil
}

func GetEventByID(db *sql.DB, eventID int) (types.Event, error) {
	query := `SELECT "event_id", "name", "description", "script", "schedule", "last_run" FROM "events" WHERE "event_id" = $1;`
	var event types.Event
	err := db.QueryRow(query, eventID).Scan(
		&event.EventID,
		&event.Name,
		&event.Description,
		&event.Script,
		&event.Schedule,
		&event.LastRun,
	)
	if err != nil {
		return types.Event{}, err
	}
	return event, nil
}

func CreateEvent(db *sql.DB, event types.Event) (int, error) {
	query := `INSERT INTO "events" ("name", "description", "script", "schedule") VALUES ($1, $2, $3, $4) RETURNING "event_id";`
	var eventID int
	err := db.QueryRow(query, event.Name, event.Description, event.Script, event.Schedule).Scan(&eventID)
	if err != nil {
		return 0, err
	}
	return eventID, nil
}

func UpdateEvent(db *sql.DB, event types.Event) error {
	query := `UPDATE "events" SET "name" = $1, "description" = $2, "script" = $3, "schedule" = $4 WHERE "event_id" = $5;`
	_, err := db.Exec(query, event.Name, event.Description, event.Script, event.Schedule, event.EventID)
	return err
}

func DeleteEvent(db *sql.DB, eventID int) error {
	query := `DELETE FROM "events" WHERE "event_id" = $1;`
	_, err := db.Exec(query, eventID)
	return err
}
